package producer

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/AlexBlackNn/kafka-avro/example-transaction/internal/broker"
	"github.com/AlexBlackNn/kafka-avro/example-transaction/internal/config"
	"github.com/AlexBlackNn/kafka-avro/example-transaction/internal/dto"
)

type sendCloser interface {
	Send(msg dto.User, topic string, key string) error
	Close()
}

type App struct {
	ServerProducer sendCloser
	log            *slog.Logger
	Cfg            *config.Config
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {

	producer, err := broker.NewProducer(cfg, log)
	if err != nil {
		return nil, err
	}

	return &App{
		ServerProducer: producer,
		log:            log,
		Cfg:            cfg,
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

			err := a.ServerProducer.Send(value, "users", "53")
			if err != nil {
				log.Fatal(err.Error())
			}
			time.Sleep(time.Second)
		}
	}
}

func (a *App) Stop() {
	a.log.Info("close kafka client")
	a.ServerProducer.Close()
}

func (a *App) GetConfig() string {
	return a.Cfg.String()
}
