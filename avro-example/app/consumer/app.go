package consumer

import (
	"context"
	"log"
	"log/slog"
	"time"

	consumer "github.com/AlexBlackNn/kafka-avro/avro-example/internal/broker/consumer"
	"github.com/AlexBlackNn/kafka-avro/avro-example/internal/config"
)

type consumeCloser interface {
	Consume() error
	Close() error
}

type App struct {
	ServerConsumer consumeCloser
	log            *slog.Logger
	Cfg            *config.Config
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {

	cons, err := consumer.New(cfg, log)
	if err != nil {
		return nil, err
	}

	return &App{
		ServerConsumer: cons,
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
			err := a.ServerConsumer.Consume()
			if err != nil {
				log.Fatal(err.Error())
			}
			time.Sleep(time.Second)
		}
	}
}

func (a *App) Stop() {
	a.log.Info("close kafka client")
	err := a.ServerConsumer.Close()
	if err != nil {
		a.log.Error(err.Error())
	}
}

func (a *App) GetConfig() string {
	return a.Cfg.String()
}
