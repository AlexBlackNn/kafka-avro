package broker

import (
	"context"
	log "log/slog"

	registrationv1 "github.com/AlexBlackNn/authloyalty/commands/proto/registration.v1/registration.v1"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/protobuf"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type Message struct {
	UUID    string
	Balance int
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
	MessageChan  chan *MessageReceived
	workEnable   bool
}

var tracer = otel.Tracer(
	"loyalty service",
	trace.WithInstrumentationVersion(contrib.SemVersion()),
)

// New returns kafka consumer with schema registry
func New(cfg *config.Config) (*Broker, error) {
	confluentConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  cfg.Kafka.KafkaURL,
		"group.id":           "1",
		"session.timeout.ms": 6000,
		"auto.offset.reset":  "earliest"})
	if err != nil {
		return nil, err
	}

	MessageChan := make(chan *MessageReceived)

	client, err := schemaregistry.NewClient(schemaregistry.NewConfig(cfg.Kafka.SchemaRegistryURL))
	if err != nil {
		return nil, err
	}

	deser, err := protobuf.NewDeserializer(client, serde.ValueSerde, protobuf.NewDeserializerConfig())
	if err != nil {
		return nil, err
	}

	deser.ProtoRegistry.RegisterMessage((&registrationv1.RegistrationMessage{}).ProtoReflect().Type())
	//TODO: registration should be got from config
	err = confluentConsumer.Subscribe("registration", nil)

	broker := &Broker{
		consumer:     confluentConsumer,
		deserializer: deser,
		MessageChan:  MessageChan,
	}
	broker.workEnable = true
	go broker.Consume()
	return broker, nil

}

// GetMessageChan returns channel to get messages
func (b *Broker) GetMessageChan() chan *MessageReceived {
	return b.MessageChan
}

// Close closes deserialization agent and kafka consumer
func (b *Broker) Close() error {
	// Consume method need to be finished before.
	//https://github.com/confluentinc/confluent-kafka-go/issues/136#issuecomment-586166364
	b.workEnable = false
	b.deserializer.Close()
	//https://docs.confluent.io/platform/current/clients/confluent-kafka-go/index.html#hdr-High_level_Consumer
	err := b.consumer.Close()
	if err != nil {
		return err
	}
	return nil
}

func (b *Broker) Consume() {
	for b.workEnable {
		//TODO: get from config
		ev := b.consumer.Poll(100)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			ctx := context.Background()
			ctx, span := tracer.Start(ctx, "kafka_message_processing")
			defer span.End()

			value, err := b.deserializer.Deserialize(*e.TopicPartition.Topic, e.Value)
			if err != nil {
				log.Error("Failed to deserialize payload: %s\n", err)
			} else {
				log.Error("%% Message on %s:\n%+v\n", e.TopicPartition, value)
			}
			if e.Headers != nil {
				log.Error("%% Headers: %v\n", e.Headers)

				headers := propagation.MapCarrier{}

				for _, recordHeader := range e.Headers {
					headers[recordHeader.Key] = string(recordHeader.Value)
				}

				propagator := otel.GetTextMapPropagator()
				ctx = propagator.Extract(ctx, headers)

				ctx, span = tracer.Start(
					ctx,
					"tracer consumer1",
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(
						semconv.MessagingDestinationName("registration"),
					),
				)
				defer span.End()
			}
			// TODO: Balance migt be transmitted from sso and extracted from protobuf, Err - take a look at docs to find out.
			b.MessageChan <- &MessageReceived{Msg: Message{UUID: string(e.Key), Balance: 100, Comment: "registration"}, Ctx: ctx, Err: nil}

		case kafka.Error:
			// Errors should generally be considered
			// informational, the client will try to
			// automatically recover.
			log.Error("%% Error: %v: %v\n", e.Code(), e)
		default:
			log.Warn("Ignored %v\n", e)
		}
	}
}
