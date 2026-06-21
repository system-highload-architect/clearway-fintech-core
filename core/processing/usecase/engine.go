package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"clearway-fintech-core/core/domain"
	"clearway-fintech-core/internal/pkg/fixedpoint"
	"clearway-fintech-core/internal/pkg/fsm"
	"clearway-fintech-core/internal/pkg/idempotency"
	"clearway-fintech-core/pb/gen" // Сгенерированные Protobuf-контракты
)

var (
	ErrTxNotFound       = errors.New("🔒 [PROCESSING CORE]: Транзакция не найдена в ОЗУ процессора")
	ErrIdempotencyClash = errors.New("🔒 [PROCESSING CORE]: Конфликт идемпотентности. Запрос уже находится в обработке (In-Flight)")
)

// LedgerClientInterface изолирует зависимость от gRPC по закону DIP (SOLID)
type LedgerClientInterface interface {
	CommitDoubleEntryTransaction(ctx context.Context, in *gen.LedgerEntryRequest) (*gen.LedgerEntryResponse, error)
}

// ProcessingEngine осуществляет Control Plane оркестрацию транзакционного графа
// FIXED: Completely refactored to align with strict SOLID patterns and unified state tracking
type ProcessingEngine struct {
	mu              sync.RWMutex
	transactions    map[string]*domain.Transaction
	stateMachine    *fsm.FiniteStateMachine
	idempotencyPool *idempotency.IdempotencyPoolManager
	ledgerClient    LedgerClientInterface
}

// NewProcessingEngine собирает транзакционный процессор и настраивает инварианты FSM-переходов
func NewProcessingEngine(ledgerClient LedgerClientInterface) *ProcessingEngine {
	// Жестко запечатываем ориентированный граф легальных b2b-переходов стейт-машины
	allowedTransitions := map[string][]string{
		domain.StateNew:           {domain.StateFraudChecking, domain.StateFailed},
		domain.StateFraudChecking: {domain.StateAcquiringHold, domain.StateFailed},
		domain.StateAcquiringHold: {domain.StateHeld, domain.StateFailed},
		domain.StateHeld:          {domain.StateCaptured, domain.StateReversed, domain.StateFailed},
		domain.StateCaptured:      {}, // Конечный стейт успешного списания
		domain.StateReversed:      {}, // Конечный стейт полной компенсации
		domain.StateFailed:        {}, // Конечный стейт отказа
	}

	// Регистрируем поддерживаемые категории TTL для пула кэшей (High Cohesion)
	ttlConfig := map[string]time.Duration{
		"TX_INSTANT":  1 * time.Minute,  // Краткосрочные переводы
		"TX_STANDARD": 30 * time.Minute, // Оплата счетов, СБП QR, 3D-Secure
	}

	poolManager := idempotency.NewIdempotencyPoolManager(50000, ttlConfig)

	// Локальный Janitor-демон автоматической зачистки просроченных сессий ОЗУ
	stopChan := make(chan struct{})
	poolManager.StartGlobalJanitorDaemon(stopChan, 1*time.Minute)

	return &ProcessingEngine{
		transactions:    make(map[string]*domain.Transaction),
		stateMachine:    fsm.NewFiniteStateMachine(allowedTransitions),
		idempotencyPool: poolManager,
		ledgerClient:    ledgerClient,
	}
}

// ExecuteHoldInitiation оркестрирует ФАЗУ АВТОР ИЗАЦИИ (Блокировка/Заморозка средств)
func (e *ProcessingEngine) ExecuteHoldInitiation(
	ctx context.Context,
	merchantID, paymentToken, idempotencyKey, categoryTag string,
	amount fixedpoint.Money,
) (string, string, error) {

	// 1. Аппаратный барьер идемпотентности: отсекаем DoS-дубликаты и атаки Double-Spending за O(1)
	cachedResult, isDuplicate, err := e.idempotencyPool.CheckOrLockTransaction(categoryTag, idempotencyKey)
	if err != nil {
		return "", "", err
	}
	if isDuplicate {
		if cachedResult.Status == "PROCESSING_IN_FLIGHT" {
			return "", "", ErrIdempotencyClash
		}
		return idempotencyKey, cachedResult.Status, nil
	}

	txID := fmt.Sprintf("tx_%d", time.Now().UnixNano())

	e.mu.Lock()
	tx := &domain.Transaction{
		ID:             txID,
		MerchantID:     merchantID,
		PaymentToken:   paymentToken,
		Amount:         amount,
		CurrentState:   domain.StateNew,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	e.transactions[txID] = tx
	e.mu.Unlock()

	// 2. Сдвиг FSM: NEW ➔ FRAUD_CHECKING
	if err := e.stateMachine.ValidateTransition(tx.CurrentState, domain.StateFraudChecking); err != nil {
		tx.CurrentState = domain.StateFailed
		return txID, tx.CurrentState, err
	}
	tx.CurrentState = domain.StateFraudChecking
	e.idempotencyPool.UpdateTransactionStep(categoryTag, idempotencyKey, domain.StateFraudChecking, 5*time.Minute)

	// 3. Сдвиг FSM: FRAUD_CHECKING ➔ ACQUIRING_HOLD (Симуляция прохода скоринг-радара)
	if err := e.stateMachine.ValidateTransition(tx.CurrentState, domain.StateAcquiringHold); err != nil {
		tx.CurrentState = domain.StateFailed
		return txID, tx.CurrentState, err
	}
	tx.CurrentState = domain.StateAcquiringHold
	e.idempotencyPool.UpdateTransactionStep(categoryTag, idempotencyKey, domain.StateAcquiringHold, 5*time.Minute)

	// 4. Сдвиг FSM: ACQUIRING_HOLD ➔ HELD (Банк-эквайер успешно заморозил деньги на балансе)
	if err := e.stateMachine.ValidateTransition(tx.CurrentState, domain.StateHeld); err != nil {
		tx.CurrentState = domain.StateFailed
		return txID, tx.CurrentState, err
	}
	tx.CurrentState = domain.StateHeld
	tx.UpdatedAt = time.Now()

	// Запечатываем финальный стейт фазы авторизации в Sliding Window кэш на 30 минут
	e.idempotencyPool.UpdateTransactionStep(categoryTag, idempotencyKey, domain.StateHeld, 30*time.Minute)

	return txID, domain.StateHeld, nil
}

// ExecuteCaptureConfirmation оркестрирует ФАЗУ РАСЧЕТА (Списание ранее замороженных копеек)
func (e *ProcessingEngine) ExecuteCaptureConfirmation(ctx context.Context, txID string) (string, error) {
	e.mu.Lock()
	tx, exists := e.transactions[txID]
	e.mu.Unlock()

	if !exists {
		return "", ErrTxNotFound
	}

	// Проверяем инвариант графа переходов FSM: имеет ли право транзакция списаться из текущего HELD?
	if err := e.stateMachine.ValidateTransition(tx.CurrentState, domain.StateCaptured); err != nil {
		return tx.CurrentState, err
	}

	// МЕЖСЕРВЕРНАЯ ИНТЕГРАЦИЯ ДАННЫХ (Data Plane): Вызываем Ledger Бухгалтерию по gRPC
	// Списываем средства со счета плательщика Давида и зачисляем на торговый счет мерчанта
	grpcResponse, err := e.ledgerClient.CommitDoubleEntryTransaction(ctx, &gen.LedgerEntryRequest{
		DebitWalletId:  "wallet_david_buyer",
		CreditWalletId: "wallet_merchant_shop",
		AmountUnits:    tx.Amount.Units,
		TxId:           tx.ID,
		Description:    fmt.Sprintf("Окончательное b2b-списание по транзакции %s мерчанта %s", tx.ID, tx.MerchantID),
	})

	if err != nil || !grpcResponse.IsCommitted {
		tx.CurrentState = domain.StateFailed
		tx.UpdatedAt = time.Now()
		return domain.StateFailed, fmt.Errorf("🔒 [LEDGER REJECTION]: Бухгалтерская книга отклонила клиринг: %s", grpcResponse.GetErrorMessage())
	}

	// Транзакция завершена с абсолютным успехом
	tx.CurrentState = domain.StateCaptured
	tx.UpdatedAt = time.Now()

	return domain.StateCaptured, nil
}

// ExecuteReversalCancellation оркестрирует ФАЗУ КОМПЕНСАЦИИ (Мгновенная отмена холдов и возвраты)
func (e *ProcessingEngine) ExecuteReversalCancellation(ctx context.Context, txID string) (string, error) {
	e.mu.Lock()
	tx, exists := e.transactions[txID]
	e.mu.Unlock()

	if !exists {
		return "", ErrTxNotFound
	}

	// Проверяем по графу FSM легальность отката
	if err := e.stateMachine.ValidateTransition(tx.CurrentState, domain.StateReversed); err != nil {
		return tx.CurrentState, err
	}

	// Разблокируем средства (в промышленном Ledger тут идет обратная компенсирующая двойная запись)
	tx.CurrentState = domain.StateReversed
	tx.UpdatedAt = time.Now()

	return domain.StateReversed, nil
}
