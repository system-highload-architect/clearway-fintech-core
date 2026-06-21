# 📋 SOFTWARE REQUIREMENT SPECIFICATION (SRS): FRAUD RADAR

[English version below]

## 🇷🇺 РУССКАЯ ВЕРСИЯ

### 1. Функциональные требования к скорингу
Модуль `core/fraud` обязан на лету анализировать контекст транзакции, выявлять паттерны бот-активностей и блокировать мошеннические списания до того, как запрос уйдет в банк-эквайер [1.1, 2.1].

### 2. Пороговые значения и SLA
*   **sub-5ms Порог задержки**: время оценки рисков одной транзакции не должно превышать **5 миллисекунд** для 99.99-го процентиля. Использование регулярных выражений запрещено [1.1].
*   **Фрод-Критерии**: система обязана мгновенно разворачивать флаг блокировки `is_fraudulent = true` при совпадении IP с черным списком ОЗУ или при объеме разового платежа по тестовому БИНу свыше 100 000.00 рублей [1.1].

---

## 🇺🇸 ENGLISH VERSION

### 1. Security Plane Risk Assessment Scope
The `core/fraud` layer evaluates telemetry context vectors to score and drop malicious payment events before authorization commits [1.1, 2.1].

### 2. Telemetry SLA Thresholds
*   **Microsecond Execution**: risk verification cycles must complete within a strict $\le 5\text{ms}$ budget limit.
*   **Heuristic Slicing**: instantly flag and drop payments exceeding volume bounds on generic testing bins.
