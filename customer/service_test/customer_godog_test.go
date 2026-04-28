package customer_service_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/siderustler/go-ecommerce/adapters"
	db_harness "github.com/siderustler/go-ecommerce/adapters/harness_test"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	integrationDB *adapters.Database
	container     *postgrescontainer.PostgresContainer
)

func TestMain(m *testing.M) {
	var err error
	integrationDB, container, err = db_harness.SetupIntegrationTestDatabase(context.Background())
	if err != nil {
		log.Printf("setting up customer integration database: %v", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	if integrationDB != nil {
		if err := integrationDB.Close(); err != nil {
			log.Printf("closing customer integration database: %v", err)
		}
	}
	if container != nil {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("terminating customer postgres test container: %v", err)
		}
	}

	os.Exit(exitCode)
}

func TestCustomerServiceFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run customer service feature tests")
	}
}
