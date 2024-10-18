package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/AlexBlackNn/kafka-avro/producer/internal/broker"
	"github.com/AlexBlackNn/kafka-avro/producer/internal/config"
	"github.com/AlexBlackNn/kafka-avro/producer/internal/dto"
	"github.com/AlexBlackNn/kafka-avro/producer/internal/logger"
)

type sendCloser interface {
	Send(msg dto.User, topic string, key string) error
	Close()
}

type App struct {
	ServerProducer sendCloser
	log            *slog.Logger
}

func New() (*App, error) {

	cfg := config.New()
	log := logger.New(cfg.Env)

	producer, err := broker.New(cfg, log)
	if err != nil {
		return nil, err
	}

	return &App{
		ServerProducer: producer,
		log:            log,
	}, nil
}

func (a *App) Start(ctx context.Context) {
	a.log.Info("producer starts")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			value := dto.User{
				Name:            "First user",
				Favorite_number: 42,
				Favorite_color:  "blue",
			}

			a.ServerProducer.Send(value, "users", "53")
			time.Sleep(time.Second)
		}
	}
}

func (a *App) Stop() {
	a.log.Info("close kafka client")
	a.ServerProducer.Close()
}
