package harness_test

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/siderustler/go-ecommerce/adapters"
	harness "github.com/siderustler/go-ecommerce/adapters/harness_test"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	integrationDB *adapters.Database
	container     *postgrescontainer.PostgresContainer
)

func TestMain(m *testing.M) {
	var err error
	integrationDB, container, err = harness.SetupIntegrationTestDatabase(context.Background())
	if err != nil {
		log.Printf("setting up harness integration database: %v", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	if integrationDB != nil {
		if err := integrationDB.Close(); err != nil {
			log.Printf("closing integration database: %v", err)
		}
	}
	if container != nil {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("terminating postgres test container: %v", err)
		}
	}

	os.Exit(exitCode)
}

func TestDatabaseConnection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	if err := integrationDB.PingContext(ctx); err != nil {
		t.Fatalf("ping database: %v", err)
	}

	tx, err := integrationDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			t.Fatalf("rollback transaction: %v", err)
		}
	})

	txDB := integrationDB.WithTransactionExecutor(tx)
	customerID := "harness-" + uuid.NewString()
	const (
		name  = "Harness User"
		email = "harness@example.com"
		phone = "+48100100100"
	)

	if _, err := txDB.ExecContext(
		ctx,
		`INSERT INTO customers (customer_id, name, email, phone) VALUES ($1, $2, $3, $4)`,
		customerID,
		name,
		email,
		phone,
	); err != nil {
		t.Fatalf("insert customer: %v", err)
	}

	var gotName, gotEmail, gotPhone string
	if err := txDB.QueryRowContext(
		ctx,
		`SELECT name, email, phone FROM customers WHERE customer_id = $1`,
		customerID,
	).Scan(&gotName, &gotEmail, &gotPhone); err != nil {
		t.Fatalf("query inserted customer: %v", err)
	}

	if gotName != name || gotEmail != email || gotPhone != phone {
		t.Fatalf("unexpected customer values: got (%q, %q, %q)", gotName, gotEmail, gotPhone)
	}
}

func TestPerTestTransaction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tx, err := integrationDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}

	txDB := integrationDB.WithTransactionExecutor(tx)
	customerID := "rollback-" + uuid.NewString()

	if _, err := txDB.ExecContext(
		ctx,
		`INSERT INTO customers (customer_id, name, email, phone) VALUES ($1, $2, $3, $4)`,
		customerID,
		"Rollback User",
		"rollback@example.com",
		"+48100200200",
	); err != nil {
		t.Fatalf("insert customer in transaction: %v", err)
	}

	var seenInTx string
	if err := txDB.QueryRowContext(
		ctx,
		`SELECT customer_id FROM customers WHERE customer_id = $1`,
		customerID,
	).Scan(&seenInTx); err != nil {
		t.Fatalf("query customer inside transaction: %v", err)
	}
	if seenInTx != customerID {
		t.Fatalf("unexpected customer id in transaction: got %q want %q", seenInTx, customerID)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback transaction: %v", err)
	}

	var count int
	if err := integrationDB.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM customers WHERE customer_id = $1`,
		customerID,
	).Scan(&count); err != nil {
		t.Fatalf("query customer count after rollback: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected rolled back customer to be absent, got count %d", count)
	}
}
