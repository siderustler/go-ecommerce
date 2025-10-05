package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/siderustler/go-ecommerce/ports"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	httpServer := ports.NewHttpServer()
	err = httpServer.Run(context.TODO(), ":8080")
	if err != nil {
		log.Fatal(err)
	}
}
