package fsm

import (
	"errors"
	"sync"
)

var ErrInvalidStateTransition = errors.New("🔒 [FSM GUARD]: Нелегальный сдвиг фазы транзакции. Операция заблокирована")

// FiniteStateMachine контролирует инварианты жизненного цикла транзакций
// FIXED: Enforced oriented graph transitions validation to prevent payment state injection attacks
type FiniteStateMachine struct {
	mu         sync.RWMutex
	allowedMap map[string]map[string]struct{} // Текущий стейт ➔ Набор разрешенных следующих стейтов
}

// NewFiniteStateMachine строит автомат на базе разрешенной матрицы переходов, которую определяет бизнес-слой
func NewFiniteStateMachine(transitions map[string][]string) *FiniteStateMachine {
	fsm := &FiniteStateMachine{
		allowedMap: make(map[string]map[string]struct{}),
	}

	for fromState, toStates := range transitions {
		fsm.allowedMap[fromState] = make(map[string]struct{})
		for _, toState := range toStates {
			fsm.allowedMap[fromState][toState] = struct{}{}
		}
	}

	return fsm
}

// ValidateTransition атомарно проверяет, имеет ли право транзакция сдвинуться из currentState в nextState
func (f *FiniteStateMachine) ValidateTransition(currentState, nextState string) error {
	f.mu.RLock()
	allowedNextStates, exists := f.allowedMap[currentState]
	f.mu.RUnlock()

	if !exists {
		return ErrInvalidStateTransition
	}

	if _, allowed := allowedNextStates[nextState]; !allowed {
		return ErrInvalidStateTransition
	}

	return nil
}
