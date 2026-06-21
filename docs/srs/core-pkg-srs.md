# 📋 SOFTWARE REQUIREMENT SPECIFICATION (SRS): CORE PKG ALGORITHMS

[English version below]

## 🇷🇺 РУССКАЯ ВЕРСИЯ

### 1. Требования к вычислительной сложности
Пакет общего назначения `internal/pkg` инкапсулирует наносекундный инструментарий, обеспечивающий производительность бизнес-модулей ядра.

### 2. Критерии эффективности ОЗУ и CPU
*   **Fixed-Point Математика**: вычисления обязаны происходить в плоском `int64` (копейки) с ручным побайтовым парсингом, исключая float64 и накладные расходы `math/big` на аллокации памяти в куче [1.1].
*   **Lock-Free TPS Лимитер**: ограничение флуда обязано работать без мьютексов, используя исключительно атомарные процессорные инструкции `Compare-And-Swap` (CAS), предотвращая клинчи многопоточности [1.1].
*   **Отказоустойчивость**: пакет обязаны снабжаться автоматами изоляции сбоев `circuitbreaker` и экспоненциальными ретраерами с криптографическим шумом `Jitter` [1.1].

---

## 🇺🇸 ENGLISH VERSION

### 1. Algorithmic Complexity Benchmarks
The `internal/pkg` isolates generic performance primitives, enforcing bounded memory overhead limits across low-level cycles.

### 2. Resource Management Primitives
*   **Zero-Heap Math**: monetary calculations must bypass floating-point variables via optimized integer fixed-point scales [1.1].
*   **Lock-Free CAS Limiters**: token bucket flood gates must leverage primitive atomic CPU instructions instead of heavy operating system mutex locks [1.1].
