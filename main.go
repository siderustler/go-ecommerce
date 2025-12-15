package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/siderustler/go-ecommerce/ports"
	"github.com/siderustler/go-ecommerce/services"
	"github.com/stripe/stripe-go/v83"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	stripe.Key = os.Getenv("STRIPE_SERVER_KEY")

	services := services.NewServices()
	httpServer := ports.NewHttpServer(services)
	err = httpServer.Run(context.TODO(), ":8080")
	if err != nil {
		log.Fatal(err)
	}
}
