package registry

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrCommandRegistered = errors.New("🔒 [REGISTRY ERROR]: Данный код команды уже жестко заблокирован в ОЗУ")
	ErrCommandNotFound   = errors.New("🔒 [REGISTRY ERROR]: Неизвестный код команды. Операция отклонена шлюзом")
)

// ExecutionHandler задает сигнатуру чистой, абстрактной b2b функции обработки
type ExecutionHandler func(ctx context.Context, payload []byte) ([]byte, error)

// ActionDispatcher реализует наносекундный роутер команд без switch-case и if-else
// FIXED: Engineered a reflection-free map dispatcher to minimize processing branch mispredictions
type ActionDispatcher struct {
	mu       sync.RWMutex
	handlers map[string]ExecutionHandler
}

// NewActionDispatcher инициализирует пустую таблицу диспетчеризации
func NewActionDispatcher() *ActionDispatcher {
	return &ActionDispatcher{
		handlers: make(map[string]ExecutionHandler),
	}
}

// RegisterAction намертво биндит функцию к строковому коду команды (например, "ISO_8583_PURCHASE")
func (d *ActionDispatcher) RegisterAction(cmdCode string, handler ExecutionHandler) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.handlers[cmdCode]; exists {
		return ErrCommandRegistered
	}

	d.handlers[cmdCode] = handler
	return nil
}

// ExecuteAction за O(1) извлекает функцию из ОЗУ и запускает её исполнение
func (d *ActionDispatcher) ExecuteAction(ctx context.Context, cmdCode string, payload []byte) ([]byte, error) {
	d.mu.RLock()
	handler, exists := d.handlers[cmdCode]
	d.mu.RUnlock()

	if !exists {
		return nil, ErrCommandNotFound
	}

	// Вызываем функцию напрямую без рефлексии
	return handler(ctx, payload)
}
