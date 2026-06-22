package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
)

type fakeShoeRepo struct {
	create         func(ctx context.Context, shoe *domain.Shoe) error
	findByID       func(ctx context.Context, id, userID uint) (*domain.Shoe, error)
	listActive     func(ctx context.Context, userID uint, limit, offset int) ([]*domain.Shoe, int64, error)
	listAllActive  func(ctx context.Context, userID uint) ([]*domain.Shoe, error)
	deactivate     func(ctx context.Context, id, userID uint) error
}

func (f *fakeShoeRepo) Create(ctx context.Context, shoe *domain.Shoe) error {
	if f.create != nil {
		return f.create(ctx, shoe)
	}
	shoe.ID = 1
	return nil
}

func (f *fakeShoeRepo) FindByID(ctx context.Context, id, userID uint) (*domain.Shoe, error) {
	if f.findByID != nil {
		return f.findByID(ctx, id, userID)
	}
	return nil, domain.ErrShoeNotFound
}

func (f *fakeShoeRepo) ListActive(ctx context.Context, userID uint, limit, offset int) ([]*domain.Shoe, int64, error) {
	if f.listActive != nil {
		return f.listActive(ctx, userID, limit, offset)
	}
	return nil, 0, nil
}

func (f *fakeShoeRepo) ListAllActive(ctx context.Context, userID uint) ([]*domain.Shoe, error) {
	if f.listAllActive != nil {
		return f.listAllActive(ctx, userID)
	}
	return nil, nil
}

func (f *fakeShoeRepo) Deactivate(ctx context.Context, id, userID uint) error {
	if f.deactivate != nil {
		return f.deactivate(ctx, id, userID)
	}
	return nil
}

func (f *fakeShoeRepo) AddDistance(ctx context.Context, id uint, distance float64) error {
	return nil
}

func TestShoeUseCase_Create_setsDefaults(t *testing.T) {
	t.Parallel()

	purchaseDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	var saved *domain.Shoe
	uc := NewShoeUseCase(&fakeShoeRepo{
		create: func(_ context.Context, shoe *domain.Shoe) error {
			saved = shoe
			shoe.ID = 5
			return nil
		},
	})

	got, err := uc.Create(context.Background(), CreateShoeInput{
		UserID:       3,
		Brand:        "Brooks",
		Model:        "Ghost",
		PurchaseDate: purchaseDate,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 5 || !got.IsActive {
		t.Fatalf("got=%+v", got)
	}
	if saved.Brand != "Brooks" || saved.UserID != 3 {
		t.Fatalf("saved=%+v", saved)
	}
}

func TestShoeUseCase_List_distinctBrands(t *testing.T) {
	t.Parallel()

	uc := NewShoeUseCase(&fakeShoeRepo{
		listActive: func(_ context.Context, userID uint, limit, offset int) ([]*domain.Shoe, int64, error) {
			return []*domain.Shoe{
				{ID: 1, Brand: "Nike"},
				{ID: 2, Brand: "Asics"},
			}, 2, nil
		},
		listAllActive: func(_ context.Context, _ uint) ([]*domain.Shoe, error) {
			return []*domain.Shoe{
				{ID: 1, Brand: "Nike"},
				{ID: 2, Brand: "Asics"},
				{ID: 3, Brand: "Nike"},
			}, nil
		},
	})

	out, err := uc.List(context.Background(), 1, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if out.Total != 2 || len(out.Shoes) != 2 {
		t.Fatalf("shoes=%d total=%d", len(out.Shoes), out.Total)
	}
	if len(out.Brands) != 2 || out.Brands[0] != "Nike" || out.Brands[1] != "Asics" {
		t.Fatalf("brands=%v", out.Brands)
	}
}

func TestShoeUseCase_Delete_delegatesToRepo(t *testing.T) {
	t.Parallel()

	called := false
	uc := NewShoeUseCase(&fakeShoeRepo{
		deactivate: func(_ context.Context, id, userID uint) error {
			called = true
			if id != 9 || userID != 2 {
				t.Fatalf("id=%d userID=%d", id, userID)
			}
			return nil
		},
	})

	if err := uc.Delete(context.Background(), 9, 2); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("Deactivate was not called")
	}
}

func TestShoeUseCase_Create_repoError(t *testing.T) {
	t.Parallel()

	uc := NewShoeUseCase(&fakeShoeRepo{
		create: func(_ context.Context, _ *domain.Shoe) error {
			return errors.New("db error")
		},
	})

	_, err := uc.Create(context.Background(), CreateShoeInput{
		UserID:       1,
		Brand:        "X",
		Model:        "Y",
		PurchaseDate: time.Now(),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
