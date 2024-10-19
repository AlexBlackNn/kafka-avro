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
	log.Printf("application starts %s with cfg %s \n", application.Cfg.Kafka.Type, application.Cfg)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	application.Start(ctx)
}
