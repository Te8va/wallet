package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgres struct {
	*pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *postgres {
	return &postgres{pool}
}

func GetPgxPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("repository.GetPgxPool: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("repository.GetPgxPool: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.GetPgxPool: %w", err)
	}

	return pool, nil
}

func (p *postgres) WithTransaction(ctx context.Context, txFunc func(pgx.Tx) error) error {
	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("repository.WithTransaction: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository.WithTransaction: %w", err)
	}

	err = txFunc(tx)
	if err != nil {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			resErr := errors.Join(err, rollbackErr)
			return fmt.Errorf("repository.WithTransaction: %w", resErr)
		}

		return fmt.Errorf("repository.WithTransaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("repository.WithTransaction: %w", err)
	}

	return nil
}
