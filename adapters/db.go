package adapters

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Database struct {
	*sql.DB
}

func OpenDB(uri string) (*Database, error) {
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, fmt.Errorf("connecting to db: %w", err)
	}
	return &Database{db}, nil
}

func (d *Database) RunInTx(ctx context.Context, opts *sql.TxOptions, txFunc func(tx *sql.Tx) error) (err error) {
	tx, err := d.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	return txFunc(tx)
}
