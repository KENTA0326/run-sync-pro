package usecase

import (
	"context"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
)

// CreateTrainingLogInput は走行ログ作成の入力。
type CreateTrainingLogInput struct {
	UserID       uint
	TrainingDate time.Time
	Distance     float64
	Duration     int
	Pace         string
	Memo         string
	Kind         int
	ShoeID       uint
}

// TrainingLogUseCase は走行ログ関連のユースケース。
type TrainingLogUseCase struct {
	logRepo  domain.TrainingLogRepository
	shoeRepo domain.ShoeRepository
}

// NewTrainingLogUseCase は TrainingLogUseCase を生成する。
func NewTrainingLogUseCase(logRepo domain.TrainingLogRepository, shoeRepo domain.ShoeRepository) *TrainingLogUseCase {
	return &TrainingLogUseCase{logRepo: logRepo, shoeRepo: shoeRepo}
}

// Create は走行ログを作成し、シューズの累計距離を更新する。
func (uc *TrainingLogUseCase) Create(ctx context.Context, input CreateTrainingLogInput) error {
	_, err := uc.shoeRepo.FindByID(ctx, input.ShoeID, input.UserID)
	if err != nil {
		return err
	}

	log := &domain.TrainingLog{
		UserID:       input.UserID,
		TrainingDate: input.TrainingDate,
		Distance:     input.Distance,
		Duration:     input.Duration,
		Pace:         input.Pace,
		Memo:         input.Memo,
		Kind:         input.Kind,
		ShoeID:       input.ShoeID,
	}

	if err := uc.logRepo.Create(ctx, log); err != nil {
		return err
	}

	return uc.shoeRepo.AddDistance(ctx, input.ShoeID, input.Distance)
}

// GetByID は走行ログ1件を取得する。
func (uc *TrainingLogUseCase) GetByID(ctx context.Context, id, userID uint) (*domain.TrainingLog, error) {
	return uc.logRepo.FindByID(ctx, id, userID)
}

// TrainingLogListOutput は走行ログ一覧の出力。
type TrainingLogListOutput struct {
	Logs  []*domain.TrainingLog
	Total int64
}

// List は走行ログ一覧を取得する。
func (uc *TrainingLogUseCase) List(ctx context.Context, userID uint, limit, offset int) (*TrainingLogListOutput, error) {
	logs, total, err := uc.logRepo.List(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return &TrainingLogListOutput{Logs: logs, Total: total}, nil
}

// ListAll はユーザーの全走行ログを取得する（解析用）。
func (uc *TrainingLogUseCase) ListAll(ctx context.Context, userID uint) ([]*domain.TrainingLog, error) {
	return uc.logRepo.ListAll(ctx, userID)
}
