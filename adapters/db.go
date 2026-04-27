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
	Executor
	db      *sql.DB
	RunInTx func(ctx context.Context, opts *sql.TxOptions, txFunc func(tx *sql.Tx) error) error
}

func (d Database) WithTransactionExecutor(tx *sql.Tx) *Database {
	txDb := &Database{
		Executor: tx,
		db:       d.db,
	}
	txDb.RunInTx = func(ctx context.Context, opts *sql.TxOptions, txFunc func(tx *sql.Tx) error) error {
		return txFunc(tx)
	}
	return txDb
}

func OpenDB(uri string) (*Database, error) {
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, fmt.Errorf("connecting to db: %w", err)
	}
	database := &Database{Executor: db, db: db}
	database.RunInTx = database.defaultRunInTx
	return database, nil
}

func (d *Database) defaultRunInTx(ctx context.Context, opts *sql.TxOptions, txFunc func(tx *sql.Tx) error) (err error) {
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	return txFunc(tx)
}

func (d *Database) PingContext(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *Database) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, opts)
}

func (d *Database) Close() error {
	return d.db.Close()
}
