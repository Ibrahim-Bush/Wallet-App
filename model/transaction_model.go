package model

import (
	"time"
)

type Transaction struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	WalletID        int       `json:"wallet_id" gorm:"not null"`
	Type            string    `json:"type" gorm:"not null"`
	Amount          int       `json:"amount" gorm:"not null"`
	Category        string    `json:"category" gorm:"not null"`
	Note            string    `json:"note"`
	RelatedWalletID *int      `json:"related_wallet_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type Transaction_summary struct {
	Category string `json:"category"`
	Total    int    `json:"total"`
}
