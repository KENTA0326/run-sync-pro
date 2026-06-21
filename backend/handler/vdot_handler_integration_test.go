package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KENTA0326/run-sync-pro/internal/domainservice"
	"github.com/gin-gonic/gin"
)

// 実装の VDOTCalculator を注入し、HTTP から計算結果まで一気通貫で検証する。
func TestVDOTCalculate_integration_realCalculator(t *testing.T) {
	t.Parallel()

	calc := domainservice.NewVDOTCalculator()
	h := newTestHandlers(nil, nil, nil, nil, calc)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/vdot/calculate", map[string]any{
		"distance_meters": 10000,
		"time_seconds":    2400,
	})

	h.VDOTCalculate(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["vdot"] != 51.9 {
		t.Fatalf("vdot=%v want 51.9", body["vdot"])
	}
	if body["riegel_exponent"] != 1.08 {
		t.Fatalf("riegel_exponent=%v want 1.08", body["riegel_exponent"])
	}

	riegel, ok := body["riegel_predictions"].(map[string]any)
	if !ok {
		t.Fatalf("riegel_predictions=%T", body["riegel_predictions"])
	}
	if riegel["full_seconds"] != float64(11363) {
		t.Fatalf("full_seconds=%v want 11363", riegel["full_seconds"])
	}

	paces, ok := body["paces"].(map[string]any)
	if !ok {
		t.Fatalf("paces=%T", body["paces"])
	}
	if paces["marathon_sec_per_km"] == nil || paces["marathon_sec_per_km"].(float64) <= 0 {
		t.Fatalf("unexpected marathon pace: %v", paces["marathon_sec_per_km"])
	}
}
