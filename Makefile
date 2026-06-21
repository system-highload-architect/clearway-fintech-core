# --- КАН ОНИЧЕСКИЙ FINTECH MAKEFILE ДЛЯ ГЕНЕРАЦИИ GRPC-КОНТРАКТОВ ---

.PHONY: generate sync

generate:
	@echo "📡 [PROTOC COMPILER]: Запуск компиляции бинарных b2b gRPC-интерфейсов..."
	@mkdir -p pb/gen
	
	# Запуск компиляции с жестким выравниванием путей под Go Workspaces
	# FIXED: Enforced strict paths injection via single execution pass
	protoc --proto_path=pb \
		--go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		pb/gateway.proto pb/processing.proto pb/wallet.proto pb/fraud.proto
	
	@echo "🪐 [PROTOC COMPILER]: gRPC-мосты успешно сгенерированы."

sync:
	@echo "🔄 [GO WORKSPACES]: Намертво синхронизируем графы импортов модулей..."
	@go work sync
	@echo "🏆 [FINISH]: Кластер полностью готов к материализации бизнес-логики."
