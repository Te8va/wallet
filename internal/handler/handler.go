package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Te8va/wallet/internal/domain"
	appErrors "github.com/Te8va/wallet/internal/errors"
)

//go:generate mockgen -source=handler.go -destination=mocks/walhandler_mock.gen.go -package=mocks
type Wallet interface {
	ProcessTransaction(ctx context.Context, walletID string, opType domain.OperationType, amount int64) error
	GetBalance(ctx context.Context, walletID string) (int64, error)
}

type WalletHandler struct {
	srv Wallet
}

func NewWalletHandler(srv Wallet) *WalletHandler {
	return &WalletHandler{srv: srv}
}

func (h *WalletHandler) WalletOperationHandler(w http.ResponseWriter, r *http.Request) {

	var req domain.WalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.WalletID == "" {
		sendErrorResponse(w, "Wallet ID is required", http.StatusBadRequest)
		return
	}

	if req.OperationType != domain.DEPOSIT && req.OperationType != domain.WITHDRAW {
		sendErrorResponse(w, "Operation type must be DEPOSIT or WITHDRAW", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		sendErrorResponse(w, "Amount must be more than 0", http.StatusBadRequest)
		return
	}

	err := h.srv.ProcessTransaction(r.Context(), req.WalletID, req.OperationType, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, appErrors.ErrInsufficientFunds):
			sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *WalletHandler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {

	walletID := chi.URLParam(r, "walletId")

	if walletID == "" {
		sendErrorResponse(w, "Wallet ID is required", http.StatusBadRequest)
		return
	}

	balance, err := h.srv.GetBalance(r.Context(), walletID)
	if err != nil {
		if errors.Is(err, appErrors.ErrWalletNotFound) {
			sendErrorResponse(w, "Wallet not found", http.StatusNotFound)
			return
		}
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := domain.BalanceResponse{
		WalletID: walletID,
		Balance:  balance,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(domain.ErrorResponse{Error: message})
}
