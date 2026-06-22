package domain

import "time"

// TrainingLog は走行ログのドメインエンティティ。
type TrainingLog struct {
	ID           uint
	UserID       uint
	TrainingDate time.Time
	Distance     float64
	Duration     int
	Pace         string
	Memo         string
	Kind         int
	ShoeID       uint
	Shoe         *Shoe
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
