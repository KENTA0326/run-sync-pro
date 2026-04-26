package model

import (
	"time"

	"gorm.io/gorm"
)

type Shoe struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `json:"user_id"`
	Brand         string         `json:"brand"`
	Model         string         `json:"model"`
	PurchaseDate  time.Time      `json:"purchase_date"`
	TotalDistance float64        `gorm:"default:0" json:"total_distance"` // 累計走行距離
	IsActive      bool           `gorm:"default:true" json:"is_active"`   // まだ履いているか
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
