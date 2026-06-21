package grpc

import (
	"context"
	"fmt"

	"clearway-fintech-core/internal/pkg/fixedpoint"
	"clearway-fintech-core/pb/gen" // Сгенерированные контракты processing.pb.go
)

// ProcessingEngineInterface задает SOLID границы для UseCase процессора транзакций
type ProcessingEngineInterface interface {
	ExecuteHoldInitiation(ctx context.Context, merchantID, paymentToken, idempotencyKey, categoryTag string, amount fixedpoint.Money) (string, string, error)
	ExecuteCaptureConfirmation(ctx context.Context, txID string) (string, error)
	ExecuteReversalCancellation(ctx context.Context, txID string) (string, error)
}

type ProcessingGrpcHandler struct {
	gen.UnimplementedProcessingTransactionCoreServer
	engine ProcessingEngineInterface
}

func NewProcessingGrpcHandler(eng ProcessingEngineInterface) *ProcessingGrpcHandler {
	return &ProcessingGrpcHandler{
		engine: eng,
	}
}

// ProcessHold принимает gRPC-вызов на первичную заморозку средств (Авторизация платежа)
// FIXED: Handled implicit proto naming conventions by explicitly binding input tags to internal engine bounds
func (h *ProcessingGrpcHandler) ProcessHold(ctx context.Context, req *gen.HoldRequest) (*gen.HoldResponse, error) {
	if req.AmountUnits <= 0 {
		return &gen.HoldResponse{
			IsSuccess: false,
			ErrorCode: "🔒 [gRPC PROCESSING]: Сумма холдирования должна быть строго больше нуля",
		}, nil
	}

	// Оборачиваем сырые сетевые копейки в наш атомарный тип Fixed-Point
	amountMoney := fixedpoint.NewMoneyFromInt64(req.AmountUnits)

	// Запускаем транзакционный контур авторизации
	txId, currentState, err := h.engine.ExecuteHoldInitiation(
		ctx,
		req.MerchantId, // Protobuf-генерация snake_case tags -> MerchantId
		req.PaymentToken,
		req.IdempotencyKey,
		req.CategoryTag,
		amountMoney,
	)

	if err != nil {
		return &gen.HoldResponse{
			IsSuccess: false,
			ErrorCode: fmt.Sprintf("🔒 [gRPC HOLD TRANSACTION REJECTED]: %v", err),
		}, nil
	}

	return &gen.HoldResponse{
		TransactionId: txId,
		CurrentState:  currentState,
		IsSuccess:     true,
	}, nil
}

// ProcessCapture принимает gRPC-вызов на окончательное списание денег с вызовом Ledger книги
func (h *ProcessingGrpcHandler) ProcessCapture(ctx context.Context, req *gen.CaptureRequest) (*gen.CaptureResponse, error) {
	nextState, err := h.engine.ExecuteCaptureConfirmation(ctx, req.TransactionId)
	if err != nil {
		return &gen.CaptureResponse{
			IsSuccess:    false,
			ErrorCode:    fmt.Sprintf("🔒 [gRPC CAPTURE FAILURE]: %v", err),
			CurrentState: nextState,
		}, nil
	}

	return &gen.CaptureResponse{
		CurrentState: nextState,
		IsSuccess:    true,
	}, nil
}

// ProcessReversal принимает gRPC-вызов отмены транзакции и возврата овердрафтов
func (h *ProcessingGrpcHandler) ProcessReversal(ctx context.Context, req *gen.ReversalRequest) (*gen.ReversalResponse, error) {
	nextState, err := h.engine.ExecuteReversalCancellation(ctx, req.TransactionId)
	if err != nil {
		return &gen.ReversalResponse{
			IsSuccess:    false,
			CurrentState: nextState,
		}, nil
	}

	return &gen.ReversalResponse{
		CurrentState: nextState,
		IsSuccess:    true,
	}, nil
}
