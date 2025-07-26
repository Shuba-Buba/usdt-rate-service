package domain

import "time"

type OrderBookEntry struct {
	Price  string `json:"price"`
	Volume string `json:"volume"`
	Amount string `json:"amount"`
	Factor string `json:"factor"`
	Type   string `json:"type"`
}

type Rate struct {
	ID        int64            `db:"id"`
	Timestamp time.Time        `db:"timestamp"`
	Asks      []OrderBookEntry `json:"asks"`
	Bids      []OrderBookEntry `json:"bids"`
}
