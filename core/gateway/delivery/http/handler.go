package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings" // Добавлен импорт пакета строк

	"clearway-fintech-core/internal/pkg/fixedpoint"
)

type GatewayUseCaseInterface interface {
	ExecuteTokenization(ctx context.Context, pan, holder, expiry, cvv string) (string, error)
}

type GatewayHttpHandler struct {
	useCase GatewayUseCaseInterface
}

func NewGatewayHttpHandler(uc GatewayUseCaseInterface) *GatewayHttpHandler {
	return &GatewayHttpHandler{
		useCase: uc,
	}
}

type CreatePaymentTokenRequest struct {
	CardNumber   string `json:"card_number"`
	CardHolder   string `json:"card_holder"`
	ExpiryDate   string `json:"expiry_date"`
	CVV          string `json:"cvv"`
	AmountString string `json:"amount"`
}

type TokenResponse struct {
	Success      bool   `json:"success"`
	Token        string `json:"token,omitempty"`
	ErrorMessage string `json:"error,omitempty"`
}

func (h *GatewayHttpHandler) HandleTokenizeRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(TokenResponse{Success: false, ErrorMessage: "Method Not Allowed"})
		return
	}

	var req CreatePaymentTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TokenResponse{Success: false, ErrorMessage: "Bad Request JSON"})
		return
	}

	// ИСПРАВЛЕНО (Ультимативное очищение входящего PAN): Вырезаем любые пробелы прямо на входе!
	req.CardNumber = strings.ReplaceAll(req.CardNumber, " ", "")

	_, err := fixedpoint.NewMoneyFromString(req.AmountString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TokenResponse{Success: false, ErrorMessage: err.Error()})
		return
	}

	// Запускаем безопасную токенизацию карты по стандарту PCI-DSS
	token, err := h.useCase.ExecuteTokenization(r.Context(), req.CardNumber, req.CardHolder, req.ExpiryDate, req.CVV)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(TokenResponse{Success: false, ErrorMessage: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(TokenResponse{Success: true, Token: token})
}
