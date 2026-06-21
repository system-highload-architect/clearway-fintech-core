package idempotency

import (
	"sync"
	"time"
)

// IdempotencyPoolManager координирует независимые изолированные инстансы LRU-кэшей,
// разделенные по строковым тегам категорий операций, которые определяет сам сервис.
// FIXED: Swapped arrays with tag-based maps to guarantee compile-time and runtime cohesion
type IdempotencyPoolManager struct {
	mu         sync.RWMutex
	pools      map[string]*IdempotencyLRUEngine
	ttlMapping map[string]time.Duration // Тег категории ➔ Фиксированное время жизни (TTL)
}

// NewIdempotencyPoolManager инициализирует менеджер пулов на основе мапы конфигурации,
// переданной сервисом. Ключ — это имя операции (например, "instant", "sbp_qr"), значение — её TTL.
func NewIdempotencyPoolManager(capacityPerPool int, ttlConfig map[string]time.Duration) *IdempotencyPoolManager {
	m := &IdempotencyPoolManager{
		pools:      make(map[string]*IdempotencyLRUEngine),
		ttlMapping: make(map[string]time.Duration),
	}

	// Атомарно аллоцируем изолированные шарды ОЗУ под каждый зарегистрированный тег
	for tag, ttl := range ttlConfig {
		m.ttlMapping[tag] = ttl
		m.pools[tag] = NewIdempotencyLRUEngine(capacityPerPool)
	}

	return m
}

// CheckOrLockTransaction маршрутизирует транзакцию в нужный пул на основе строкового тега категории
func (m *IdempotencyPoolManager) CheckOrLockTransaction(categoryTag string, txID string) (TxResult, bool, error) {
	m.mu.RLock()
	targetPool, existsPool := m.pools[categoryTag]
	defaultTTL, existsTTL := m.ttlMapping[categoryTag]
	m.mu.RUnlock()

	// Страховой барьер: если сервис передал неизвестный тег, лениво инициализируем дефолтный 5-минутный контур
	if !existsPool || !existsTTL {
		m.mu.Lock()
		targetPool, existsPool = m.pools[categoryTag]
		if !existsPool {
			defaultTTL = 5 * time.Minute
			m.ttlMapping[categoryTag] = defaultTTL
			targetPool = NewIdempotencyLRUEngine(1000) // Защитная емкость
			m.pools[categoryTag] = targetPool
		}
		m.mu.Unlock()
	}

	// Вызываем локальный безэлокационный LRU-фильтр конкретного изолированного шарда
	return targetPool.CheckOrLock(txID, defaultTTL)
}

// UpdateTransactionStep осуществляет Sliding Window сброс дедлайна внутри целевого тег-контура
func (m *IdempotencyPoolManager) UpdateTransactionStep(categoryTag string, txID string, nextStatus string, customStepTTL time.Duration) {
	m.mu.RLock()
	targetPool, exists := m.pools[categoryTag]
	m.mu.RUnlock()

	if exists {
		targetPool.UpdateProgressState(txID, nextStatus, customStepTTL)
	}
}

// GetCategoryTTL позволяет сервису нативно запросить базовый TTL для любого установленного типа операции
func (m *IdempotencyPoolManager) GetCategoryTTL(categoryTag string) (time.Duration, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ttl, exists := m.ttlMapping[categoryTag]
	return ttl, exists
}

// StartGlobalJanitorDaemon запускает раздельный последовательный обход пулов по таймеру
func (m *IdempotencyPoolManager) StartGlobalJanitorDaemon(stopChan chan struct{}, checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)
	go func() {
		for {
			select {
			case <-stopChan:
				ticker.Stop()
				return
			case <-ticker.C:
				m.mu.RLock()
				// Поочередно зачищаем каждый пул. Обход гарантированно эффективен,
				// так как внутри одного тега все дедлайны выстроены в строгом хронологическом порядке!
				for _, pool := range m.pools {
					_ = pool.EvictExpiredJanitorLoop()
				}
				m.mu.RUnlock()
			}
		}
	}()
}
