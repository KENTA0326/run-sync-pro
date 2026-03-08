package model

import (
	"time"

	"gorm.io/gorm"
)

type TrainingLog struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `json:"user_id"`
	TrainingDate time.Time      `json:"training_date"`
	Distance     float64        `json:"distance"`      // km単位
	Duration     int            `json:"duration"`      // 合計秒数（例: 3600 = 1時間）
	Pace         string         `json:"pace"`          // 例: "5:30"
	Memo         string         `json:"memo"`
	Kind         int            `json:"kind"`          // 0:ジョグ, 1:LSD, 2:ペース走, 3:インターバル
	ShoeID       uint           `json:"shoe_id"`       // 使用したシューズ
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}