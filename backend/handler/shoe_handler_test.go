package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/testutil"
	"github.com/KENTA0326/run-sync-pro/internal/usecase"
	"github.com/gin-gonic/gin"
)

func authedGET(path string, userID uint) *http.Request {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	return req.WithContext(domain.WithUserID(req.Context(), domain.UserID(userID)))
}

func TestCreateShoe_success(t *testing.T) {
	t.Parallel()

	purchaseDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	h := newTestHandlers(nil, &fakeShoeUC{
		create: func(_ context.Context, input usecase.CreateShoeInput) (*domain.Shoe, error) {
			if input.UserID != 10 || input.Brand != "Nike" || input.Model != "Pegasus" {
				t.Fatalf("unexpected input: %+v", input)
			}
			return &domain.Shoe{
				ID:            3,
				UserID:        10,
				Brand:         "Nike",
				Model:         "Pegasus",
				PurchaseDate:  purchaseDate,
				TotalDistance: 0,
				IsActive:      true,
			}, nil
		},
	}, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := jsonPOST("/shoes", map[string]string{
		"brand":         "Nike",
		"model":         "Pegasus",
		"purchase_date": "2026-01-15",
	})
	req = req.WithContext(domain.WithUserID(req.Context(), domain.UserID(10)))
	c.Request = req

	h.CreateShoe(c)

	testutil.AssertResponseJSON(t, rec, http.StatusOK, map[string]any{
		"id":             float64(3),
		"user_id":        float64(10),
		"brand":          "Nike",
		"model":          "Pegasus",
		"purchase_date":  "2026-01-15",
		"total_distance": 0.0,
		"is_active":      true,
	})
}

func TestCreateShoe_unauthorized(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(nil, &fakeShoeUC{}, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/shoes", map[string]string{
		"brand":         "Nike",
		"model":         "Pegasus",
		"purchase_date": "2026-01-15",
	})

	h.CreateShoe(c)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestListShoes_success(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(nil, &fakeShoeUC{
		list: func(_ context.Context, userID uint, limit, offset int) (*usecase.ShoeListOutput, error) {
			if userID != 5 || limit != 50 || offset != 0 {
				t.Fatalf("userID=%d limit=%d offset=%d", userID, limit, offset)
			}
			return &usecase.ShoeListOutput{
				Shoes: []*domain.Shoe{
					{ID: 1, UserID: 5, Brand: "Asics", Model: "Magic", IsActive: true},
				},
				Brands: []string{"Asics"},
				Total:  1,
			}, nil
		},
	}, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = authedGET("/shoes", 5)

	h.ListShoes(c)

	testutil.AssertResponseJSON(t, rec, http.StatusOK, map[string]any{
		"shoes": []any{
			map[string]any{
				"id":             float64(1),
				"user_id":        float64(5),
				"brand":          "Asics",
				"model":          "Magic",
				"purchase_date":  "0001-01-01",
				"total_distance": 0.0,
				"is_active":      true,
			},
		},
		"brands": []any{"Asics"},
		"pagination": map[string]any{
			"limit":  float64(50),
			"offset": float64(0),
			"total":  float64(1),
		},
	})
}
