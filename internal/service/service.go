package service

import (
	"context"

	"github.com/Te8va/wallet/internal/domain"
)

//go:generate mockgen -source=service.go -destination=mocks/wallet_mock.gen.go -package=mocks
type walletServ interface {
	ProcessTransaction(ctx context.Context, walletID string, opType domain.OperationType, amount int64) error
	GetBalance(ctx context.Context, walletID string) (int64, error)
}

type WalletService struct {
	repo walletServ
}

func NewWalletService(repo walletServ) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) ProcessTransaction(ctx context.Context, walletID string, opType domain.OperationType, amount int64) error {
	return s.repo.ProcessTransaction(ctx, walletID, opType, amount)
}

func (s *WalletService) GetBalance(ctx context.Context, walletID string) (int64, error) {
	return s.repo.GetBalance(ctx, walletID)
}
