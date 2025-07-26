package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/Shuba-Buba/usdt-rate-service/internal/domain"
)

type RateRepository struct {
	db *sqlx.DB
}

func NewRateRepository(db *sqlx.DB) *RateRepository {
	return &RateRepository{db: db}
}

func (r *RateRepository) Save(ctx context.Context, rate domain.Rate) error {
	query := `
        INSERT INTO public.rates (timestamp, data)
        VALUES ($1, $2)
    `

	data := struct {
		Asks []domain.OrderBookEntry `json:"asks"`
		Bids []domain.OrderBookEntry `json:"bids"`
	}{
		Asks: rate.Asks,
		Bids: rate.Bids,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal order book data: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, rate.Timestamp, dataJSON)
	if err != nil {
		return fmt.Errorf("failed to save rate: %w", err)
	}

	return nil
}

func (r *RateRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
