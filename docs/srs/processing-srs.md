# 📋 SOFTWARE REQUIREMENT SPECIFICATION (SRS): PROCESSING ENGINE

[English version below]

## 🇷🇺 РУССКАЯ ВЕРСИЯ

### 1. Функциональные требования к оркестрации
Модуль `core/processing` обязан координировать транзакционные потоки, проверять легальность сдвига фаз платежа через стейт-машину и блокировать атаки повторного списания (Double-Spending) [1.1, 2.1].

### 2. Матрица FSM и Лимиты ОЗУ
*   **Неизменяемость переходов**: смена стейтов обязана подчиняться ориентированному графу: `NEW ➔ FRAUD_CHECKING ➔ ACQUIRING_HOLD ➔ HELD ➔ CAPTURED`. Любой нелегальный сдвиг обязан мгновенно переводить платеж в `FAILED` [1.1].
*   **Пул Идемпотентности**: движок обязан использовать маппированные Sliding Window кэши. При обновлении стейта транзакции время жизни сессии (TTL) обязано автоматически сбрасываться на полную исходную величину текущего шага.

---

## 🇺🇸 ENGLISH VERSION

### 1. Control Plane Lifecycle Requirements
The `core/processing` domain validates state transitions, orchestrates inter-service workflows, and mitigates race conditions across multi-threaded operations [1.1, 2.1].

### 2. Architectural Invariants
*   **Transition Enforcement**: state manipulation rules are governed by a deterministic oriented transition graph matrix.
*   **Sliding Expiration**: every progress upgrade step within the state machine must reset volatile cache timers to their full initial bucket values.
