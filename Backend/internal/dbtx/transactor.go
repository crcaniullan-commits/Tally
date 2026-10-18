package dbtx

import (
	"context"
	"database/sql"
	"fmt"
)

type Transactor struct {
	db *sql.DB
}

func NewTransactor(db *sql.DB) *Transactor {
	return &Transactor{db: db}
}

func (t *Transactor) WithinTxWithResult[T any](ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
	var zero T

	tx, err := t.db.BeginTx(ctx, nil)

	if err != nil {
		return zero, fmt.Errorf("abriendo transacción: %w", err)
	}

	ctx = WithTx(ctx, tx)

	result, err := fn(ctx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return zero, fmt.Errorf("error: %w (rollback también falló: %w)", err, rbErr)
		}
		return zero, err
	}

	if err := tx.Commit(); err != nil {
		return zero, fmt.Errorf("haciendo commit: %w", err)
	}

	return result, nil
}

func (t *Transactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {

	_, err := t.WithinTxWithResult(ctx, func(ctx context.Context) (struct{}, error) {
		err := fn(ctx)
		return struct{}{}, err
	})

	return err
}
