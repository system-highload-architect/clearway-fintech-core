# 🚀 FINTECH CLUSTER BUILDING REGULATION / РЕГЛАМЕНТ СБОРКИ ПЛАТЕЖНОГО ШЛЮЗА

[English version below]

## 🇷🇺 РУССКАЯ ВЕРСИЯ

Данный документ содержит пошаговый регламент низкоуровневой сборки изолированных модулей через Go Workspaces и локального запуска API-шлюза.

### 🛠️ Этап 1. Холодная сшивка Go-модулей (Go 1.25.0+)
Выполните следующие команды в терминале Git Bash из корня репозитория для жесткого выравнивания графа импортов:

```bash
# Инициализация модуля Protobuf-контрактов
cd pb && go mod init clearway-fintech-core/pb && go mod edit -go=1.25.0 && cd ..

# Инициализация модуля системного шасси
cd internal && go mod init clearway-fintech-core/internal && go mod edit -go=1.25.0 && cd ..

# Инициализация модуля Edge Ingress прокси-сервера
cd services/cloud-routing-proxy && go mod init clearway-fintech-core/services/cloud-routing-proxy && go mod edit -go=1.25.0 && cd ..

# Инициализация бизнес-ядра Модульного Монолита
cd core && go mod init clearway-fintech-core/core && go mod edit -go=1.25.0 && cd ..

# Сборка чистого b2b рабочего пространства
rm -f go.work && go work init
go work use ./pb ./internal ./services/cloud-routing-proxy ./core
go work sync
```

### 📡 Этап 2. Генерация gRPC-контрактов и Локальный Запуск
```bash
# Компиляция Protobuf gRPC-мостов
protoc --proto_path=pb --go_out=pb --go_opt=paths=source_relative --go-grpc_out=pb --go-grpc_opt=paths=source_relative pb/gateway.proto pb/processing.proto pb/wallet.proto pb/fraud.proto

# Локальный старт API Gateway и Модульного Монолита
go run services/cloud-routing-proxy/cmd/main.go
```

---

## 🇺🇸 ENGLISH VERSION

Operational manual for cold infrastructure initialization, Go Workspaces compilation, and gateway bootstrapping.

### 🛠️ Step 1. Go Workspaces Module Scaffolding (Go 1.25.0+)
```bash
cd pb && go mod init clearway-fintech-core/pb && go mod edit -go=1.25.0 && cd ..
cd internal && go mod init clearway-fintech-core/internal && go mod edit -go=1.25.0 && cd ..
cd services/cloud-routing-proxy && go mod init clearway-fintech-core/services/cloud-routing-proxy && go mod edit -go=1.25.0 && cd ..
cd core && go mod init clearway-fintech-core/core && go mod edit -go=1.25.0 && cd ..
rm -f go.work && go work init
go work use ./pb ./internal ./services/cloud-routing-proxy ./core
go work sync
```

### 📡 Step 2. gRPC Stubs Generation & Local Bootstrapping
```bash
protoc --proto_path=pb --go_out=pb --go_opt=paths=source_relative --go-grpc_out=pb --go-grpc_opt=paths=source_relative pb/gateway.proto pb/processing.proto pb/wallet.proto pb/fraud.proto
go run services/cloud-routing-proxy/cmd/main.go
```
