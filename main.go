package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/siderustler/go-ecommerce/customer"
	"github.com/siderustler/go-ecommerce/ports"
	"github.com/siderustler/go-ecommerce/product"
	"github.com/siderustler/go-ecommerce/store"
	"github.com/siderustler/go-ecommerce/store/repository"
	"github.com/stripe/stripe-go/v83"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	stripe.Key = os.Getenv("STRIPE_SERVER_KEY")

	customerServices := customer.NewServices()
	productServices := product.NewServices()
	storeRepo, err := repository.NewRepository(context.Background(), nil)
	storeServices := store.NewServices(storeRepo)
	httpServer := ports.NewHttpServer(customerServices, productServices, storeServices)
	err = httpServer.Run(context.TODO(), ":8080")
	if err != nil {
		log.Fatal(err)
	}
}
