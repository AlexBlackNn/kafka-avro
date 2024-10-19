package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/AlexBlackNn/kafka-avro/example-transaction/app"
)

func main() {

	application, err := app.Fabric()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("application starts with cfg %s \n", application.GetConfig())
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	application.Start(ctx)
}
