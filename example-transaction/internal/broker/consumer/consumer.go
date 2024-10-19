package broker

import (
	"context"
	"log/slog"

	"github.com/AlexBlackNn/kafka-avro/example-transaction/internal/config"
	"github.com/AlexBlackNn/kafka-avro/example-transaction/internal/dto"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avro"
	"go.opentelemetry.io/otel/propagation"
)

type Message struct {
	UUID    string
	Balance int
	Type    string
	Comment string
}

type MessageReceived struct {
	Msg Message
	Ctx context.Context
	Err error
}

type Broker struct {
	consumer     *kafka.Consumer
	deserializer serde.Deserializer
	log          *slog.Logger
}

// New returns kafka consumer with schema registry
func New(cfg *config.Config, log *slog.Logger) (*Broker, error) {
	confluentConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  cfg.Kafka.KafkaURL,
		"group.id":           "1",
		"session.timeout.ms": 6000,
		"auto.offset.reset":  "earliest"})
	if err != nil {
		return nil, err
	}

	client, err := schemaregistry.NewClient(schemaregistry.NewConfig(cfg.Kafka.SchemaRegistryURL))
	if err != nil {
		return nil, err
	}

	deser, err := avro.NewSpecificDeserializer(client, serde.ValueSerde, avro.NewDeserializerConfig())
	if err != nil {
		return nil, err
	}

	// TODO topic get from config
	err = confluentConsumer.Subscribe("users", nil)

	broker := &Broker{
		consumer:     confluentConsumer,
		deserializer: deser,
		log:          log,
	}
	return broker, nil
}

// Close closes deserialization agent and kafka consumer
// WARNING: Consume method need to be finished before.
// https://github.com/confluentinc/confluent-kafka-go/issues/136#issuecomment-586166364
func (b *Broker) Close() error {
	b.deserializer.Close()
	//https://docs.confluent.io/platform/current/clients/confluent-kafka-go/index.html#hdr-High_level_Consumer
	err := b.consumer.Close()
	if err != nil {
		return err
	}
	return nil
}

func (b *Broker) Consume() error {
	//TODO: get timeout from config
	ev := b.consumer.Poll(100)
	if ev == nil {
		return nil
	}

	switch e := ev.(type) {
	case *kafka.Message:
		var msg dto.User

		err := b.deserializer.DeserializeInto(*e.TopicPartition.Topic, e.Value, &msg)
		if err != nil {
			b.log.Error(
				"Failed to deserialize payload",
				"err", err.Error(),
			)
			return err
		} else {
			b.log.Info(
				"Message received",
				"topic", e.TopicPartition, "message", msg,
			)
		}

		if e.Headers != nil {
			headers := propagation.MapCarrier{}

			for _, recordHeader := range e.Headers {
				headers[recordHeader.Key] = string(recordHeader.Value)
			}
		}

	case kafka.Error:
		// Errors should generally be considered
		// informational, the client will try to
		// automatically recover.
		b.log.Error("kafka.Error", "code", e.Code(), "err", e.Error())
	default:
		b.log.Warn("Event:", "msg", e.String())
	}
	return nil
}
