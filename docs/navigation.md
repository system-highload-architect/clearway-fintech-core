# 🗺️ ENTERPRISE ARCHITECTURE NAVIGATION INDEX / КАРТА ТЗ И СПЕЦИФИКАЦИЙ КЛАССТЕРА

[English version below]

## 🇷🇺 РУССКАЯ ВЕРСИЯ

Данный навигационный манифест структурирует полную техническую документацию WebRTC-Mesh платформы Clearway Fintech Core. Проект разбит на требования (SRS) и детальные архитектурные спецификации реализации (Specification) [2.1].

### 📋 1. ЭШЕЛОН ТЕХНИЧЕСКИХ ЗАДАНИЙ (DOCS/SRS/ - 7 ФАЙЛОВ)
*   🌐 [docs/srs/cloud-routing-proxy-srs.md](srs/cloud-routing-proxy-srs.md) — ТЗ: L7 Ingress Gateway балансировщик трафика мерчантов.
*   🔒 [docs/srs/gateway-srs.md](srs/gateway-srs.md) — ТЗ: PCI-DSS Ingress токенизатор и крипто-периметр карт.
*   🎰 [docs/srs/processing-srs.md](srs/processing-srs.md) — ТЗ: Control Plane оркестратор и стейт-машина FSM.
*   🗄️ [docs/srs/ledger-srs.md](srs/ledger-srs.md) — ТЗ: Data Plane Append-Only распределенная бухгалтерская книга.
*   🛡️ [docs/srs/fraud-srs.md](srs/fraud-srs.md) — ТЗ: Security Plane радар скоринга фрод-атак.
*   🧮 [docs/srs/core-pkg-srs.md](srs/core-pkg-srs.md) — ТЗ: требования к производительности и сложности алгоритмов pkg.
*   📦 [docs/srs/infrastructure-srs.md](srs/infrastructure-srs.md) — ТЗ: критерии контейнеризации Go Workspaces кластера.

### 📐 2. ЭШЕЛОН АРХИТЕКТУРНЫХ ОПИСАНИЙ (DOCS/SPECIFICATION/ - 8 ФАЙЛОВ)
*   🌐 [docs/specification/cloud-routing-proxy-spec.md](specification/cloud-routing-proxy-spec.md) — спецификация: Архитектура Ingress-маршрутизации шасси.
*   🔒 [docs/specification/gateway-spec.md](specification/gateway-spec.md) — спецификация: реализация AES-256-GCM токенизации карт.
*   🎰 [docs/specification/processing-spec.md](specification/processing-spec.md) — спецификация: логика FSM-движка холдов транзакций.
*   🗄️ [docs/specification/ledger-spec.md](specification/ledger-spec.md) — спецификация: проводки двойной записи и защита овердрафта.
*   🛡️ [docs/specification/fraud-spec.md](specification/fraud-spec.md) — спецификация: паттерн-скоринг рисков по черным спискам IP.
*   🧪 [docs/specification/pkg-math-spec.md](specification/pkg-math-spec.md) — спецификация функции: Zero-Allocation Fixed-Point математика копеек.
*   ⏰ [docs/specification/pkg-idempotency-spec.md](specification/pkg-idempotency-spec.md) — спецификация функции: пул идемпотентности и Sliding Window LRU.
*   🚦 [docs/specification/pkg-limiter-spec.md](specification/pkg-limiter-spec.md) — спецификация функции: Lock-Free CPU-Atomic CAS лимитер TPS.

---

## 🇺🇸 ENGLISH VERSION

Operational software requirement specification layout indexing the entire cluster ecosystem.

### 📋 1. SOFTWARE REQUIREMENT SPECIFICATIONS LAYER (DOCS/SRS/ - 7 FILES)
*   [docs/srs/cloud-routing-proxy-srs.md](srs/cloud-routing-proxy-srs.md) — SRS: L7 Edge Ingress load-balancer.
*   [docs/srs/gateway-srs.md](srs/gateway-srs.md) — SRS: PCI-DSS data tokenizer cipher bounds.
*   [docs/srs/processing-srs.md](srs/processing-srs.md) — SRS: Control Plane lifecycle state machine requirements.
*   [docs/srs/ledger-srs.md](srs/ledger-srs.md) — SRS: Data Plane double-entry accounting book constraints.
*   [docs/srs/fraud-srs.md](srs/fraud-srs.md) — SRS: Security Plane pattern-matching risk assessment bounds.
*   [docs/srs/core-pkg-srs.md](srs/core-pkg-srs.md) — SRS: Algorithmic complexity bounds for frameworks.
*   [docs/srs/infrastructure-srs.md](srs/infrastructure-srs.md) — SRS: Multi-stage Go Workspace build deployment rules.

### 📐 2. LOW-LEVEL COMPONENT SPECIFICATIONS LAYER (DOCS/SPECIFICATION/ - 8 FILES)
*   [docs/specification/cloud-routing-proxy-spec.md](specification/cloud-routing-proxy-spec.md) — spec: Configuration parsing and L7 network reverse proxy.
*   [docs/specification/gateway-spec.md](specification/gateway-spec.md) — spec: AES-256-GCM symmetric block cipher integration.
*   [docs/specification/processing-spec.md](specification/processing-spec.md) — spec: transactional FSM orchestrator execution flows.
*   [docs/specification/ledger-spec.md](specification/ledger-spec.md) — spec: double-entry immutable accounting logs database stubs.
*   [docs/specification/fraud-spec.md](specification/fraud-spec.md) — spec: sub-5ms risk score evaluation pipelines.
*   [docs/specification/pkg-math-spec.md](specification/pkg-math-spec.md) — function spec: Zero-heap fixed-point cents arithmetic.
*   [docs/specification/pkg-idempotency-spec.md](specification/pkg-idempotency-spec.md) — function spec: memory-sharded sliding window LRU pool.
*   [docs/specification/pkg-limiter-spec.md](specification/pkg-limiter-spec.md) — function spec: Mutex-free lock-free CAS token bucket limiter.
