package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Shuba-Buba/usdt-rate-service/internal/domain"
	"github.com/Shuba-Buba/usdt-rate-service/pkg/grinex"
)

type RateRepository interface {
	Save(ctx context.Context, rate domain.Rate) error
	Ping(ctx context.Context) error
}

type RateService struct {
	repo   RateRepository
	client *grinex.Client
	logger *zap.Logger
}

func NewRateService(repo RateRepository, client *grinex.Client, logger *zap.Logger) *RateService {
	return &RateService{
		repo:   repo,
		client: client,
		logger: logger,
	}
}

func (s *RateService) GetRates(ctx context.Context) (*domain.Rate, error) {
	s.logger.Info("Fetching USDT rates from Grinex")

	depth, err := s.client.GetDepth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get depth: %w", err)
	}

	if len(depth.Asks) == 0 || len(depth.Bids) == 0 {
		return nil, fmt.Errorf("empty depth data received")
	}

	asks := make([]domain.OrderBookEntry, len(depth.Asks))
	for i, ask := range depth.Asks {
		asks[i] = domain.OrderBookEntry{
			Price:  ask.Price,
			Volume: ask.Volume,
			Amount: ask.Amount,
			Factor: ask.Factor,
			Type:   ask.Type,
		}
	}

	// Маппинг bids из API в доменную модель
	bids := make([]domain.OrderBookEntry, len(depth.Bids))
	for i, bid := range depth.Bids {
		bids[i] = domain.OrderBookEntry{
			Price:  bid.Price,
			Volume: bid.Volume,
			Amount: bid.Amount,
			Factor: bid.Factor,
			Type:   bid.Type,
		}
	}

	timestamp := time.Unix(depth.Timestamp, 0)
	rate := domain.Rate{
		Timestamp: timestamp,
		Asks:      asks,
		Bids:      bids,
	}

	if err := s.repo.Save(ctx, rate); err != nil {
		return nil, fmt.Errorf("failed to save rate: %w", err)
	}

	s.logger.Info("Rate saved successfully")

	return &rate, nil
}

func (s *RateService) HealthCheck(ctx context.Context) error {
	if err := s.repo.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}
