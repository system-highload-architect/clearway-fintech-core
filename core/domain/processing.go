package domain

import (
	"clearway-fintech-core/internal/pkg/fixedpoint"
	"time"
)

// Константы жестко типизированных стейтов транзакции для FSM автомата
const (
	StateNew           = "NEW"
	StateFraudChecking = "FRAUD_CHECKING"
	StateAcquiringHold = "ACQUIRING_HOLD"
	StateHeld          = "HELD"
	StateCaptured      = "CAPTURED"
	StateReversed      = "REVERSED"
	StateFailed        = "FAILED"
)

// Transaction описывает полную финансовую проводку внутри процессора
type Transaction struct {
	ID             string           // У集中ированный ID платежа
	MerchantID     string           // Кому принадлежат деньги
	PaymentToken   string           // Обезличенная карта из Gateway
	Amount         fixedpoint.Money // Точная сумма в копейках (int64)
	CurrentState   string           // Текущая фаза из fsm.FiniteStateMachine
	IdempotencyKey string           // Сквозной защитный b2b-ключ
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
