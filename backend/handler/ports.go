package handler

import (
	"context"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/model"
)

// Auth は認証処理のポート。利用側(handler)が必要な振る舞いだけを書く。
type Auth interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
	GenerateToken(userID uint) (string, error)
}

// TrainingAnalyzer は月別解析のポート。
type TrainingAnalyzer interface {
	AnalyzeByMonth(logs []model.TrainingLog) model.AnalysisResponse
}

// VDOTCalculator は VDOT・ペース・リーゲル予測のポート。
type VDOTCalculator interface {
	CalculateVDOT(distanceMeters, timeSeconds float64) float64
	CalculateTrainingPaces(vdot float64) model.TrainingPaces
	CalculateRiegelPredictions(distanceMeters, timeSeconds, exponent float64) model.RacePredictions
}

// MarathonSplits はフルマラソンスプリット生成のポート。
type MarathonSplits interface {
	GenerateFullMarathonSplits(paceSecPerKm float64) []model.SplitRow
}

// MonthlyAnalysisCache は月別解析結果のキャッシュポート。
type MonthlyAnalysisCache interface {
	Get(ctx context.Context, userID domain.UserID) (model.AnalysisResponse, bool, error)
	Set(ctx context.Context, userID domain.UserID, res model.AnalysisResponse) error
	Invalidate(ctx context.Context, userID domain.UserID) error
}
