package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Te8va/wallet/internal/domain"
	appErrors "github.com/Te8va/wallet/internal/errors"
	"github.com/Te8va/wallet/internal/handler/mocks"
)

func setupTestHandler(t *testing.T) (*gomock.Controller, *mocks.MockWallet, *WalletHandler) {
	ctrl := gomock.NewController(t)
	mockWallet := mocks.NewMockWallet(ctrl)
	handler := NewWalletHandler(mockWallet)
	return ctrl, mockWallet, handler
}

func TestWalletOperationHandler(t *testing.T) {
	ctrl, mockWallet, handler := setupTestHandler(t)
	defer ctrl.Finish()

	testCases := []struct {
		name        string
		contentType string
		body        interface{}
		mockServ    func()
		wantCode    int
		mockErr     string
	}{
		{
			name:        "successful deposit",
			contentType: "application/json",
			body: domain.WalletRequest{
				WalletID:      "123e4567-e89b-12d3-a456-426614174000",
				OperationType: domain.DEPOSIT,
				Amount:        1000,
			},
			mockServ: func() {
				mockWallet.EXPECT().ProcessTransaction(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000", domain.DEPOSIT, int64(1000)).Return(nil)
			},
			wantCode: http.StatusOK,
			mockErr:  "",
		},
		{
			name:        "successful withdraw",
			contentType: "application/json",
			body: domain.WalletRequest{
				WalletID:      "123e4567-e89b-12d3-a456-426614174000",
				OperationType: domain.WITHDRAW,
				Amount:        500,
			},
			mockServ: func() {
				mockWallet.EXPECT().ProcessTransaction(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000", domain.WITHDRAW, int64(500)).Return(nil)
			},
			wantCode: http.StatusOK,
			mockErr:  "",
		},
		{
			name:        "invalid content type",
			contentType: "text/plain",
			body:        "not json",
			mockServ:    func() {},
			wantCode:    http.StatusBadRequest,
			mockErr:     `{"error":"Invalid request body"}`,
		},
		{
			name:        "missing wallet ID",
			contentType: "application/json",
			body: domain.WalletRequest{
				WalletID:      "",
				OperationType: domain.DEPOSIT,
				Amount:        1000,
			},
			mockServ: func() {},
			wantCode: http.StatusBadRequest,
			mockErr:  `{"error":"Wallet ID is required"}`,
		},
		{
			name:        "invalid operation type",
			contentType: "application/json",
			body: map[string]interface{}{
				"valletId":      "123e4567-e89b-12d3-a456-426614174000",
				"operationType": "DEPOSITT",
				"amount":        1000,
			},
			mockServ: func() {},
			wantCode: http.StatusBadRequest,
			mockErr:  `{"error":"Operation type must be DEPOSIT or WITHDRAW"}`,
		},
		{
			name:        "zero amount",
			contentType: "application/json",
			body: domain.WalletRequest{
				WalletID:      "123e4567-e89b-12d3-a456-426614174000",
				OperationType: domain.DEPOSIT,
				Amount:        0,
			},
			mockServ: func() {},
			wantCode: http.StatusBadRequest,
			mockErr:  `{"error":"Amount must be more than 0"}`,
		},
		{
			name:        "negative amount",
			contentType: "application/json",
			body: domain.WalletRequest{
				WalletID:      "123e4567-e89b-12d3-a456-426614174000",
				OperationType: domain.DEPOSIT,
				Amount:        -100,
			},
			mockServ: func() {},
			wantCode: http.StatusBadRequest,
			mockErr:  `{"error":"Amount must be more than 0"}`,
		},
		{
			name:        "insufficient funds",
			contentType: "application/json",
			body: domain.WalletRequest{
				WalletID:      "123e4567-e89b-12d3-a456-426614174000",
				OperationType: domain.WITHDRAW,
				Amount:        5000,
			},
			mockServ: func() {
				mockWallet.EXPECT().ProcessTransaction(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000", domain.WITHDRAW, int64(5000)).Return(appErrors.ErrInsufficientFunds)
			},
			wantCode: http.StatusBadRequest,
			mockErr:  `{"error":"insufficient funds"}`,
		},
		{
			name:        "internal server error",
			contentType: "application/json",
			body: domain.WalletRequest{
				WalletID:      "123e4567-e89b-12d3-a456-426614174000",
				OperationType: domain.DEPOSIT,
				Amount:        1000,
			},
			mockServ: func() {
				mockWallet.EXPECT().ProcessTransaction(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000", domain.DEPOSIT, int64(1000)).Return(appErrors.ErrWalletNotFound)
			},
			wantCode: http.StatusInternalServerError,
			mockErr:  `{"error":"Internal server error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var bodyBytes []byte
			switch v := tc.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", tc.contentType)

			tc.mockServ()

			w := httptest.NewRecorder()
			handler.WalletOperationHandler(w, req)

			require.Equal(t, tc.wantCode, w.Code)
			if tc.mockErr != "" {
				require.JSONEq(t, tc.mockErr, w.Body.String())
			}
		})
	}
}

func TestGetBalanceHandler(t *testing.T) {
	ctrl, mockWallet, handler := setupTestHandler(t)
	defer ctrl.Finish()

	testCases := []struct {
		name     string
		walletID string
		mockServ func()
		wantCode int
		mockErr  string
	}{
		{
			name:     "successful",
			walletID: "123e4567-e89b-12d3-a456-426614174000",
			mockServ: func() {
				mockWallet.EXPECT().GetBalance(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000").Return(int64(1500), nil)
			},
			wantCode: http.StatusOK,
			mockErr:  `{"wallet_id":"123e4567-e89b-12d3-a456-426614174000","balance":1500}`,
		},
		{
			name:     "wallet not found",
			walletID: "123e4567-e89b-12d3-a456-426614174000",
			mockServ: func() {
				mockWallet.EXPECT().GetBalance(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000").Return(int64(0), appErrors.ErrWalletNotFound)
			},
			wantCode: http.StatusNotFound,
			mockErr:  `{"error":"Wallet not found"}`,
		},
		{
			name:     "empty wallet ID",
			walletID: "",
			mockServ: func() {},
			wantCode: http.StatusBadRequest,
			mockErr:  `{"error":"Wallet ID is required"}`,
		},
		{
			name:     "internal server error",
			walletID: "123e4567-e89b-12d3-a456-426614174000",
			mockServ: func() {
				mockWallet.EXPECT().GetBalance(gomock.Any(), "123e4567-e89b-12d3-a456-426614174000").Return(int64(0), appErrors.ErrInsufficientFunds)
			},
			wantCode: http.StatusInternalServerError,
			mockErr:  `{"error":"Internal server error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+tc.walletID, nil)

			if tc.walletID != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("walletId", tc.walletID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}

			tc.mockServ()

			w := httptest.NewRecorder()
			handler.GetBalanceHandler(w, req)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.mockErr, w.Body.String())
		})
	}
}
