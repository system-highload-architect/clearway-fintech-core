# 📋 SOFTWARE REQUIREMENT SPECIFICATION (SRS): INFRASTRUCTURE LAB

[English version below]

## 🇷🇺 РУССКАЯ ВЕРСИЯ

### 1. Требования к компиляции и изоляции модулей
Инфраструктурный слой регламентирует правила сборки монорепозитория и требования к развертыванию изолированных бинарных пакетов [2.1].

### 2. Критерии сшивки модулей Go Workspaces
*   **Сшивка go.work**: все 4 модуля (`/core`, `/services/*`, `/internal`, `/pb`) обязаны жестко компилироваться на единой версии **`go 1.25.0`** [2.1]. Смешение версий в воркспейсе запрещено.
*   **Независимость деплоя**: структура папок обязана позволять скомпилировать любую директорию из `core/*` в виде отдельного Docker-контейнера или пода Kubernetes, так как они общаются строго по gRPC-контрактам из `pb/` [2.1].

---

## 🇺🇸 ENGLISH VERSION

### 1. Compilation & Monorepo Infrastructure Scope
Defines compilation boundaries, strict multi-module packaging configurations, and deployment runtime rules [2.1].

### 2. Multi-Module Scaffolding Rules
*   **Unified Workspace**: every isolated dependency path inside `go.work` must enforce a matching, unfragmented **`go 1.25.0`** compiler standard [2.1].
*   **Decoupled Partitioning**: component structure dictates that business sub-domains must maintain absolute transport independence.
