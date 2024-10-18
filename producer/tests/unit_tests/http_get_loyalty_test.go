package unit_tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlexBlackNn/authloyalty/loyalty/app/serverhttp"
	"github.com/AlexBlackNn/authloyalty/loyalty/cmd/router"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/dto"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/logger"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/services/loyaltyservice"
	"github.com/AlexBlackNn/authloyalty/loyalty/tests/unit_tests/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

type LoyaltySuite struct {
	suite.Suite
	application *serverhttp.App
	client      http.Client
	srv         *httptest.Server
}

func (ls *LoyaltySuite) SetupSuite() {
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

func (ls *LoyaltySuite) BeforeTest(suiteName, testName string) {
	// Starts server with first random port.
	ls.srv = httptest.NewServer(router.NewChiRouter(
		ls.application.Cfg,
		ls.application.Log,
		ls.application.HandlersV1,
		ls.application.HealthChecker,
	))
}

func (ls *LoyaltySuite) AfterTest(suiteName, testName string) {
	ls.srv = nil
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(LoyaltySuite))
}

func (ls *LoyaltySuite) TestHttpServerRegisterHappyPath() {
	type Want struct {
		code        int
		response    dto.UserLoyalty
		contentType string
	}

	test := struct {
		name string
		url  string
		body []byte
		want Want
	}{
		name: "get userLoyalty",
		url:  "/loyalty/79d3ac44-5857-4185-ba92-1a224fbacb51",
		want: Want{
			code:        http.StatusOK,
			contentType: "application/json",
			response: dto.UserLoyalty{
				UUID:    "79d3ac44-5857-4185-ba92-1a224fbacb51",
				Balance: 1000,
			},
		},
	}
	// stop server when tests finished
	defer ls.srv.Close()

	ls.Run(test.name, func() {
		url := ls.srv.URL + test.url
		request, err := http.NewRequest(http.MethodGet, url, nil)
		ls.NoError(err)
		res, err := ls.client.Do(request)
		ls.NoError(err)
		ls.Equal(test.want.code, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		ls.NoError(err)

		var response dto.UserLoyalty
		err = json.Unmarshal(body, &response)
		ls.NoError(err)
		ls.Equal(test.want.response.Balance, response.Balance)
		ls.Equal(test.want.response.UUID, response.UUID)
	})
}
