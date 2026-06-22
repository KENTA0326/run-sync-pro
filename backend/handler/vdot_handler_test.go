package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/testutil"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/gin-gonic/gin"
)

func TestVDOTCalculate_success(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(nil, nil, nil, nil, &fakeVDOT{
		vdot: 52.3,
		paces: model.TrainingPaces{
			EasyMinSecPerKm: 330,
			EasyMaxSecPerKm: 360,
		},
		riegel: model.RacePredictions{
			HalfSeconds: 5400,
			FullSeconds: 11400,
		},
	})

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/vdot/calculate", map[string]any{
		"distance_meters": 10000,
		"time_seconds":    2400,
	})

	h.VDOTCalculate(c)

	testutil.AssertResponseJSON(t, rec, http.StatusOK, map[string]any{
		"vdot":            52.3,
		"paces":           map[string]any{"easy_min_sec_per_km": 330.0, "easy_max_sec_per_km": 360.0, "marathon_sec_per_km": 0.0, "threshold_sec_per_km": 0.0, "interval_sec_per_km": 0.0, "repetition_sec_per_km": 0.0},
		"riegel_exponent": 1.08,
		"riegel_predictions": map[string]any{
			"full_seconds":  float64(11400),
			"half_seconds":  float64(5400),
			"ten_k_seconds": float64(0),
			"five_k_seconds": float64(0),
		},
	})
}

func TestVDOTCalculate_customRiegelExponent(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(nil, nil, nil, nil, &fakeVDOT{vdot: 45.0})

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/vdot/calculate", map[string]any{
		"distance_meters": 5000,
		"time_seconds":    1200,
		"riegel_exponent": 1.12,
	})

	h.VDOTCalculate(c)

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["riegel_exponent"] != 1.12 {
		t.Fatalf("riegel_exponent=%v want 1.12", body["riegel_exponent"])
	}
}

func TestVDOTCalculate_invalidInput(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(nil, nil, nil, nil, &fakeVDOT{})

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/vdot/calculate", map[string]any{
		"distance_meters": 0,
		"time_seconds":    2400,
	})

	h.VDOTCalculate(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestVDOTCalculate_zeroVDOT(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(nil, nil, nil, nil, &fakeVDOT{vdot: 0})

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/vdot/calculate", map[string]any{
		"distance_meters": 1000,
		"time_seconds":    300,
	})

	h.VDOTCalculate(c)

	testutil.AssertResponseJSON(t, rec, http.StatusBadRequest, map[string]any{
		"error": map[string]any{
			"code":    apperrors.CodeInvalidInput,
			"message": "VDOTを算出できません。距離とタイムを確認してください（3km以上推奨）",
		},
	})
}
