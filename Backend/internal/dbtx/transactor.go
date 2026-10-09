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

func (t *Transactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("abriendo transacción: %w", err)
	}

	ctx = WithTx(ctx, tx)

	if err := fn(ctx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error: %w (rollback también falló: %w)", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
