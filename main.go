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
	productsRepo := repository.NewRepository(db)
	productServices := product.NewServices(productsRepo)

	customerRepo := customerRepository.NewRepository(db)
	customerServices := customer.NewServices(customerRepo)

	storeRepo := store_repository.NewRepository(db)
	storeServices := store.NewServices(storeRepo)

	httpServer := ports.NewHttpServer(customerServices, productServices, storeServices)
	err = httpServer.Run(context.TODO(), ":8080")
	if err != nil {
		log.Fatal(err)
	}
}
