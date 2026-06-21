package handler

import (
	"context"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/usecase"
	"github.com/KENTA0326/run-sync-pro/model"
)

// Auth は認証処理のポート。利用側(handler)が必要な振る舞いだけを書く。
type Auth interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
	GenerateToken(userID uint) (string, error)
}

// AuthUseCasePort はユーザー認証ユースケースのポート。
type AuthUseCasePort interface {
	SignUp(ctx context.Context, input usecase.SignUpInput) (*usecase.SignUpOutput, error)
	Login(ctx context.Context, input usecase.LoginInput) (*usecase.LoginOutput, error)
}

// ShoeUseCasePort はシューズ関連ユースケースのポート。
type ShoeUseCasePort interface {
	Create(ctx context.Context, input usecase.CreateShoeInput) (*domain.Shoe, error)
	GetByID(ctx context.Context, id, userID uint) (*domain.Shoe, error)
	List(ctx context.Context, userID uint, limit, offset int) (*usecase.ShoeListOutput, error)
	Delete(ctx context.Context, id, userID uint) error
}

// TrainingLogUseCasePort は走行ログ関連ユースケースのポート。
type TrainingLogUseCasePort interface {
	Create(ctx context.Context, input usecase.CreateTrainingLogInput) error
	GetByID(ctx context.Context, id, userID uint) (*domain.TrainingLog, error)
	List(ctx context.Context, userID uint, limit, offset int) (*usecase.TrainingLogListOutput, error)
	ListAll(ctx context.Context, userID uint) ([]*domain.TrainingLog, error)
}

// PasswordResetUseCasePort はパスワードリセットのポート。
type PasswordResetUseCasePort interface {
	RequestReset(ctx context.Context, input usecase.RequestResetInput) (*usecase.RequestResetOutput, error)
	ConfirmReset(ctx context.Context, input usecase.ConfirmResetInput) error
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

// PasswordResetTTL はリセットトークンのTTLを返す関数型。
type PasswordResetTTLFunc func() time.Duration

// FrontendBaseURLFunc はフロントエンドのベースURLを返す関数型。
type FrontendBaseURLFunc func() string
