package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/siderustler/go-ecommerce/basket"
	"github.com/siderustler/go-ecommerce/customer"
	"github.com/siderustler/go-ecommerce/ports"
	"github.com/siderustler/go-ecommerce/product"
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
	basketServices := basket.NewServices()
	httpServer := ports.NewHttpServer(customerServices, productServices, basketServices)
	err = httpServer.Run(context.TODO(), ":8080")
	if err != nil {
		log.Fatal(err)
	}
}
