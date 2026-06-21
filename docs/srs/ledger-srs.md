# 📋 SOFTWARE REQUIREMENT SPECIFICATION (SRS): TRANSACTIONAL LEDGER

[English version below]

## 🇷🇺 РУССКАЯ ВЕРСИЯ

### 1. Назначение и Финансовые Инварианты
Модуль `core/ledger` обязан выполнять роль неизменяемой бухгалтерской книги учета балансов кошельков мерчантов [2.1]. Система обязана гарантировать соблюдение закона двойной записи: сумма дебетов обязана строго равняться сумме кредитов [1.1].

### 2. Системные ограничения СУБД
*   **Append-Only Режим**: операции `UPDATE` и `DELETE` над проводками запрещены. Баланс мерчанта обязан вычисляться как сумма исторических проводок. Любое изменение — это новая запись `INSERT` [1.1].
*   **Защита овердрафта**: списание средств обязано атомарно прерываться с ошибкой, если на дебетовом балансе кошелька не хватает копеек [1.1]. Вычисления обязаны происходить строго через `fixedpoint.Money`.

---

## 🇺🇸 ENGLISH VERSION

### 1. Financial Equilibrium Scope
The `core/ledger` subsystem aggregates financial movements using strict double-entry ledger bookkeeping architectures [1.1, 2.1].

### 2. Technical Invariants
*   **Immutable Append-Only Log**: database execution layer must completely ban data updates and deletes.
*   **Overdraft Mitigation**: balances must never drop below zero; deficit conditions must instantly block commit actions via fixed-point checks [1.1].
