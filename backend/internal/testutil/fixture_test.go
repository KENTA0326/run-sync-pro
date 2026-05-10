package testutil_test

import (
	"testing"

	"github.com/KENTA0326/run-sync-pro/internal/testutil"
	"github.com/KENTA0326/run-sync-pro/internal/timeutil"
	"github.com/KENTA0326/run-sync-pro/model"
)

func TestLoadFixture_trainingLogsTwoMonths(t *testing.T) {
	t.Parallel()
	var logs []model.TrainingLog
	testutil.LoadFixture(t, "training_logs/two_months.json", &logs)
	if len(logs) != 2 {
		t.Fatalf("len=%d want 2", len(logs))
	}
	if logs[0].Distance != 10 || logs[0].Duration != 3600 {
		t.Fatalf("first log: %+v", logs[0])
	}
	if logs[0].TrainingDate.IsZero() {
		t.Fatal("training_date not decoded")
	}
	if got := timeutil.MonthKey(logs[0].TrainingDate); got != "2026-03" {
		t.Fatalf("MonthKey: got %q want 2026-03", got)
	}
}
