package usecase

import (
	"context"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
)

// CreateShoeInput はシューズ登録の入力。
type CreateShoeInput struct {
	UserID       uint
	Brand        string
	Model        string
	PurchaseDate time.Time
}

// ShoeUseCase はシューズ関連のユースケース。
type ShoeUseCase struct {
	repo domain.ShoeRepository
}

func NewShoeUseCase(repo domain.ShoeRepository) *ShoeUseCase {
	return &ShoeUseCase{repo: repo}
}

// Create はシューズを新規登録する。
func (uc *ShoeUseCase) Create(ctx context.Context, input CreateShoeInput) (*domain.Shoe, error) {
	shoe := &domain.Shoe{
		UserID:       input.UserID,
		Brand:        input.Brand,
		Model:        input.Model,
		PurchaseDate: input.PurchaseDate,
		IsActive:     true,
	}
	if err := uc.repo.Create(ctx, shoe); err != nil {
		return nil, err
	}
	return shoe, nil
}

// GetByID はシューズ1件を取得する。
func (uc *ShoeUseCase) GetByID(ctx context.Context, id, userID uint) (*domain.Shoe, error) {
	return uc.repo.FindByID(ctx, id, userID)
}

// ShoeListOutput はシューズ一覧の出力。
type ShoeListOutput struct {
	Shoes  []*domain.Shoe
	Brands []string
	Total  int64
}

// List はアクティブなシューズ一覧を取得する。
func (uc *ShoeUseCase) List(ctx context.Context, userID uint, limit, offset int) (*ShoeListOutput, error) {
	shoes, total, err := uc.repo.ListActive(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	allShoes, err := uc.repo.ListAllActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	brands := distinctBrands(allShoes)

	return &ShoeListOutput{Shoes: shoes, Brands: brands, Total: total}, nil
}

// Delete はシューズを論理削除する。
func (uc *ShoeUseCase) Delete(ctx context.Context, id, userID uint) error {
	return uc.repo.Deactivate(ctx, id, userID)
}

func distinctBrands(shoes []*domain.Shoe) []string {
	seen := make(map[string]struct{})
	var brands []string
	for _, s := range shoes {
		if _, ok := seen[s.Brand]; !ok {
			seen[s.Brand] = struct{}{}
			brands = append(brands, s.Brand)
		}
	}
	return brands
}
