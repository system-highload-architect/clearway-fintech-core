internal/pkg/
├── fixedpoint/         # Точная финтех-математика копеек (без float64!) [1.1]
├── fsm/                # Нативная стейт-машина транзакций Finite State Machine [1.1]
├── registry/           # O(1) Таблица функций команд (без switch-case) [1.1]
├── idempotency/        # Фильтр дубликатов транзакций по Idempotency-Key [1.1]
├── circuitbreaker/     # Автомат изоляции сбоев внешних банковских API
├── backoff/            # Экспоненциальный ретраер с Jitter-шумом [1.1]
├── panichandler/       # Безопасный изолятор паник асинхронных горутин
├── ratelimit/          # CAS Lock-Free ограничитель TPS мерчантов [1.1]
├── context/            # gRPC-проброс метаданных, токенов и Correlation-ID
└── stringutils/        # Финтех-валидатор Луна и маскирование карт по PCI-DSS [1.1]


# 1. Инициализируем модуль Protobuf-контрактов
cd pb
go mod init clearway-fintech-core/pb
cd ..

# 2. Инициализируем модуль системного общего шасси и pkg-алгоритмов
cd internal
go mod init clearway-fintech-core/internal
cd ..

# 3. Инициализируем модуль инфраструктурного L7 API Gateway
cd services/cloud-routing-proxy
go mod init clearway-fintech-core/services/cloud-routing-proxy
cd ..

# 4. Инициализируем модуль бизнес-ядра Модульного Монолита
cd core
go mod init clearway-fintech-core/core
cd ..

# 5. Сносим старые привязки и инициализируем чистое b2b рабочее пространство
rm -f go.work
go work init

# 6. Подключаем все 4 изолированных go.mod модуля в контур сборки Go Work
go work use ./pb
go work use ./internal
go work use ./services/cloud-routing-proxy
go work use ./core

# 7. Намертво синхронизируем именованные графы импортов кластера
go work sync
