package domain

import "time"

// Shoe はシューズのドメインエンティティ。
type Shoe struct {
	ID            uint
	UserID        uint
	Brand         string
	Model         string
	PurchaseDate  time.Time
	TotalDistance float64
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
