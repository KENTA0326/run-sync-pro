package domainservice

import (
	"math"
	"testing"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/testutil"
	"github.com/KENTA0326/run-sync-pro/model"
)

func TestAnalyzeByMonth(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name          string
		fixture       string
		logs          []model.TrainingLog
		wantMonths    []string
		wantTotalRuns int
		wantTotalDist float64
		wantTotalDur  int
	}{
		{
			name:          "empty_nil_slice",
			logs:          nil,
			wantMonths:    nil,
			wantTotalRuns: 0,
			wantTotalDist: 0,
			wantTotalDur:  0,
		},
		{
			name:          "empty_zero_length",
			logs:          []model.TrainingLog{},
			wantMonths:    nil,
			wantTotalRuns: 0,
			wantTotalDist: 0,
			wantTotalDur:  0,
		},
		{
			name: "fixture_two_months",
			// 仕様変更時は internal/testutil/testdata の JSON とこの期待値をセットで更新する。
			fixture:       "training_logs/two_months.json",
			wantMonths:    []string{"2026-03", "2026-04"},
			wantTotalRuns: 2,
			wantTotalDist: 15,
			wantTotalDur:  5400,
		},
		{
			name: "serial_group_logs_below_threshold",
			logs: []model.TrainingLog{
				{TrainingDate: base, Distance: 5, Duration: 1800},
				{TrainingDate: base.Add(2 * time.Hour), Distance: 5, Duration: 1800},
			},
			wantMonths:    []string{"2026-03"},
			wantTotalRuns: 2,
			wantTotalDist: 10,
			wantTotalDur:  3600,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var logs []model.TrainingLog
			switch {
			case tc.fixture != "":
				testutil.LoadFixture(t, tc.fixture, &logs)
			default:
				logs = tc.logs
			}

			res := AnalyzeByMonth(logs)

			if len(res.MonthlyReports) != len(tc.wantMonths) {
				t.Fatalf("MonthlyReports len=%d want %d (%v)", len(res.MonthlyReports), len(tc.wantMonths), tc.wantMonths)
			}
			for i, ym := range tc.wantMonths {
				if res.MonthlyReports[i].YearMonth != ym {
					t.Fatalf("report[%d].YearMonth=%q want %q", i, res.MonthlyReports[i].YearMonth, ym)
				}
			}
			if res.TotalRunCount != tc.wantTotalRuns {
				t.Fatalf("TotalRunCount=%d want %d", res.TotalRunCount, tc.wantTotalRuns)
			}
			if res.TotalDistance != tc.wantTotalDist {
				t.Fatalf("TotalDistance=%v want %v", res.TotalDistance, tc.wantTotalDist)
			}
			if res.TotalDuration != tc.wantTotalDur {
				t.Fatalf("TotalDuration=%d want %d", res.TotalDuration, tc.wantTotalDur)
			}
		})
	}
}

func TestAnalyzeByMonth_vdotAggregation(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	// 10km 40:00 → VDOT 51.9（service/vdot_test.go と同じゴールデン値）
	res := AnalyzeByMonth([]model.TrainingLog{
		{TrainingDate: base, Distance: 10, Duration: 2400},
		{TrainingDate: base.Add(time.Hour), Distance: 5, Duration: 1200}, // 5km 20:00 → 49.8
	})

	if len(res.MonthlyReports) != 1 {
		t.Fatalf("reports=%d want 1", len(res.MonthlyReports))
	}
	r := res.MonthlyReports[0]
	if r.MaxVDOT != 51.9 {
		t.Fatalf("MaxVDOT=%v want 51.9", r.MaxVDOT)
	}
	wantAvg := (51.9 + 49.8) / 2
	if math.Abs(r.AvgVDOT-wantAvg) > 0.01 {
		t.Fatalf("AvgVDOT=%v want %v", r.AvgVDOT, wantAvg)
	}
}

func TestAnalyzeByMonth_parallelGroupAndAnalyzeWorkers(t *testing.T) {
	t.Parallel()
	// groupByMonthParallelThreshold 以上で groupLogsByMonth と AnalyzeByMonth が並列経路へ入る。
	// Add/Done/Wait やチャネルの不整合があればデッドロックまたは誤結果になる。
	base := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	n := groupByMonthParallelThreshold + 1
	logs := make([]model.TrainingLog, n)
	for i := range logs {
		logs[i] = model.TrainingLog{
			TrainingDate: base.Add(time.Duration(i) * time.Hour),
			Distance:     5,
			Duration:     1800,
		}
	}
	res := AnalyzeByMonth(logs)
	if len(res.MonthlyReports) < 1 {
		t.Fatalf("expected at least one month report, got %d", len(res.MonthlyReports))
	}
	if res.TotalRunCount != n {
		t.Fatalf("TotalRunCount=%d want %d", res.TotalRunCount, n)
	}
}
