# 📋 SOFTWARE REQUIREMENT SPECIFICATION (SRS): CLOUD ROUTING PROXY

[English version below]

## 🇷🇺 РУССКАЯ ВЕРСИЯ

### 1. Функциональное назначение контура
Компонент `cloud-routing-proxy` является единственным бронированным Ingress-узлом периметра сети. Он обязан осуществлять терминацию входящего HTTP/REST трафика, раздавать статические файлы UI-дашборда мерчанта и проксировать запросы на внутренние сегменты кластера [2.1].

### 2. Системные требования и SLA
*   **Изоляция конфигурации**: сервер обязан инициализироваться на базе системного абстрактного шасси `internal/chassis` и считывать файл `config.yaml` строго на этапе старта домена [2.1].
*   **Метрики производительности**: задержка на этапе маршрутизации (Proxy Latency) обязана составлять $\le 3.0\text{мс}$ для 99.9-го процентиля при пиковой нагрузке до 30 000 RPS.
*   **Favicon & Favicon Guard**: запросы к `/favicon.ico` обязаны аппаратно перехватываться за 1 такт и возвращать `204 No Content` без аллокаций памяти.

---

## 🇺🇸 ENGLISH VERSION

### 1. Functional Scope & Purpose
The `cloud-routing-proxy` microservice orchestrates traffic boundaries at the edge of the secure perimeter. It acts as a stateless Ingress load-balancer rendering UI assets and multiplexing HTTP payloads into upstream endpoints [2.1].

### 2. Operational Constraints & SLA
*   **Routing Overhead**: strict processing latency budget is bound to $\le 3.0\text{ms}$ at p99.9 under 30k RPS load limits.
*   **Asset Decoupling**: static file streaming must avoid blocking application threads by leveraging non-blocking OS system calls.
