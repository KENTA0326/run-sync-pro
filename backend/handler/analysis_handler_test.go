package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/testutil"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/gin-gonic/gin"
)

func TestMonthlyReport_cacheHit_skipsDBAndAnalyzer(t *testing.T) {
	t.Parallel()

	want := model.AnalysisResponse{
		MonthlyReports: []model.MonthlyReport{
			{YearMonth: "2026-03", RunCount: 5, TotalDistance: 50},
		},
		TotalRunCount: 5,
		TotalDistance: 50,
	}

	analyzerCalled := false
	h := newTestHandlers(nil, nil, &fakeAnalyzer{
		analyze: func(_ []model.TrainingLog) model.AnalysisResponse {
			analyzerCalled = true
			return model.AnalysisResponse{}
		},
	}, &fakeAnalysisCache{
		get: func(_ context.Context, userID domain.UserID) (model.AnalysisResponse, bool, error) {
			if userID != 7 {
				t.Fatalf("userID=%d want 7", userID)
			}
			return want, true, nil
		},
	}, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/analysis/monthly", nil)
	req = req.WithContext(domain.WithUserID(req.Context(), domain.UserID(7)))
	c.Request = req

	h.MonthlyReport(c)

	if analyzerCalled {
		t.Fatal("analyzer should not be called on cache hit")
	}
	testutil.AssertResponseJSON(t, rec, http.StatusOK, map[string]any{
		"monthly_reports": []any{
			map[string]any{
				"year_month":         "2026-03",
				"total_distance":     50.0,
				"total_duration":     float64(0),
				"run_count":          float64(5),
				"avg_pace_sec_per_km": 0.0,
				"avg_vdot":           0.0,
				"max_vdot":           0.0,
			},
		},
		"total_distance":  50.0,
		"total_duration":  float64(0),
		"total_run_count": float64(5),
	})
}
