# 🎰 LOW-LEVEL SPECIFICATION: PROCESSING ENGINE / CONTROL PLANE

[English version below]

## 🇷🇺 РУССКАЯ ВЕРСИЯ

### 1. Реализация Архитектуры Оркестрации
Модуль `core/processing/usecase/engine.go` является координатором транзакций [2.1]. Он имплементирует интерфейсы и управляет пулом идемпотентности, стейт-машиной и gRPC-коммуникациями плоскости данных [1.1, 2.1].

### 📊 Диаграмма Вызовов и Атомарных Переходов (Processing Pipeline):
```mermaid
sequenceDiagram
    autonumber
    participant Proxy as 🌐 cloud-routing-proxy
    participant Proc as 🎰 ProcessingEngine
    participant Pool as ⏰ IdempotencyPoolManager
    participant FSM as 🔗 fsm.FiniteStateMachine
    participant Ledger as 🗄️ core/ledger (gRPC Client)

    Proxy->>Proc: Call: ExecuteHoldInitiation(categoryTag, idempotencyKey)
    Proc->>Pool: CheckOrLockTransaction(categoryTag, idempotencyKey)
    Pool-->>Proc: Status: OK (Уникальный инфлайт-запрос заблокирован в ОЗУ пула)
    
    Proc->>FSM: ValidateTransition(NEW ➔ FRAUD_CHECKING)
    Proc->>Pool: UpdateTransactionStep(FRAUD_CHECKING, TTL: 5 мин)
    
    Proc->>FSM: ValidateTransition(ACQUIRING_HOLD ➔ HELD)
    Proc->>Pool: UpdateTransactionStep(HELD, Продление Sliding TTL: 30 мин)
    Proc-->>Proxy: Hold Успешен! (Возврат стейта HELD и tx_ID)

    Proxy->>Proc: Call: ExecuteCaptureConfirmation(txID)
    Proc->>FSM: ValidateTransition(HELD ➔ CAPTURED)
    Proc->>Ledger: gRPC: CommitDoubleEntryTransaction(Debit, Credit, Amount)
    Ledger-->>Proc: Protobuf Response: IsCommitted = True
    Proc-->>Proxy: Списание подтверждено! Стейт: CAPTURED
```

---

## 🇺🇸 ENGLISH VERSION

### 1. Orchestration Core Mechanics
Coordinates transactional flows by coupling atomic FSM maps with an extensible gRPC data conduit [1.1, 2.1]. Utilizes structural interfaces under explicit dependency inversion guidelines [1.1].
