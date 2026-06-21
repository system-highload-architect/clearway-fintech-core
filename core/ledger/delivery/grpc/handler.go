package grpc

import (
	"context"
	"fmt"

	"clearway-fintech-core/core/domain"
	"clearway-fintech-core/internal/pkg/fixedpoint"
	"clearway-fintech-core/pb/gen" // Сгенерированные контракты wallet.pb.go
)

// LedgerUseCaseInterface задает границы локального b2b usecase
type LedgerUseCaseInterface interface {
	CommitDoubleEntry(ctx context.Context, debitWalletID, creditWalletID string, amount fixedpoint.Money, txID, description string) (*domain.Wallet, *domain.Wallet, error)
	GetBalance(ctx context.Context, walletID string) (domain.Wallet, error)
}

type LedgerGrpcHandler struct {
	gen.UnimplementedLedgerWalletEngineServer
	useCase LedgerUseCaseInterface
}

func NewLedgerGrpcHandler(uc LedgerUseCaseInterface) *LedgerGrpcHandler {
	return &LedgerGrpcHandler{
		useCase: uc,
	}
}

// CommitDoubleEntryTransaction принимает gRPC-вызов списания/зачисления средств мерчантов
// FIXED: Transformed raw integer proto fields straight into immutable fixedpoint structures
func (h *LedgerGrpcHandler) CommitDoubleEntryTransaction(ctx context.Context, req *gen.LedgerEntryRequest) (*gen.LedgerEntryResponse, error) {
	if req.AmountUnits <= 0 {
		return &gen.LedgerEntryResponse{
			IsCommitted:  false,
			ErrorMessage: "🔒 [gRPC LEDGER]: Сумма проводки должна быть строго больше нуля",
		}, nil
	}

	// Укладываем сырые int64 копейки из сети в наш сверхбыстрый тип
	amountMoney := fixedpoint.NewMoneyFromInt64(req.AmountUnits)

	// Запускаем каноническую двойную запись в ОЗУ-репозитории
	debitW, creditW, err := h.useCase.CommitDoubleEntry(
		ctx,
		req.DebitWalletId,
		req.CreditWalletId,
		amountMoney,
		req.TxId,
		req.Description,
	)

	if err != nil {
		return &gen.LedgerEntryResponse{
			IsCommitted:  false,
			ErrorMessage: fmt.Sprintf("🔒 [gRPC LEDGER TRANSACTION FAILURE]: %v", err),
		}, nil
	}

	return &gen.LedgerEntryResponse{
		IsCommitted:  true,
		NewDebitBal:  debitW.AvailableBal.Units,
		NewCreditBal: creditW.AvailableBal.Units,
	}, nil
}

// GetWalletBalance возвращает текущее состояние и резервы кошелька по gRPC
func (h *LedgerGrpcHandler) GetWalletBalance(ctx context.Context, req *gen.BalanceRequest) (*gen.BalanceResponse, error) {
	wallet, err := h.useCase.GetBalance(ctx, req.WalletId)
	if err != nil {
		return nil, err
	}

	return &gen.BalanceResponse{
		AvailableUnits: wallet.AvailableBal.Units,
		ReservedUnits:  wallet.ReservedBal.Units,
		Currency:       wallet.Currency,
	}, nil
}
