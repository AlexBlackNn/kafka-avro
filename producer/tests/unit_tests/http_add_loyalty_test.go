package unit_tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlexBlackNn/authloyalty/loyalty/app/serverhttp"
	"github.com/AlexBlackNn/authloyalty/loyalty/cmd/router"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/logger"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/services/loyaltyservice"
	"github.com/AlexBlackNn/authloyalty/loyalty/tests/unit_tests/mocks"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

type LoyaltyAddSuite struct {
	suite.Suite
	application *serverhttp.App
	client      http.Client
	srv         *httptest.Server
}

func (ls *LoyaltyAddSuite) SetupSuite() {
	var err error

	cfg := config.MustLoadByPath("../../config/local.yaml")
	log := logger.New(cfg.Env)

	ctrl := gomock.NewController(ls.T())
	defer ctrl.Finish()

	userLoyalty := &domain.UserLoyalty{
		UUID: "79d3ac44-5857-4185-ba92-1a224fbacb51", Balance: 1000,
	}

	loyaltyStorageMock := mocks.NewMockloyaltyStorage(ctrl)
	loyaltyStorageMock.EXPECT().
		GetLoyalty(gomock.Any(), gomock.Any()).
		Return(context.Background(), userLoyalty, nil).
		AnyTimes()

	brokerMock := mocks.NewMockloyaltyBroker(ctrl)
	brokerMock.EXPECT().
		GetMessageChan().
		Return(nil).
		AnyTimes()

	loyalService := loyaltyservice.New(
		cfg,
		log,
		brokerMock,
		loyaltyStorageMock,
	)

	// http server
	ls.application, err = serverhttp.New(cfg, log, loyalService)
	ls.Suite.NoError(err)
	ls.client = http.Client{Timeout: 3 * time.Second}
}

func (ls *LoyaltyAddSuite) BeforeTest(suiteName, testName string) {
	// Starts server with first random port.
	ls.srv = httptest.NewServer(router.NewChiRouter(
		ls.application.Cfg,
		ls.application.Log,
		ls.application.HandlersV1,
		ls.application.HealthChecker,
	))
}

func (ls *LoyaltyAddSuite) AfterTest(suiteName, testName string) {
	ls.srv = nil
}

func TestAddSuite(t *testing.T) {
	suite.Run(t, new(LoyaltyAddSuite))
}

func (ls *LoyaltyAddSuite) TestHttpServerRegisterHappyPath() {
	ls.Run("registration", func() {
		//////
		mockCluster, err := kafka.NewMockCluster(1)
		ls.NoError(err)
		defer mockCluster.Close()

		broker := mockCluster.BootstrapServers()
		producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
		ls.NoError(err)

		topic := "registration"

		payload := []byte{0, 0, 0, 0, 1, 0, 10, 13, 116, 101, 115, 116, 64, 116, 101, 115, 116, 46, 99, 111, 109, 18, 6, 115, 116, 114, 105, 110, 103}

		headers := []kafka.Header{{Key: "request-Id", Value: []byte("header values are binary")}}

		err = producer.Produce(&kafka.Message{
			Key:            []byte("79d3ac44-5857-4185-ba92-1a224fbacb51"),
			TopicPartition: kafka.TopicPartition{Topic: &topic},
			Value:          payload,
			Headers:        headers,
		}, nil)
		ls.NoError(err)

		e := <-producer.Events()
		m := e.(*kafka.Message)

		ls.NoError(m.TopicPartition.Error)

	})
}
