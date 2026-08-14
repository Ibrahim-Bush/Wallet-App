package model

type Wallet struct {
	ID      int `json:"id" gorm:"primaryKey"`
	UserID  int `json:"user_id" gorm:"unique;not null"`
	Balance int `json:"balance" gorm:"default:0;not null"`
}

type Transfer_request struct {
	Amount     int    `json:"amount"`
	Category   string `json:"category"`
	Note       string `json:"note"`
	ToUsername string `json:"to_username"`
}
