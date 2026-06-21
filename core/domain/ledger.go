package domain

import (
	"clearway-fintech-core/internal/pkg/fixedpoint"
	"time"
)

// Wallet представляет балансовый паспорт кошелька мерчанта или банка
type Wallet struct {
	ID           string           // Уникальный b2b-номер счета (UUID/PAN)
	MerchantID   string           // Владелец кошелька (Мерчант / Эквайер)
	AvailableBal fixedpoint.Money // Доступные средства (Units в копейках)
	ReservedBal  fixedpoint.Money // Замороженные на этапе HOLD средства
	Currency     string           // ISO-код валюты (RUB, USD)
	UpdatedAt    time.Time
}

// LedgerEntry запечатывает атомарную запись Double-Entry лога.
// Запись полностью иммутабельна. Сумма дебетов всегда равна сумме кредитов.
// FIXED: Formed structural transaction data layout to guarantee audit compliance
type LedgerEntry struct {
	EntryID        string           // Уникальный ID бухгалтерской проводки
	TxID           string           // Ссылка на ID родительской транзакции
	DebitWalletID  string           // С какого счета списали
	CreditWalletID string           // На какой счет зачислили
	Amount         fixedpoint.Money // Точная сумма в копейках
	Description    string           // Назначение платежа
	CreatedAt      time.Time        // Время проведения по часам ОЗУ
}
