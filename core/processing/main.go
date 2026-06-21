package main

import (
	"clearway-fintech-core/internal/pkg/idempotency"

	"time"
)

const (
	TxTypeInstant = "TX_INSTANT" // Моментальный перевод (1 мин)
	TxTypeSbpQR   = "TX_SBP_QR"  // Оплата по коду (30 мин)
	TxTypeCredit  = "TX_CREDIT"  // Краткосрочный лимит (2 часа)
)

// TODO
func main() {
	// Жестко связанная мапа b2b-политик времени
	fintechTTLConfig := map[string]time.Duration{
		TxTypeInstant: 1 * time.Minute,
		TxTypeSbpQR:   30 * time.Minute,
		TxTypeCredit:  2 * time.Hour,
	}

	// Создаем пул, который идеально сопряжен с нашими строковыми контрактами!
	_ = idempotency.NewIdempotencyPoolManager(50000, fintechTTLConfig)
}
