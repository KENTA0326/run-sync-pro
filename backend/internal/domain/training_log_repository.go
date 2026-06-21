package domain

import "context"

// TrainingLogRepository は走行ログの永続化操作を定義するインターフェース。
type TrainingLogRepository interface {
	Create(ctx context.Context, log *TrainingLog) error
	FindByID(ctx context.Context, id, userID uint) (*TrainingLog, error)
	List(ctx context.Context, userID uint, limit, offset int) ([]*TrainingLog, int64, error)
	ListAll(ctx context.Context, userID uint) ([]*TrainingLog, error)
}
