package domain

import "context"

// ShoeRepository はシューズの永続化操作を定義するインターフェース。
type ShoeRepository interface {
	Create(ctx context.Context, shoe *Shoe) error
	FindByID(ctx context.Context, id, userID uint) (*Shoe, error)
	ListActive(ctx context.Context, userID uint, limit, offset int) ([]*Shoe, int64, error)
	ListAllActive(ctx context.Context, userID uint) ([]*Shoe, error)
	Deactivate(ctx context.Context, id, userID uint) error
	AddDistance(ctx context.Context, id uint, distance float64) error
}
