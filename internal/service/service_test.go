package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Te8va/wallet/internal/domain"
	appErrors "github.com/Te8va/wallet/internal/errors"
	"github.com/Te8va/wallet/internal/service"
	"github.com/Te8va/wallet/internal/service/mocks"
)

func TestWalletService_ProcessTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockwalletServ(ctrl)
	svc := service.NewWalletService(mockRepo)

	walletID := "123e4567-e89b-12d3-a456-426614174000"

	testCases := []struct {
		name        string
		opType      domain.OperationType
		amount      int64
		mockRepo    func()
		expectedErr error
	}{
		{
			name:   "successful deposit",
			opType: domain.DEPOSIT,
			amount: 1000,
			mockRepo: func() {
				mockRepo.EXPECT().ProcessTransaction(gomock.Any(), walletID, domain.DEPOSIT, int64(1000)).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "successful withdraw",
			opType: domain.WITHDRAW,
			amount: 500,
			mockRepo: func() {
				mockRepo.EXPECT().ProcessTransaction(gomock.Any(), walletID, domain.WITHDRAW, int64(500)).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "insufficient funds",
			opType: domain.WITHDRAW,
			amount: 5000,
			mockRepo: func() {
				mockRepo.EXPECT().ProcessTransaction(gomock.Any(), walletID, domain.WITHDRAW, int64(5000)).Return(appErrors.ErrInsufficientFunds)
			},
			expectedErr: appErrors.ErrInsufficientFunds,
		},
		{
			name:   "wallet not found",
			opType: domain.DEPOSIT,
			amount: 100,
			mockRepo: func() {
				mockRepo.EXPECT().ProcessTransaction(gomock.Any(), walletID, domain.DEPOSIT, int64(100)).Return(appErrors.ErrWalletNotFound)
			},
			expectedErr: appErrors.ErrWalletNotFound,
		},
		{
			name:   "repository error",
			opType: domain.DEPOSIT,
			amount: 100,
			mockRepo: func() {
				mockRepo.EXPECT().ProcessTransaction(gomock.Any(), walletID, domain.DEPOSIT, int64(100)).Return(errors.New("database connection error"))
			},
			expectedErr: errors.New("database connection error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockRepo()

			err := svc.ProcessTransaction(context.Background(), walletID, tc.opType, tc.amount)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWalletService_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockwalletServ(ctrl)
	svc := service.NewWalletService(mockRepo)

	walletID := "123e4567-e89b-12d3-a456-426614174000"

	testCases := []struct {
		name            string
		mockRepo        func()
		expectedBalance int64
		expectedErr     error
	}{
		{
			name: "success wallet exists",
			mockRepo: func() {
				mockRepo.EXPECT().GetBalance(gomock.Any(), walletID).Return(int64(1500), nil)
			},
			expectedBalance: 1500,
			expectedErr:     nil,
		},
		{
			name: "wallet not found",
			mockRepo: func() {
				mockRepo.EXPECT().GetBalance(gomock.Any(), walletID).Return(int64(0), appErrors.ErrWalletNotFound)
			},
			expectedBalance: 0,
			expectedErr:     appErrors.ErrWalletNotFound,
		},
		{
			name: "repository error",
			mockRepo: func() {
				mockRepo.
					EXPECT().GetBalance(gomock.Any(), walletID).Return(int64(0), errors.New("database error"))
			},
			expectedBalance: 0,
			expectedErr:     errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockRepo()

			balance, err := svc.GetBalance(context.Background(), walletID)

			require.Equal(t, tc.expectedBalance, balance)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
