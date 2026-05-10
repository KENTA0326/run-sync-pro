package handler

import (
	"github.com/KENTA0326/run-sync-pro/service"
	"gorm.io/gorm"
)

// Handlers は HTTP ハンドラとその依存（DB・ドメインサービス）をまとめる。
type Handlers struct {
	db       *gorm.DB
	auth     Auth
	analyzer TrainingAnalyzer
	vdot     VDOTCalculator
	splits   MarathonSplits
}

// NewHandlers は依存を明示的に渡して構築する（テストでモックを注入する想定）。
func NewHandlers(
	db *gorm.DB,
	auth Auth,
	analyzer TrainingAnalyzer,
	vdot VDOTCalculator,
	splits MarathonSplits,
) *Handlers {
	return &Handlers{
		db:       db,
		auth:     auth,
		analyzer: analyzer,
		vdot:     vdot,
		splits:   splits,
	}
}

// NewDefaultHandlers は本番相当の service 具体実装で Handlers を組み立てる。
func NewDefaultHandlers(db *gorm.DB) *Handlers {
	return NewHandlers(
		db,
		service.NewAuth(),
		service.NewAnalyzer(),
		service.NewVDOTCalculator(),
		service.NewMarathonSplits(),
	)
}
