package producer

import (
	"context"
	"errors"
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
	a.log.Info("producer starts")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			value, err := a.readInput()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = a.ServerProducer.Send(*value, a.Cfg.Kafka.Topic, "53")
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

func (a *App) readInput() (*dto.User, error) {
	// TODO: it would be better to have data validation here

	var (
		name           string
		favoriteNumber int64
		favoriteColor  string
		command        string
	)

	fmt.Print("Command: ")
	_, err := fmt.Scanln(&command)
	if err != nil {
		a.log.Error(err.Error())
		return nil, errors.New("reading command failed")
	}

	if command == "exit" {
		return nil, errors.New("terminate")
	}

	fmt.Print("Enter name: ")
	_, err = fmt.Scanln(&name)
	if err != nil {
		a.log.Error(err.Error())
		return nil, errors.New("reading name failed")
	}

	fmt.Print("Enter favorite number: ")
	_, err = fmt.Scanln(&favoriteNumber)
	if err != nil {
		a.log.Error(err.Error())
		return nil, errors.New("reading favorite number failed")
	}

	fmt.Print("Enter favorite color: ")
	_, err = fmt.Scanln(&favoriteColor)
	if err != nil {
		return nil, errors.New("reading favorite color failed")
	}

	value := &dto.User{
		Name:            name,
		Favorite_number: favoriteNumber,
		Favorite_color:  favoriteColor,
	}

	return value, nil
}
