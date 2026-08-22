package model

type Budget struct {
	ID           int    `json:"id" gorm:"primaryKey"`
	UserID       int    `json:"user_id" gorm:"not null;uniqueIndex:idx_user_category"`
	Category     string `json:"category" gorm:"not null;uniqueIndex:idx_user_category"`
	MonthlyLimit int    `json:"monthly_limit" gorm:"not null"`
}

type Create_budget_request struct {
	Category     string `json:"category" binding:"required"`
	MonthlyLimit int    `json:"monthly_limit" binding:"required"`
}

type Transaction_response struct {
	Transaction *Transaction `json:"transaction"`
	Warning     string       `json:"warning,omitempty"`
}

type Budget_status struct {
	Category     string `json:"category"`
	MonthlyLimit int    `json:"monthly_limit"`
	Spent        int    `json:"spent"`
	OverBudget   bool   `json:"over_budget"`
}
