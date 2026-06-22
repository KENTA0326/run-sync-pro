package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Email     Email          `gorm:"uniqueIndex;not null;column:email;type:text" json:"email"`
	Password  string         `json:"-"`
	Role      int            `gorm:"default:1" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Shoes        []Shoe        `gorm:"foreignKey:UserID;references:ID" json:"-"`
	TrainingLogs []TrainingLog `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
