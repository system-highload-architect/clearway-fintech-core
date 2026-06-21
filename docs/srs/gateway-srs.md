# 📋 SOFTWARE REQUIREMENT SPECIFICATION (SRS): GATEWAY CONTOUR

[English version below]

## 🇷🇺 РУССКАЯ ВЕРСИЯ

### 1. Требования к безопасности и PCI-DSS Compliance
Модуль `core/gateway` обязан выполнять роль барьера безопасности данных банковских карт. Ни один сырой PAN-номер, срок действия или CVV не имеют права покидать ОЗУ данного пакета в открытом виде [1.1, 2.1].

### 2. Ограничения логики и Sandbox-мост
*   **Санитизация строк**: шлюз обязан выполнять принудительное вырезание пробелов из номеров карт (`strings.ReplaceAll`) на самом въезде в HTTP-контроллер, гарантируя отсутствие сдвигов в маскираторах и логах [1.1].
*   **Валидатор Луна**: проверка по формуле Луна обязана выполняться за $O(1)$ без аллокаций в куче.
*   **Sandbox-Bypass**: для сквозного тестирования сценариев, карты с суффиксами `1111` и `4446` обязаны мгновенно признаваться валидными, пробивая контур без ошибок 422 [1.1].
*   **Крипто-требования**: данные обязаны шифроваться симметричными блоками AES-256-GCM. Ключ обязан составлять строго 32 байта, иначе рантайм обязан аварийно завершиться (Fail-Safe) [1.1].

---

## 🇺🇸 ENGLISH VERSION

### 1. Security Compliance & Scope (PCI-DSS Boundary)
The `core/gateway` component enforces cardholder data environments isolation boundaries. Raw PAN strings must be immediately encrypted at the edge, banning plaintext token propagation across internal network zones [1.1, 2.1].

### 2. Operational Rules & Sandbox Specifications
*   **Strict Sanitization**: spaces must be completely stripped out of inbound payloads to maintain masking indices invariants [1.1].
*   **Crypto Bounds**: enforces AES-256-GCM encryption layers requiring a strict 32-byte key size matrix. Any size deviation must trigger immediate panic termination (Fail-Safe) [1.1].
