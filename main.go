package main

import (
	"context"
	"database/sql"
	"encoding/gob"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/siderustler/go-ecommerce/customer"
	customerRepository "github.com/siderustler/go-ecommerce/customer/repository"
	"github.com/siderustler/go-ecommerce/ports"
	"github.com/siderustler/go-ecommerce/product"
	"github.com/siderustler/go-ecommerce/product/repository"
	"github.com/siderustler/go-ecommerce/store"
	store_repository "github.com/siderustler/go-ecommerce/store/repository"
	"github.com/stripe/stripe-go/v84"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	gob.Register(map[string]interface{}{})
	stripe.Key = os.Getenv("STRIPE_SERVER_KEY")

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URI"))
	if err != nil {
		panic(fmt.Errorf("connecting to db: %w", err))
	}
	productsRepo, err := repository.NewRepository(context.Background(), db)
	if err != nil {
		panic(fmt.Errorf("creating products repo: %w", err))
	}
	productServices := product.NewServices(productsRepo)

	customerRepo, err := customerRepository.NewRepository(context.Background(), db)
	if err != nil {
		panic(fmt.Errorf("creating customer repo: %w", err))
	}
	customerServices := customer.NewServices(customerRepo)

	storeRepo, err := store_repository.NewRepository(context.Background(), db)
	if err != nil {
		panic(fmt.Errorf("creating store repo: %w", err))
	}
	storeServices := store.NewServices(storeRepo)

	httpServer := ports.NewHttpServer(customerServices, productServices, storeServices)
	err = httpServer.Run(context.TODO(), ":8080")
	if err != nil {
		log.Fatal(err)
	}
}
