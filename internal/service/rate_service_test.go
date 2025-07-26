package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Shuba-Buba/usdt-rate-service/internal/domain"
	"github.com/Shuba-Buba/usdt-rate-service/internal/service"
	"github.com/Shuba-Buba/usdt-rate-service/pkg/grinex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, rate domain.Rate) error {
	args := m.Called(ctx, rate)
	return args.Error(0)
}

func (m *MockRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockGrinexClient struct {
	mock.Mock
}

func (m *MockGrinexClient) GetDepth(ctx context.Context) (*grinex.DepthResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(*grinex.DepthResponse), args.Error(1)
}

func TestRateService_GetRates_Success(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := new(MockRepository)
	client := new(MockGrinexClient)

	expectedDepth := &grinex.DepthResponse{
		Timestamp: time.Now().Unix(),
		Asks: []grinex.OrderBookEntry{
			{Price: "100.0", Volume: "10.0", Amount: "1000.0", Factor: "0.1", Type: "limit"},
		},
		Bids: []grinex.OrderBookEntry{
			{Price: "99.0", Volume: "5.0", Amount: "495.0", Factor: "0.1", Type: "limit"},
		},
	}
	client.On("GetDepth", ctx).Return(expectedDepth, nil)

	repo.On("Save", ctx, mock.AnythingOfType("domain.Rate")).Return(nil)

	rateService := service.NewRateService(repo, client, logger)

	rate, err := rateService.GetRates(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, rate)
	assert.Equal(t, 1, len(rate.Asks))
	assert.Equal(t, 1, len(rate.Bids))
	assert.Equal(t, "100.0", rate.Asks[0].Price)
	assert.Equal(t, "99.0", rate.Bids[0].Price)

	client.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestRateService_GetRates_ClientError(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := new(MockRepository)
	client := new(MockGrinexClient)

	expectedErr := errors.New("API error")
	client.On("GetDepth", ctx).Return((*grinex.DepthResponse)(nil), expectedErr)

	rateService := service.NewRateService(repo, client, logger)

	rate, err := rateService.GetRates(ctx)

	assert.Nil(t, rate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get depth")

	client.AssertExpectations(t)
	repo.AssertNotCalled(t, "Save")
}

func TestRateService_GetRates_EmptyData(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := new(MockRepository)
	client := new(MockGrinexClient)

	expectedDepth := &grinex.DepthResponse{
		Timestamp: time.Now().Unix(),
		Asks:      []grinex.OrderBookEntry{},
		Bids:      []grinex.OrderBookEntry{},
	}
	client.On("GetDepth", ctx).Return(expectedDepth, nil)

	rateService := service.NewRateService(repo, client, logger)

	rate, err := rateService.GetRates(ctx)

	assert.Nil(t, rate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty depth data")

	client.AssertExpectations(t)
	repo.AssertNotCalled(t, "Save")
}

func TestRateService_GetRates_SaveError(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := new(MockRepository)
	client := new(MockGrinexClient)

	expectedDepth := &grinex.DepthResponse{
		Timestamp: time.Now().Unix(),
		Asks: []grinex.OrderBookEntry{
			{Price: "100.0", Volume: "10.0", Amount: "1000.0", Factor: "0.1", Type: "limit"},
		},
		Bids: []grinex.OrderBookEntry{
			{Price: "99.0", Volume: "5.0", Amount: "495.0", Factor: "0.1", Type: "limit"},
		},
	}
	client.On("GetDepth", ctx).Return(expectedDepth, nil)

	expectedErr := errors.New("database error")
	repo.On("Save", ctx, mock.AnythingOfType("domain.Rate")).Return(expectedErr)

	rateService := service.NewRateService(repo, client, logger)

	rate, err := rateService.GetRates(ctx)

	assert.Nil(t, rate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save rate")

	client.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestRateService_HealthCheck_Success(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := new(MockRepository)
	client := new(MockGrinexClient)

	repo.On("Ping", ctx).Return(nil)

	rateService := service.NewRateService(repo, client, logger)

	err := rateService.HealthCheck(ctx)

	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestRateService_HealthCheck_Error(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := new(MockRepository)
	client := new(MockGrinexClient)

	expectedErr := errors.New("connection failed")
	repo.On("Ping", ctx).Return(expectedErr)

	rateService := service.NewRateService(repo, client, logger)

	err := rateService.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database health check failed")

	repo.AssertExpectations(t)
}
