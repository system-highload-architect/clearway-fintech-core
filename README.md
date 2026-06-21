# 🪐 CLEARWAY PAY FINTECH PROCESSING CORE (OPERATOR CLASS)

Distributed high-performance transactional payment core engineered under strict PCI-DSS segregation patterns. Powered by Go 1.25.0 Workspaces, an immutable Double-Entry Bookkeeping Append-Only Ledger, a reflection-free O(1) Functional Command Dispatcher, and a Lock-Free CPU-Atomic CAS TPS rate limiter [1.1, 2.1].

Высокопроизводительное распределенное транзакционное финтех-ядро процессинга операторского класса, спроектированное по строгим стандартам изоляции контуров PCI-DSS. Платформа построена на базе Go 1.25.0 Workspaces, иммутабельной книги балансов двойной записи (Append-Only Ledger), безрефлексивного O(1) диспетчера команд и Lock-Free атомарного лимитера TPS на регистрах CPU [1.1, 2.1].

---

## 🛠️ DOCUMENTATION MAP & RUNTIME MANUAL / КАРТА СПЕЦИФИКАЦИЙ И ЗАПУСК

*   🚀 **[LAUNCH.md](LAUNCH.md)** — step-by-step cold initialization, Go Workspace compilation, gRPC stubs build, and deployment guide [2.1].
*   🚀 **[LAUNCH.md](LAUNCH.md)** — пошаговый регламент холодной инициализации, сборки Go воркспейсов, компиляции gRPC и запуска шлюза [2.1].
*   🗺️ **[docs/navigation.md](docs/navigation.md)** — unified index link board routing to detailed Software Requirement Specifications (SRS) and Low-Level Component Specifications [2.1].
*   🗺️ **[docs/navigation.md](docs/navigation.md)** — единая навигационная карта, ведущая к детальным Техническим Заданиям (SRS) и спецификациям каждого модуля [2.1].

---

## 📐 1. GLOBAL ARCHITECTURE & TOPOLOGY (DATA FLOW)

The ecosystem is segregated into three strictly isolated tiers: **Ingress Edge Plane** (stateless routing proxy), **Control Plane Core** (Finite State Machine transaction orchestrator), and **Data Plane Ledger** (immutable accounting volumes) [2.1].

Экосистема разделена на три изолированных эшелона: **Ingress Edge Plane** (сетевой прокси-регулировщик), **Control Plane Core** (стейт-машина оркестрации шагов транзакции) и **Data Plane Ledger** (иммутабельная книга балансов) [2.1].

### 📊 Comprehensive Request-Response Sequence Diagram / Подробная диаграмма вызовов

```mermaid
sequenceDiagram
    autonumber
    actor Customer as 💳 UI Client Form
    participant Proxy as 🌐 services/cloud-routing-proxy (:8080)
    participant Gateway as 🔒 core/gateway (PCI-DSS)
    participant Processing as 🎰 core/processing (FSM Engine)
    participant Fraud as 🛡️ core/fraud (Scoring Radar)
    participant Ledger as 🗄️ core/ledger (Data Plane Book)

    %% PHASE 1: INGRESS TOKENIZATION
    Note over Customer, Gateway: Phase 1: Ingress Edge Card Tokenization (PCI-DSS Perimeter)
    Customer->>Proxy: HTTP POST /api/v1/gateway/tokenize {"card_number": "4111...", "amount": "150.25"}
    Proxy->>Proxy: Extract L7 routing metrics via Consistent Hash Ring configuration
    Proxy->>Gateway: Forward raw JSON payload inside the internal Docker bridge network
    Gateway->>Gateway: Execute stringutils.VerifyLuhnAlgorithm() check over raw PAN (0ns, reflection-free)
    Gateway->>Gateway: Encrypt raw parameters via crypto.CryptoVault (AES-256-GCM blocks)
    Gateway-->>Proxy: Return stateless transaction surrogate string payload token (tok_pki_*)
    Proxy-->>Customer: HTTP 200 OK Token Response (Sensitive PAN data wiped from memory heaps)

    %% PHASE 2: PROCESSING & CONTROL FLOW
    Note over Customer, Processing: Phase 2: Transaction Orchestration & State Machine Steps
    Customer->>Proxy: HTTP GET /api/v1/charge (Initiates secure billing transaction lifecycle)
    Proxy->>Processing: gRPC: ProcessHold(HoldRequest{payment_token, amount_units, idempotency_key})
    
    Processing->>Processing: idempotency.IdempotencyPoolManager.CheckOrLockTransaction() via tag maps [O(1)]
    alt Idempotency Key Duplicate Detected
        Processing-->>Proxy: Return previously cached transaction result state straight from RAM
    end
    
    Processing->>Processing: fsm.FiniteStateMachine shifts state: NEW ➔ FRAUD_CHECKING
    Processing->>Fraud: gRPC: CheckFraudScore(client_ip, device_fingerprint, amount_units)
    Fraud->>Fraud: Compute pattern risk variables ( sub-5ms compliance SLA threshold checking)
    Fraud-->>Processing: FraudCheckResponse { is_fraudulent: false, risk_score: 0.03 }
    
    Processing->>Processing: fsm.FiniteStateMachine shifts state: FRAUD_CHECKING ➔ ACQUIRING_HOLD
    Note over Processing: Simulating bank acquiring network response latency
    Processing->>Processing: fsm.FiniteStateMachine shifts state: ACQUIRING_HOLD ➔ HELD
    Processing->>Processing: UpdateTransactionStep() -> Reset Sliding Window TTL inside memory pool to 30 Min

    %% PHASE 3:Persistency and Clearing Double-Entry Bookkeeping
    Note over Processing, Ledger: Phase 3: Financial Settlement & Double-Entry Accounting
    Processing->>Ledger: gRPC: CommitDoubleEntryTransaction(debit_wallet, credit_wallet, amount_units)
    Ledger->>Ledger: Extract target wallets, evaluate strict fixedpoint.Money.Sub() overdraft barriers
    Ledger->>Ledger: Append immutable LedgerEntry transaction record data slice into repository map
    Note over Ledger: DB Operations are strictly INSERT ONLY. UPDATE/DELETE queries are natively banned.
    Ledger-->>Processing: LedgerEntryResponse { is_committed: true, new_balances_states }
    
    Processing->>Processing: fsm.FiniteStateMachine shifts state: HELD ➔ CAPTURED (Final state reached)
    Processing-->>Proxy: gRPC: HoldResponse { is_success: true, current_state: "CAPTURED" }
    Proxy-->>Customer: HTTP 200 OK JSON Processing Success Visual Response Array
```

---

## 🎰 2. STRUCTURAL DOMAIN DECOUPLING / РАЗДЕЛЕНИЕ КОНТУРОВ СИСТЕМЫ

1.  **`services/` (Network Infrastructure Edge)**: ontains stateless reverse-proxies, configuration chassis routers, and L7 switches [2.1]. Free from any financial context or business-logic variables [1.1].
2.  **`core/` (Transactional Monolith Core)**: encapsulates high-availability banking sub-domains [2.1]. Bound strictly to interface abstraction models, allowing immediate zero-code-change microservice partitioning onto separate machines or Kubernetes nodes [1.1, 2.1].
3.  **`internal/pkg/` (Blazing Fast Shared Frameworks)**: reflection-free, zero-allocation algorithms общего назначения ($O(1)$ maps registry dispatchers, CPU-atomic CAS token buckets, memory-sharded sliding-window caches) [1.1].
