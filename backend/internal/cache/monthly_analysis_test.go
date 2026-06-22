package cache

import (
	"context"
	"testing"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestMonthlyAnalysis_GetSetInvalidate(t *testing.T) {
	t.Parallel()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	c := NewMonthlyAnalysisForTest(client, time.Minute)

	ctx := context.Background()
	userID := domain.UserID(42)
	want := model.AnalysisResponse{
		MonthlyReports: []model.MonthlyReport{{YearMonth: "2026-05", RunCount: 3}},
		TotalRunCount:  3,
	}

	_, hit, err := c.Get(ctx, userID)
	if err != nil || hit {
		t.Fatalf("initial get: hit=%v err=%v", hit, err)
	}

	if err := c.Set(ctx, userID, want); err != nil {
		t.Fatal(err)
	}

	got, hit, err := c.Get(ctx, userID)
	if err != nil || !hit {
		t.Fatalf("second get: hit=%v err=%v", hit, err)
	}
	if got.TotalRunCount != want.TotalRunCount {
		t.Fatalf("got %+v", got)
	}

	if err := c.Invalidate(ctx, userID); err != nil {
		t.Fatal(err)
	}
	_, hit, err = c.Get(ctx, userID)
	if err != nil || hit {
		t.Fatalf("after invalidate: hit=%v err=%v", hit, err)
	}
}

func TestNoopMonthlyAnalysis(t *testing.T) {
	t.Parallel()
	c := NoopMonthlyAnalysis{}
	_, hit, err := c.Get(context.Background(), 1)
	if err != nil || hit {
		t.Fatal("noop should miss")
	}
}
