package ssoclient

import (
	"context"

	ssov1 "github.com/AlexBlackNn/authloyalty/commands/proto/sso/gen"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SSOClient struct {
	AuthClient ssov1.AuthClient
}

func New(cfg *config.Config) (*SSOClient, error) {
	grpcClient, err := grpc.NewClient(
		cfg.SSOAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}
	authClient := ssov1.NewAuthClient(grpcClient)
	return &SSOClient{AuthClient: authClient}, nil
}

func (sc *SSOClient) IsJWTValid(ctx context.Context, tracer trace.Tracer, token string) bool {
	ctx, span := tracer.Start(ctx, "sso client: IsJWTValid",
		trace.WithAttributes(attribute.String("operation", "IsJWTValid")))
	defer span.End()

	respIsValid, err := sc.AuthClient.Validate(ctx, &ssov1.ValidateRequest{Token: token})
	if err != nil {
		return false
	}
	return respIsValid.GetSuccess()
}

func (sc *SSOClient) IsAdmin(ctx context.Context, tracer trace.Tracer, uuid string) bool {
	ctx, span := tracer.Start(ctx, "sso client: IsAdmin",
		trace.WithAttributes(attribute.String("operation", "IsAdmin")))
	defer span.End()

	respIsValid, err := sc.AuthClient.IsAdmin(ctx, &ssov1.IsAdminRequest{UserId: uuid})
	if err != nil {
		return false
	}
	return respIsValid.GetIsAdmin()
}
