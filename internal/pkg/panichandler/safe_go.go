package panichandler

import (
	"fmt"
	"runtime/debug"
)

// LoggerAbstract задает минимальный анонимный интерфейс логера для пакета
type LoggerAbstract interface {
	Error(template string, args ...any)
}

// SafeGo запускает функцию в изолированной горутине с перехватом паник.
// Предотвращает крах всего бинарника платежной системы, бережно сохраняя рантайм живым.
// Wrapped concurrent execution threads inside a deterministic recover boundary to mitigate process crashes
func SafeGo(log LoggerAbstract, task func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := string(debug.Stack())
				if log != nil {
					log.Error("🔒 [PANIC ISOLATOR ALERT]: Перехвачен критический крах горутины: %v\nСтек трейс:\n%s", r, stackTrace)
				} else {
					fmt.Printf("🔒 [PANIC ISOLATOR ALERT]: %v\n%s\n", r, stackTrace)
				}
			}
		}()
		task()
	}()
}
