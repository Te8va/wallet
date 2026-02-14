package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Te8va/wallet/internal/domain"
	appErrors "github.com/Te8va/wallet/internal/errors"
)

type WalletRepository struct {
	db *pgxpool.Pool
}

func NewWalletRepository(db *pgxpool.Pool) (*WalletRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	return &WalletRepository{db: db}, nil
}

func (r *WalletRepository) ProcessTransaction(ctx context.Context, walletID string, opType domain.OperationType, amount int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var balance int64
	err = tx.QueryRow(ctx,
		`SELECT balance FROM wallet WHERE id = $1 FOR UPDATE`,
		walletID,
	).Scan(&balance)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = tx.Exec(ctx,
				`INSERT INTO wallet (id, balance) VALUES ($1, $2)`,
				walletID, 0,
			)
			if err != nil {
				return fmt.Errorf("failed to create wallet: %w", err)
			}
			balance = 0
		} else {
			return fmt.Errorf("failed to get wallet: %w", err)
		}
	}

	if opType == domain.WITHDRAW && balance < amount {
		return appErrors.ErrInsufficientFunds
	}

	var delta int64
	if opType == domain.DEPOSIT {
		delta = amount
	} else {
		delta = -amount
	}

	tag, err := tx.Exec(ctx,
		`UPDATE wallet SET balance = balance + $1 WHERE id = $2`,
		delta, walletID,
	)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return appErrors.ErrWalletNotFound
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *WalletRepository) GetBalance(ctx context.Context, walletID string) (int64, error) {
	query := `SELECT balance FROM wallet WHERE id = $1`

	var balance int64
	err := r.db.QueryRow(ctx, query, walletID).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, appErrors.ErrWalletNotFound
		}
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}
