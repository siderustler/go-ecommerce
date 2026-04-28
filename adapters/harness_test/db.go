package db_harness

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/siderustler/go-ecommerce/adapters"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func SetupIntegrationTestDatabase(ctx context.Context) (*adapters.Database, *postgres.PostgresContainer, error) {
	schemaPath, err := filepath.Abs(filepath.Join("..", "..", "sql", "01_schema.sql"))
	if err != nil {
		return nil, nil, err
	}
	seedPath, err := filepath.Abs(filepath.Join("..", "..", "sql", "02_seed.sql"))
	if err != nil {
		return nil, nil, err
	}
	container, err := postgres.Run(ctx,
		"postgres:18",
		postgres.WithDatabase("ecomm_integration"),
		postgres.WithUsername("user"),
		postgres.WithPassword("secret"),
		postgres.WithOrderedInitScripts(schemaPath, seedPath),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, nil, err
	}

	connectionString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	db, err := adapters.OpenDB(connectionString)
	if err != nil {
		return nil, nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, nil, err
	}

	return db, container, nil
}

func BeginTestTransaction(ctx context.Context, db *adapters.Database) (teardown func() error, tx *sql.Tx, err error) {
	if err := db.PingContext(ctx); err != nil {
		return nil, nil, fmt.Errorf("ping database: %v", err)
	}

	tx, err = db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		return nil, nil, fmt.Errorf("begin transaction: %v", err)
	}

	teardown = func() error {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			return fmt.Errorf("rollback transaction: %v", err)
		}
		return nil
	}
	return teardown, tx, nil
}
