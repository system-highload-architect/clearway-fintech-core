# 🌐 LOW-LEVEL SPECIFICATION: CLOUD ROUTING PROXY / API INGRESS

[English version below]

## 🇷🇺 РУССКАЯ ВЕРСИЯ

### 1. Архитектура Шасси и Инициализация Конфигурации
Компонент `services/cloud-routing-proxy` (Порт `:8080`) инициализируется на базе абстрактного системного шасси `internal/chassis` [2.1]. На этапе старта функция `chassis.LoadConfigAbstract()` парсит файл `config.yaml` в структуру `chassis.BaseConfig`, выставляя порты рантайма.

### 📊 Диаграмма Маршрутизации и Раздачи Статики (Ingress Invariant):
```mermaid
sequenceDiagram
    autonumber
    participant Client as 🌐 UI Браузер
    participant Proxy as 🚀 cloud-routing-proxy
    participant Gateway as 🔒 core/gateway (HTTP REST)
    participant Disk as 💾 web/static/ (Дисковое ОЗУ)

    Client->>Proxy: GET / (Запрос дашборда)
    Proxy->>Disk: Чтение локального файла web/index.html без StripPrefix
    Disk-->>Proxy: Буфер данных index.html
    Proxy-->>Client: HTTP 200 OK HTML (Стили подгружаются из /static/css/)

    Client->>Proxy: POST /api/v1/gateway/tokenize
    Proxy->>Proxy: L7 Маршрутизация каскада байт по внутреннему мосту
    Proxy->>Gateway: Реверс-проксирование HTTP POST во внутренний порт модуля
    Gateway-->>Proxy: JSON: {"success": true, "token": "tok_pki_*"}
    Proxy-->>Client: HTTP 200 OK Token Response
```

---

## 🇺🇸 ENGLISH VERSION

### 1. Ingress Layer Routing Mechanics
The Ingress proxy maps public-facing URI endpoints to isolated container ports. It hosts a native file server mounting the `web/` directory directly under the strict `/static/` context location [2.1].
