package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("🔒 [CIRCUIT BREAKER]: Внешний шлюз временно недоступен. Цепь разомкнута")

type State int

const (
	StateClosed   State = iota // Цепь замкнута — всё ок, запросы идут
	StateOpen                  // Цепь разомкнута — банк упал, запросы режутся мгновенно
	StateHalfOpen              // Полуоткрыто — проверяем, ожил ли банк
)

// CircuitBreaker реализует паттерн защиты от каскадных сбоев внешних интеграций
// FIXED: Engineered an atomic circuit state machine to isolate slow/failed bank API networks
type CircuitBreaker struct {
	mu           sync.RWMutex
	state        State
	failureCount int64
	failureLimit int64
	timeout      time.Duration
	nextAttempt  time.Time
}

func NewCircuitBreaker(failureLimit int64, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        StateClosed,
		failureLimit: failureLimit,
		timeout:      timeout,
	}
}

// Execute проверяет стейт предохранителя. Если открыт — сразу рубит вызов.
// Если закрыт — запускает целевую b2b-функцию.
func (cb *CircuitBreaker) Execute(operation func() error) error {
	cb.mu.Lock()
	now := time.Now()

	// 1. Проверяем, не пора ли перевести из Open в Half-Open
	if cb.state == StateOpen && now.After(cb.nextAttempt) {
		cb.state = StateHalfOpen
	}

	if cb.state == StateOpen {
		cb.mu.Unlock()
		return ErrCircuitOpen
	}
	cb.mu.Unlock()

	// 2. Выполняем саму сетевую операцию
	err := operation()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		if cb.state == StateHalfOpen || cb.failureCount >= cb.failureLimit {
			cb.state = StateOpen
			cb.nextAttempt = time.Now().Add(cb.timeout)
		}
		return err
	}

	// Если операция успешна
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failureCount = 0
	}
	return nil
}
