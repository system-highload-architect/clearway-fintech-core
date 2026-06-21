package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Инжектируем UseCase и хэндлеры из нашего Модульного Монолита
	gatewayHttp "clearway-fintech-core/core/gateway/delivery/http"
	gatewayUc "clearway-fintech-core/core/gateway/usecase"
	ledgerGrpc "clearway-fintech-core/core/ledger/delivery/grpc"
	ledgerUc "clearway-fintech-core/core/ledger/usecase"
	processingGrpc "clearway-fintech-core/core/processing/delivery/grpc"
	processingUc "clearway-fintech-core/core/processing/usecase"
	"clearway-fintech-core/internal/chassis"
	"clearway-fintech-core/internal/pkg/fixedpoint"
)

var cfg chassis.BaseConfig

func main() {
	if err := chassis.LoadConfigAbstract("services/cloud-routing-proxy/config.yaml", &cfg); err != nil {
		fmt.Printf("⚠️ [CONFIG WARN]: Не удалось загрузить config.yaml, фолбэк на дефолты: %v\n", err)
		cfg.ServerPort = ":8080" // Фолбэк-дефолт
	}

	fmt.Printf("📡 [CHASSIS]: Конфигурация успешно загружена. Окружение: %s | Ин ingress-порт: %s\n", cfg.Environment, cfg.ServerPort)

	fmt.Println("🚀 [API GATEWAY]: Запуск Ingress Эшелона Локального Монолита...")

	// 1. Инициализируем слои и зависимости "на берегу" (DI-Контейнер ОЗУ)
	ledgerCore := ledgerUc.NewLedgerUseCase()
	ledgerHandler := ledgerGrpc.NewLedgerGrpcHandler(ledgerCore)

	// Процессинг принимает Ledger напрямую через интерфейс (0% сетевых задержек в монолите)
	processingCore := processingUc.NewProcessingEngine(ledgerHandler)
	_ = processingGrpc.NewProcessingGrpcHandler(processingCore)

	tokenizeCore, err := gatewayUc.NewTokenizeUseCase()
	if err != nil {
		fmt.Printf("🔒 [API GATEWAY CRITICAL ERROR]: %v\n", err)
		os.Exit(1)
	}
	gatewayHandler := gatewayHttp.NewGatewayHttpHandler(tokenizeCore)

	// 2. Настраиваем b2b-маршрутизацию HTTP ServeMux
	mux := http.NewServeMux()

	// Ручка токенизации карт для веб-формы оплаты
	mux.HandleFunc("/api/v1/gateway/tokenize", gatewayHandler.HandleTokenizeRequest)

	mux.HandleFunc("/api/v1/charge", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Нативно вызываем фазу HOLD (Авторизация) в ОЗУ процессора
		// Передаем жестко фиксированную сумму в 150.25 рублей (15025 копеек) для демонстрации
		amountMoney := fixedpoint.NewMoneyFromInt64(15025)
		idempotencyKey := fmt.Sprintf("idem_%d", time.Now().UnixNano())

		txID, _, err := processingCore.ExecuteHoldInitiation(
			r.Context(),
			"merchant_shop_id",
			"tok_pki_sample_token_178",
			idempotencyKey,
			"TX_STANDARD",
			amountMoney,
		)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"%v"}`, err)))
			return
		}

		// Если фаза HOLD прошла успешно — мгновенно и атомарно вызываем фазу CAPTURE (Списание в Ledger)
		finalState, err := processingCore.ExecuteCaptureConfirmation(r.Context(), txID)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"%v","state":"%s"}`, err, finalState)))
			return
		}

		// Возвращаем фронтенду триумфальный статус проводки копеек по Ledger-книге
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"success":true,"tx_id":"%s","final_state":"%s"}`, txID, finalState)))
	})

	// ИСПРАВЛЕНО (Уничтожение ошибок 404 и text/plain MIME Checking):
	// Направляем FileServer на корень папки "web", убираем StripPrefix,
	// чтобы пути /static/js/... и /static/css/... идеально мапились на диск!
	// FIXED: Standardized static files router to prevent relative path mapping bugs and MIME check violations
	fileServer := http.FileServer(http.Dir("web"))
	mux.Handle("/static/", fileServer)

	// Перехватчик фавикона (Favicon Guard): возвращаем 204 No Content за 0 наносекунд,
	// предотвращая лишний дисковый оверхед и ошибки 404 в консоли!
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Рендеринг главной страницы
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			http.ServeFile(w, r, "web/index.html")
			return
		}
		// Для любых других кастомных путей вызываем стандартный сервер статики
		fileServer.ServeHTTP(w, r)
	})

	// 3. Запускаем HTTP-сервер на внешнем порту :8080
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		fmt.Println("🪐 [API GATEWAY]: Контур Входа развернут на http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Крах HTTP-рантайма: %v\n", err)
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	fmt.Println("🛑 [API GATEWAY]: Сервер безопасно остановлен.")
}
