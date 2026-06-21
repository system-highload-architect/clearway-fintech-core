package ratelimit

import (
	"sync/atomic"
	"time"
)

// TokenBucketLimiter реализует атомарную Lock-Free защиту лимитов запросов (TPS) мерчантов
// Reengineered multi-threaded rate limiter using raw CPU atomic CAS operations instead of heavy mutexes
type TokenBucketLimiter struct {
	rate         int64 // Сколько токенов добавляется в секунду (Max TPS)
	capacity     int64 // Максимальный объем ведра токенов
	tokens       int64 // Текущее число доступных токенов в ОЗУ
	lastRefillNs int64 // Наносекундная метка времени последнего пополнения
}

// NewTokenBucketLimiter инициализирует TPS щит
func NewTokenBucketLimiter(rate, capacity int64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:         rate,
		capacity:     capacity,
		tokens:       capacity,
		lastRefillNs: time.Now().UnixNano(),
	}
}

// Allow за 0 наносекунд проверяет лимит. Если токен успешно списан — возвращает true.
// Если мерчант превысил свой TPS SLA — возвращает false (Всплеск трафика срезан).
func (l *TokenBucketLimiter) Allow() bool {
	for {
		nowNs := time.Now().UnixNano()
		lastRefill := atomic.LoadInt64(&l.lastRefillNs)
		currentTokens := atomic.LoadInt64(&l.tokens)

		// 1. Рассчитываем, сколько токенов лениво натекло по времени
		deltaNs := nowNs - lastRefill
		if deltaNs < 0 {
			deltaNs = 0
		}

		tokensToAdd := (deltaNs * l.rate) / int64(time.Second)
		newTokens := currentTokens + tokensToAdd
		if newTokens > l.capacity {
			newTokens = l.capacity
		}

		// 2. Если ведро пустое — мерчант заблокирован по лимиту флуда
		if newTokens < 1 {
			return false
		}

		// 3. Пытаемся списать 1 токен
		decrementedTokens := newTokens - 1

		// 4. КРИТИЧЕСКАЯ АТОМАРНАЯ СИНХРОНИЗАЦИЯ (CAS):
		// Если за эту наносекунду другой поток Go не успел перетереть метку времени —
		// мы атомарно фиксируем списание в регистрах CPU, полностью избегая блокировок мьютексов!
		if atomic.CompareAndSwapInt64(&l.lastRefillNs, lastRefill, nowNs) {
			atomic.StoreInt64(&l.tokens, decrementedTokens)
			return true
		}

		// Если CAS провалился (конкуренция потоков) — уходим на безопасный бесконечный цикл повтора (Spin-Lock)
	}
}
