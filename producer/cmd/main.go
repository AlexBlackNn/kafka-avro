package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/AlexBlackNn/kafka-avro/producer/app"
)

func main() {

	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("application starts with cfg: ", application.Cfg)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	application.Start(ctx)
}
