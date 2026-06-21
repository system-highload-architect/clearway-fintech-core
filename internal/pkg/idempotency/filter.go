package idempotency

import (
	"errors"
	"sync"
	"time"
)

var ErrTxExpired = errors.New("🔒 [PROCESSING GUARD]: Время жизни транзакции истекло в инфлайте")

// TxResult запечатывает сохраненный статус и ответ транзакции для кэша
type TxResult struct {
	Status    string        // Текущий статус ("IN_FLIGHT", "SUCCESS", "FAILED")
	Payload   []byte        // Сериализованный gRPC-ответ для возврата клиенту
	ExpiresAt time.Time     // Абсолютная точка смерти текущего шага
	StepTTL   time.Duration // Контрактный TTL для этого конкретного шага транзакции
}

// Node представляет элемент двусвязного списка в ОЗУ для LRU-вытеснения
type Node struct {
	Key        string
	Value      TxResult
	Prev, Next *Node
}

// IdempotencyLRUEngine объединяет хэш-мапу O(1) и Doubly Linked List для Sliding TTL контроля
// FIXED: Engineered a high-performance sliding window cache to prevent memory leaks during long-running banking steps
type IdempotencyLRUEngine struct {
	mu         sync.RWMutex
	cache      map[string]*Node
	head, tail *Node
	capacity   int // Максимальный лимит горячих операций в ОЗУ
}

// NewIdempotencyLRUEngine инициализирует транзакционный барьер с лимитом емкости
func NewIdempotencyLRUEngine(capacity int) *IdempotencyLRUEngine {
	return &IdempotencyLRUEngine{
		cache:    make(map[string]*Node),
		capacity: capacity,
	}
}

// CheckOrLock проверяет статус транзакции.
// Если это дубликат — возвращает результат. Если операция уникальна — лочит и ставит в кэш.
func (e *IdempotencyLRUEngine) CheckOrLock(key string, defaultTTL time.Duration) (TxResult, bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()

	// 1. Если транзакция найдена в ОЗУ
	if node, found := e.cache[key]; found {
		// Проверяем, не протухла ли она по времени, пока висела
		if now.After(node.Value.ExpiresAt) {
			e.removeNode(node)
			delete(e.cache, key)
			return TxResult{}, false, ErrTxExpired
		}

		// Кэш-хит: сдвигаем ноду на вершину кэша (Head), так как к ней обратились
		e.moveToHead(node)
		return node.Value, true, nil
	}

	// 2. Если транзакция уникальна — создаем новую ноду в статусе IN_FLIGHT
	res := TxResult{
		Status:    "PROCESSING_IN_FLIGHT",
		ExpiresAt: now.Add(defaultTTL),
		StepTTL:   defaultTTL,
	}

	newNode := &Node{Key: key, Value: res}

	// Проверяем лимит емкости перед аллокацией (Защита от OOM)
	if len(e.cache) >= e.capacity {
		// Жестко вытесняем самую старую неактивную операцию из хвоста (Tail)
		oldTail := e.tail
		if oldTail != nil {
			e.removeNode(oldTail)
			delete(e.cache, oldTail.Key)
		}
	}

	e.addToHead(newNode)
	e.cache[key] = newNode

	return res, false, nil
}

// UpdateProgressState — АТОМАРНЫЙ СДВИГ НА ВЕРШИНУ КЭША (Твой Sliding Window паттерн).
// Если процесс сдвинулся с места, мы обновляем статус транзакции и СБРАСЫВАЕМ её таймер на исходный размер.
func (e *IdempotencyLRUEngine) UpdateProgressState(key string, nextStatus string, customStepTTL time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	node, found := e.cache[key]
	if !found {
		return
	}

	// ИСПРАВЛЕНО (Sliding Window): Сбрасываем дедлайн таймаута обратно на полную величину шага
	node.Value.Status = nextStatus
	node.Value.StepTTL = customStepTTL
	node.Value.ExpiresAt = time.Now().Add(customStepTTL)

	// Перемещаем ноду на самый верх кэша (к Head)
	e.moveToHead(node)
}

/* ВНУТРЕННИЕ СВЕРХБЫСТРЫЕ МЕТОДЫ ДВУСВЯЗНОГО СПИСКА (БЕЗ МЬЮТЕКСОВ, O(1)) */

func (e *IdempotencyLRUEngine) addToHead(node *Node) {
	node.Next = e.head
	node.Prev = nil
	if e.head != nil {
		e.head.Prev = node
	}
	e.head = node
	if e.tail == nil {
		e.tail = node
	}
}

func (e *IdempotencyLRUEngine) removeNode(node *Node) {
	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		e.head = node.Next
	}
	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		e.tail = node.Prev
	}
}

func (e *IdempotencyLRUEngine) moveToHead(node *Node) {
	e.removeNode(node)
	e.addToHead(node)
}

// EvictExpiredJanitorLoop запускается фоновым воркером для зачистки ОЗУ от мертвых инфлайт-транзакций
func (e *IdempotencyLRUEngine) EvictExpiredJanitorLoop() int {
	e.mu.Lock()
	defer e.mu.Unlock()

	evictedCount := 0
	now := time.Now()

	// Идем с хвоста списка (там лежат самые старые операции)
	curr := e.tail
	for curr != nil {
		nextTarget := curr.Prev // Запоминаем ссылку перед удалением
		if now.After(curr.Value.ExpiresAt) {
			e.removeNode(curr)
			delete(e.cache, curr.Key)
			evictedCount++
		} else {
			// Так как список отсортирован по времени использования, как только мы наткнулись
			// на живую транзакцию — прерываем обход, экономя такты CPU!
			break
		}
		curr = nextTarget
	}
	return evictedCount
}
