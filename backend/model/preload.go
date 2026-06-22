package model

import "gorm.io/gorm"

// PreloadTrainingLogAssociations は走行ログ取得時の N+1 回避用 Preload チェーン。
func PreloadTrainingLogAssociations(db *gorm.DB) *gorm.DB {
	return db.Preload("Shoe")
}
