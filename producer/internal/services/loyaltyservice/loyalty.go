package loyaltyservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/storage"
	"github.com/AlexBlackNn/authloyalty/loyalty/pkg/tracing"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/loyalty/pkg/broker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type loyaltyBroker interface {
	GetMessageChan() chan *broker.MessageReceived
}

type loyaltyStorage interface {
	AddLoyalty(
		ctx context.Context,
		loyalty *domain.UserLoyalty,
	) (context.Context, *domain.UserLoyalty, error)
	GetLoyalty(
		ctx context.Context,
		loyalty *domain.UserLoyalty,
	) (context.Context, *domain.UserLoyalty, error)
	HealthCheck(context.Context) (context.Context, error)
	Stop() error
}

type Loyalty struct {
	cfg          *config.Config
	log          *slog.Logger
	loyalBroker  loyaltyBroker
	loyalStorage loyaltyStorage
}

var tracer = otel.Tracer("loyalty service")

// New returns a new instance of Auth service
func New(
	cfg *config.Config,
	log *slog.Logger,
	loyalBroker loyaltyBroker,
	loyalStorage loyaltyStorage,
) *Loyalty {

	msgChan := loyalBroker.GetMessageChan()
	go func() {
		for msg := range msgChan {

			userLoyalty := &domain.UserLoyalty{UUID: msg.Msg.UUID, Balance: msg.Msg.Balance, Operation: "registration"}
			ctx, span := tracer.Start(msg.Ctx, "service layer: GetMessageChan",
				trace.WithAttributes(attribute.String("handler", "GetMessageChan")))
			ctx, userLoyalty, err := loyalStorage.AddLoyalty(ctx, userLoyalty)
			if err != nil {
				log.Error(err.Error(), "userLoyalty", userLoyalty)
				tracing.SpanError(span, "failed to create loyalty for user", err)
				continue
			}
			log.Info("GetMessageChan: userLoyalty", userLoyalty)
			span.AddEvent(
				"user loyalty extracted from broker message",
				trace.WithAttributes(
					attribute.String("user-id", userLoyalty.UUID),
					attribute.Int("user-id", userLoyalty.Balance),
				),
			)
			span.End()
		}
	}()

	return &Loyalty{
		cfg:          cfg,
		log:          log,
		loyalStorage: loyalStorage,
	}
}

// HealthCheck returns service health check.
func (l *Loyalty) HealthCheck(ctx context.Context) (context.Context, error) {
	log := l.log.With(
		slog.String("info", "SERVICE LAYER: HealthCheck"),
	)
	log.Info("starts getting health check")
	defer log.Info("finish getting health check")
	return l.loyalStorage.HealthCheck(ctx)
}

func (l *Loyalty) GetLoyalty(
	ctx context.Context,
	userLoyalty *domain.UserLoyalty,
) (context.Context, *domain.UserLoyalty, error) {
	const op = "SERVICE LAYER: GetLoyalty"
	ctx, span := tracer.Start(ctx, "service layer: GetLoyalty",
		trace.WithAttributes(attribute.String("handler", "GetLoyalty")))
	defer span.End()

	log := l.log.With(
		slog.String("trace-id", "trace-id"),
		slog.String("user-id", "user-id"),
	)
	log.Info("getting loyalty for user")

	ctx, userLoyalty, err := l.loyalStorage.GetLoyalty(ctx, userLoyalty)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return ctx, nil, ErrUserNotFound
		}
		tracing.SpanError(span, "failed to get loyalty", err)
		log.Error("failed to get loyalty", "err", err.Error())
		return ctx, nil, fmt.Errorf("%s: %w", op, err)
	}
	span.AddEvent(
		"user loyalty extracted",
		trace.WithAttributes(
			attribute.String("user-id", userLoyalty.UUID),
			attribute.Int("user-id", userLoyalty.Balance),
		))
	return ctx, userLoyalty, nil
}

func (l *Loyalty) AddLoyalty(
	ctx context.Context,
	userLoyalty *domain.UserLoyalty,
) (context.Context, *domain.UserLoyalty, error) {
	const op = "SERVICE LAYER: AddLoyalty"
	ctx, span := tracer.Start(ctx, "service layer: AddLoyalty",
		trace.WithAttributes(attribute.String("handler", "AddLoyalty")))
	defer span.End()

	log := l.log.With(
		slog.String("trace-id", "trace-id"),
		slog.String("user-id", "user-id"),
	)
	log.Info("add loyalty to user")

	ctx, userLoyalty, err := l.loyalStorage.AddLoyalty(ctx, userLoyalty)
	if err != nil {
		if errors.Is(err, storage.ErrNegativeBalance) {
			tracing.SpanError(span, "withdraw might lead to negative balance", err)
			log.Error("withdraw might lead to negative balance", "err", err.Error())
			return ctx, nil, ErrNegativeBalance
		}
		if errors.Is(err, storage.ErrUserNotFound) {
			tracing.SpanError(span, "withdraw might lead to negative balance", err)
			log.Error("withdraw might lead to negative balance", "err", err.Error())
			return ctx, nil, ErrUserNotFound
		}
		tracing.SpanError(span, "failed to get loyalty", err)
		log.Error("failed to get loyalty", "err", err.Error())
		return ctx, nil, fmt.Errorf("%s: %w", op, err)
	}
	span.AddEvent(
		"user loyalty extracted",
		trace.WithAttributes(
			attribute.String("user-id", userLoyalty.UUID),
			attribute.Int("user-id", userLoyalty.Balance),
		))
	return ctx, userLoyalty, nil
}
