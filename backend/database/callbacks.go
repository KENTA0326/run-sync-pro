package database

import (
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"gorm.io/gorm"
)

// RegisterCallbacks は GORM コールバックを登録する。
func RegisterCallbacks(db *gorm.DB) {
	db.Callback().Create().Before("gorm:create").Register("domain:set_created_by", setCreatedByCallback)
}

func setCreatedByCallback(db *gorm.DB) {
	if db.Statement.Context == nil || db.Statement.Schema == nil {
		return
	}
	if db.Statement.Schema.LookUpField("CreatedBy") == nil {
		return
	}
	id, ok := domain.UserIDFromStdContext(db.Statement.Context)
	if !ok {
		return
	}
	db.Statement.SetColumn("CreatedBy", id.Uint())
}
