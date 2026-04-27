package harness

import (
	"context"
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
