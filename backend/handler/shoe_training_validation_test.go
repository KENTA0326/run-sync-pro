package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/timeutil"
	"github.com/gin-gonic/gin"
)

func TestCreateShoeInput_ShouldBindJSON_caldate(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"brand":"a","model":"b","purchase_date":"13-01-2026"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	var input createShoeInput
	if err := c.ShouldBindJSON(&input); err == nil {
		t.Fatal("expected bind error for invalid caldate")
	}
}

func TestCreateTrainingLogInput_ShouldBindJSON_future_training_date(t *testing.T) {
	t.Parallel()
	future := timeutil.NowJST().AddDate(0, 0, 2).Format("2006-01-02")
	body := `{"training_date":"` + future + `","distance":1,"duration":60,"pace":"6:00","kind":0,"shoe_id":1}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var input createTrainingLogInput
	if err := c.ShouldBindJSON(&input); err == nil {
		t.Fatal("expected bind error for future training_date")
	}
}

func TestCreateTrainingLogInput_ShouldBindJSON_today_ok(t *testing.T) {
	t.Parallel()
	today := time.Date(timeutil.NowJST().Year(), timeutil.NowJST().Month(), timeutil.NowJST().Day(), 0, 0, 0, 0, timeutil.JST).Format("2006-01-02")
	body := `{"training_date":"` + today + `","distance":1,"duration":60,"pace":"6:00","kind":0,"shoe_id":1}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var input createTrainingLogInput
	if err := c.ShouldBindJSON(&input); err != nil {
		t.Fatalf("bind: %v", err)
	}
}
