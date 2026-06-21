package grpc

import (
	"context"

	"clearway-fintech-core/internal/pkg/fixedpoint"
	"clearway-fintech-core/pb/gen" // Сгенерированные контракты fraud.pb.go
)

// FraudUseCaseInterface задает SOLID границы для UseCase радара
type FraudUseCaseInterface interface {
	EvaluateTransactionRisk(ctx context.Context, clientIP, deviceFingerprint, cardBIN string, amount fixedpoint.Money) (bool, float32, string)
}

type FraudGrpcHandler struct {
	gen.UnimplementedFraudRadarEngineServer
	radar FraudUseCaseInterface
}

func NewFraudGrpcHandler(r FraudUseCaseInterface) *FraudGrpcHandler {
	return &FraudGrpcHandler{
		radar: r,
	}
}

// CheckFraudScore принимает gRPC-запрос на аудит рисков транзакции
// FIXED: Transformed raw inbound message bytes fields straight to fixedpoint domain representations
func (h *FraudGrpcHandler) CheckFraudScore(ctx context.Context, req *gen.FraudCheckRequest) (*gen.FraudCheckResponse, error) {
	// Оборачиваем копейки в наш атомарный финтех-тип
	amountMoney := fixedpoint.NewMoneyFromInt64(req.AmountUnits)

	// Запускаем наносекундную оценку рисков
	isFraud, riskScore, reason := h.radar.EvaluateTransactionRisk(
		ctx,
		req.ClientIp, // Protobuf snake_case "client_ip" -> ClientIp
		req.DeviceFingerprint,
		req.CardBin,
		amountMoney,
	)

	return &gen.FraudCheckResponse{
		IsFraudulent: isFraud,
		RiskScore:    riskScore,
		BlockReason:  reason,
	}, nil
}
