package app

import (
	"context"
	"errors"

	"github.com/AlexBlackNn/kafka-avro/avro-example/app/consumer"
	"github.com/AlexBlackNn/kafka-avro/avro-example/app/producer"
	"github.com/AlexBlackNn/kafka-avro/avro-example/internal/config"
	"github.com/AlexBlackNn/kafka-avro/avro-example/internal/logger"
)

var ErrWrongType = errors.New("wrong type")

type StartGetConfigStopper interface {
	Start(ctx context.Context)
	GetConfig() string
	Stop()
}

func Fabric() (StartGetConfigStopper, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}
	log := logger.New(cfg.Env)
	if cfg.Kafka.Type == "producer" {
		return producer.New(cfg, log)
	}
	if cfg.Kafka.Type == "consumer" {
		return consumer.New(cfg, log)
	}
	return nil, ErrWrongType
}
