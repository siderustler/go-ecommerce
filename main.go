package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/siderustler/go-ecommerce/adapters"
	"github.com/siderustler/go-ecommerce/customer"
	customerRepository "github.com/siderustler/go-ecommerce/customer/repository"
	"github.com/siderustler/go-ecommerce/ports"
	"github.com/siderustler/go-ecommerce/product"
	"github.com/siderustler/go-ecommerce/product/repository"
	"github.com/siderustler/go-ecommerce/store"
	store_repository "github.com/siderustler/go-ecommerce/store/repository"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("loading env file", "error", err)
		os.Exit(1)
	}
	logger := adapters.NewLogger()
	slog.SetDefault(logger)

	db, err := adapters.OpenDB(os.Getenv("DATABASE_URI"))
	if err != nil {
		logger.Error("opening database", "error", err)
		os.Exit(1)
	}
	productsRepo := repository.NewRepository(db)
	productServices := product.NewServices(productsRepo)

	customerRepo := customerRepository.NewRepository(db)
	customerServices := customer.NewServices(customerRepo)

	storeRepo := store_repository.NewRepository(db)
	storeServices := store.NewServices(storeRepo)

	httpServer, err := ports.NewHttpServer(customerServices, productServices, storeServices, logger)
	if err != nil {
		logger.Error("creating http server", "error", err)
		os.Exit(1)
	}
	err = httpServer.Run(context.TODO(), ":8080")
	if err != nil {
		logger.Error("running http server", "error", err)
		os.Exit(1)
	}
}
