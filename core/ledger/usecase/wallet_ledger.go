package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"clearway-fintech-core/core/domain"
	"clearway-fintech-core/internal/pkg/fixedpoint"
)

var (
	ErrWalletNotFound   = errors.New("🔒 [LEDGER CORE]: Целевой кошелек не найден в системе")
	ErrCurrencyMismatch = errors.New("🔒 [LEDGER CORE]: Кросс-валютные переводы внутри Ledger без конвертера запрещены")
)

// LedgerUseCase оркестрирует атомарный балансовый учет двойной записи
// FIXED: Built memory-sharded transaction ledger engine with strict overflow/overdraft guards
type LedgerUseCase struct {
	mu      sync.RWMutex
	wallets map[string]*domain.Wallet
	entries map[string]domain.LedgerEntry // Append-Only Лог проводок (EntryID -> Entry)
}

// NewLedgerUseCase инициализирует чистый NoSQL-эмулятор бухгалтерской книги
func NewLedgerUseCase() *LedgerUseCase {
	// Для демонстрации лениво создадим два b2b-кошелька мерчанта и системного банка
	mockWallets := make(map[string]*domain.Wallet)

	// Кошелек покупателя (Давид): изначально зальем туда 50 000.00 рублей (5000000 копеек)
	mockWallets["wallet_david_buyer"] = &domain.Wallet{
		ID:           "wallet_david_buyer",
		MerchantID:   "david_pki_inc",
		AvailableBal: fixedpoint.NewMoneyFromInt64(5000000),
		Currency:     "RUB",
		UpdatedAt:    time.Now(),
	}

	// Кошелек b2b-Магазина (Мерчант), баланс 0.00 рублей
	mockWallets["wallet_merchant_shop"] = &domain.Wallet{
		ID:           "wallet_merchant_shop",
		MerchantID:   "merchant_shop_id",
		AvailableBal: fixedpoint.NewMoneyFromInt64(0),
		Currency:     "RUB",
		UpdatedAt:    time.Now(),
	}

	return &LedgerUseCase{
		wallets: mockWallets,
		entries: make(map[string]domain.LedgerEntry),
	}
}

// CommitDoubleEntry атомарно выполняет каноническую двойную запись: Дебет ➔ Кредит
func (uc *LedgerUseCase) CommitDoubleEntry(
	ctx context.Context,
	debitWalletID, creditWalletID string,
	amount fixedpoint.Money,
	txID, description string,
) (*domain.Wallet, *domain.Wallet, error) {

	uc.mu.Lock()
	defer uc.mu.Unlock()

	// 1. Извлекаем кошельки из ОЗУ-репозитория
	debitWallet, existsFrom := uc.wallets[debitWalletID]
	creditWallet, existsTo := uc.wallets[creditWalletID]
	if !existsFrom || !existsTo {
		return nil, nil, ErrWalletNotFound
	}

	// 2. Валютный контроль периметра
	if debitWallet.Currency != creditWallet.Currency {
		return nil, nil, ErrCurrencyMismatch
	}

	// 3. АТОМАРНОЕ ВЫЧИТАНИЕ (Проверка на овердрафт): Метод Sub сам выбросит ошибку,
	// если у покупателя недостаточно копеек на балансе!
	newDebitMoney, err := debitWallet.AvailableBal.Sub(amount)
	if err != nil {
		return nil, nil, err
	}

	// 4. АТОМАРНОЕ ЗАЧИСЛЕНИЕ (Защита от переполнения int64)
	newCreditMoney, err := creditWallet.AvailableBal.Add(amount)
	if err != nil {
		return nil, nil, err
	}

	// 5. Фиксируем изменения балансов в памяти
	debitWallet.AvailableBal = newDebitMoney
	debitWallet.UpdatedAt = time.Now()

	creditWallet.AvailableBal = newCreditMoney
	creditWallet.UpdatedAt = time.Now()

	// 6. Укладываем проводку в APPEND-ONLY ЛОГ (Иммутабельный след)
	entryID := fmt.Sprintf("entry_%d", time.Now().UnixNano())
	entry := domain.LedgerEntry{
		EntryID:        entryID,
		TxID:           txID,
		DebitWalletID:  debitWalletID,
		CreditWalletID: creditWalletID,
		Amount:         amount,
		Description:    description,
		CreatedAt:      time.Now(),
	}
	uc.entries[entryID] = entry

	return debitWallet, creditWallet, nil
}

// GetBalance возвращает стейт кошелька мерчанта для аудита
func (uc *LedgerUseCase) GetBalance(ctx context.Context, walletID string) (domain.Wallet, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	w, exists := uc.wallets[walletID]
	if !exists {
		return domain.Wallet{}, ErrWalletNotFound
	}
	return *w, nil
}
