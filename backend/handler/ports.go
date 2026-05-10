package handler

import (
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/KENTA0326/run-sync-pro/service"
)

// Auth は認証処理のポート。利用側(handler)が必要な振る舞いだけを書く。
type Auth interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
	GenerateToken(userID uint) (string, error)
}

// TrainingAnalyzer は月別解析のポート。
type TrainingAnalyzer interface {
	AnalyzeByMonth(logs []model.TrainingLog) service.AnalysisResponse
}

// VDOTCalculator は VDOT・ペース・リーゲル予測のポート。
type VDOTCalculator interface {
	CalculateVDOT(distanceMeters, timeSeconds float64) float64
	CalculateTrainingPaces(vdot float64) service.TrainingPaces
	CalculateRiegelPredictions(distanceMeters, timeSeconds, exponent float64) service.RacePredictions
}

// MarathonSplits はフルマラソンスプリット生成のポート。
type MarathonSplits interface {
	GenerateFullMarathonSplits(paceSecPerKm float64) []service.SplitRow
}
