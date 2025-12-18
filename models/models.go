package models

import (
	"time"

	"github.com/google/uuid"
)

type RewardEvent struct {
	ID          uuid.UUID
	UserID      int64
	StockSymbol string
	Shares    float64
	ReferenceID    string
	RewardedAt  time.Time
	CreatedAt   time.Time
}


type LedgerEntry struct {
	ID          uuid.UUID
	UserID      int64
	EntryType   string
	StockSymbol *string
	Quantity    *float64
	AmountINR   *float64
	Direction   string
	ReferenceID uuid.UUID
	CreatedAt   time.Time
}

type User struct {
	ID        int64
	CreatedAt time.Time
	Name      string
	Email     string
	Password  string	
}