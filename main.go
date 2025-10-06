package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/siderustler/go-ecommerce/ports"
	"github.com/siderustler/go-ecommerce/services"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	services := services.NewServices()
	httpServer := ports.NewHttpServer(services)
	err = httpServer.Run(context.TODO(), ":8080")
	if err != nil {
		log.Fatal(err)
	}
}
