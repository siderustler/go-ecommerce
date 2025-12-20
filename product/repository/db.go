package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type repository struct {
	db *sql.DB
}

// FIXME
func NewRepository(ctx context.Context, db *sql.DB) (*repository, error) {
	_, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating baskets table: %w", err)
	}
	return &repository{db: db}, nil
}

func RunInTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions, txFunc func(tx *sql.Tx) error) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}

	return txFunc(tx)
}
