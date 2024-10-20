package producer

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/AlexBlackNn/kafka-avro/example-transaction/internal/broker/producer"
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

	prod, err := producer.New(cfg, log)
	if err != nil {
		return nil, err
	}

	return &App{
		ServerProducer: prod,
		log:            log,
		Cfg:            cfg,
	}, nil
}

func (a *App) Start(ctx context.Context) {

	var (
		name           string
		favoriteNumber int64
		favoriteColor  string
	)

	a.log.Info("producer starts")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// TODO: move this logic to another method
			fmt.Print("Enter name: ")
			_, err := fmt.Scanln(&name)
			if err != nil {
				a.log.Error(err.Error())
				continue
			}

			fmt.Print("Enter favorite number: ")
			_, err = fmt.Scanln(&favoriteNumber)
			if err != nil {
				a.log.Error(err.Error())
				continue
			}

			fmt.Print("Enter favorite color: ")
			_, err = fmt.Scanln(&favoriteColor)
			if err != nil {
				a.log.Error(err.Error())
				continue
			}

			value := dto.User{
				Name:            name,
				Favorite_number: favoriteNumber,
				Favorite_color:  favoriteColor,
			}

			err = a.ServerProducer.Send(value, "users", "53")
			if err != nil {
				log.Fatal(err.Error())
			}
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
