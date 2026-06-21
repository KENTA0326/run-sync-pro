package domainservice

import (
	"math"
	"testing"
)

func TestCalculateVDOT_knownRaces(t *testing.T) {
	t.Parallel()

	// ゴールデン値は本実装から算出した回帰用固定値（Jack Daniels 公式ベース）
	cases := []struct {
		name           string
		distanceMeters float64
		timeSeconds    float64
		wantVDOT       float64
	}{
		{name: "full_marathon_3h", distanceMeters: 42195, timeSeconds: 10800, wantVDOT: 53.5},
		{name: "10k_40min", distanceMeters: 10000, timeSeconds: 2400, wantVDOT: 51.9},
		{name: "5k_20min", distanceMeters: 5000, timeSeconds: 1200, wantVDOT: 49.8},
		{name: "1k_5min", distanceMeters: 1000, timeSeconds: 300, wantVDOT: 33.0},
		{name: "full_marathon_4h", distanceMeters: 42195, timeSeconds: 14400, wantVDOT: 37.9},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := CalculateVDOT(tc.distanceMeters, tc.timeSeconds)
			if got != tc.wantVDOT {
				t.Fatalf("CalculateVDOT()=%v want %v", got, tc.wantVDOT)
			}
		})
	}
}

func TestCalculateVDOT_invalidInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		distanceMeters float64
		timeSeconds    float64
	}{
		{name: "zero_distance", distanceMeters: 0, timeSeconds: 2400},
		{name: "zero_time", distanceMeters: 10000, timeSeconds: 0},
		{name: "negative_distance", distanceMeters: -100, timeSeconds: 600},
		{name: "negative_time", distanceMeters: 5000, timeSeconds: -1},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := CalculateVDOT(tc.distanceMeters, tc.timeSeconds); got != 0 {
				t.Fatalf("CalculateVDOT()=%v want 0", got)
			}
		})
	}
}

func TestCalculateVDOT_fasterTimeHigherVDOT(t *testing.T) {
	t.Parallel()

	slow := CalculateVDOT(10000, 2700) // 45:00
	fast := CalculateVDOT(10000, 2100) // 35:00
	if fast <= slow {
		t.Fatalf("faster run should yield higher VDOT: slow=%v fast=%v", slow, fast)
	}
}

func TestPredictRiegelSeconds_formula(t *testing.T) {
	t.Parallel()

	// 10km 40:00 から 5km を指数 1.08 で換算: 2400 * (5000/10000)^1.08 ≈ 1135
	got := PredictRiegelSeconds(2400, 10000, 5000, 1.08)
	if got != 1135 {
		t.Fatalf("PredictRiegelSeconds()=%d want 1135", got)
	}
}

func TestPredictRiegelSeconds_invalidInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		timeSec   float64
		baseDist  float64
		targetDist float64
		exponent  float64
	}{
		{name: "zero_time", timeSec: 0, baseDist: 10000, targetDist: 5000, exponent: 1.08},
		{name: "exponent_too_low", timeSec: 2400, baseDist: 10000, targetDist: 5000, exponent: 1.0},
		{name: "exponent_too_high", timeSec: 2400, baseDist: 10000, targetDist: 5000, exponent: 2.0},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := PredictRiegelSeconds(tc.timeSec, tc.baseDist, tc.targetDist, tc.exponent); got != 0 {
				t.Fatalf("PredictRiegelSeconds()=%d want 0", got)
			}
		})
	}
}

func TestCalculateRiegelPredictions_known10k40min(t *testing.T) {
	t.Parallel()

	got := CalculateRiegelPredictions(10000, 2400, 1.08)
	want := struct {
		full, half, tenK, fiveK int
	}{full: 11363, half: 5375, tenK: 2400, fiveK: 1135}

	if got.FullSeconds != want.full {
		t.Fatalf("FullSeconds=%d want %d", got.FullSeconds, want.full)
	}
	if got.HalfSeconds != want.half {
		t.Fatalf("HalfSeconds=%d want %d", got.HalfSeconds, want.half)
	}
	if got.TenKSeconds != want.tenK {
		t.Fatalf("TenKSeconds=%d want %d", got.TenKSeconds, want.tenK)
	}
	if got.FiveKSeconds != want.fiveK {
		t.Fatalf("FiveKSeconds=%d want %d", got.FiveKSeconds, want.fiveK)
	}
}

func TestCalculateRiegelPredictions_higherExponentLongerFull(t *testing.T) {
	t.Parallel()

	standard := CalculateRiegelPredictions(10000, 2400, 1.08)
	beginner := CalculateRiegelPredictions(10000, 2400, 1.12)
	if beginner.FullSeconds <= standard.FullSeconds {
		t.Fatalf("exponent 1.12 should predict slower full: std=%d beg=%d",
			standard.FullSeconds, beginner.FullSeconds)
	}
	if beginner.FullSeconds != 12037 {
		t.Fatalf("FullSeconds=%d want 12037", beginner.FullSeconds)
	}
}

func TestCalculateTrainingPaces_knownVDOT50(t *testing.T) {
	t.Parallel()

	got := CalculateTrainingPaces(50.0)
	want := struct {
		eMin, eMax, m, th, i, r float64
	}{eMin: 309.6, eMax: 358.1, m: 317.3, th: 260.6, i: 218.8, r: 196.9}

	assertFloatEq(t, "EasyMinSecPerKm", got.EasyMinSecPerKm, want.eMin)
	assertFloatEq(t, "EasyMaxSecPerKm", got.EasyMaxSecPerKm, want.eMax)
	assertFloatEq(t, "MarathonSecPerKm", got.MarathonSecPerKm, want.m)
	assertFloatEq(t, "ThresholdSecPerKm", got.ThresholdSecPerKm, want.th)
	assertFloatEq(t, "IntervalSecPerKm", got.IntervalSecPerKm, want.i)
	assertFloatEq(t, "RepetitionSecPerKm", got.RepetitionSecPerKm, want.r)
}

func TestCalculateTrainingPaces_intensityOrdering(t *testing.T) {
	t.Parallel()

	p := CalculateTrainingPaces(50.0)
	// 秒/km が大きいほど遅いペース。Easy(遅い側) > M > T > I > R
	if p.EasyMaxSecPerKm <= p.MarathonSecPerKm {
		t.Fatalf("easy max should be slower than marathon pace: easyMax=%v M=%v", p.EasyMaxSecPerKm, p.MarathonSecPerKm)
	}
	if p.EasyMinSecPerKm >= p.EasyMaxSecPerKm {
		t.Fatalf("easy range should be ordered: min=%v max=%v", p.EasyMinSecPerKm, p.EasyMaxSecPerKm)
	}
	if p.MarathonSecPerKm <= p.ThresholdSecPerKm {
		t.Fatalf("M should be slower than T: M=%v T=%v", p.MarathonSecPerKm, p.ThresholdSecPerKm)
	}
	if p.ThresholdSecPerKm <= p.IntervalSecPerKm {
		t.Fatalf("T should be slower than I: T=%v I=%v", p.ThresholdSecPerKm, p.IntervalSecPerKm)
	}
	if p.IntervalSecPerKm <= p.RepetitionSecPerKm {
		t.Fatalf("I should be slower than R: I=%v R=%v", p.IntervalSecPerKm, p.RepetitionSecPerKm)
	}
}

func TestCalculateTrainingPaces_higherVDOT_fasterPaces(t *testing.T) {
	t.Parallel()

	slow := CalculateTrainingPaces(40.0)
	fast := CalculateTrainingPaces(55.0)
	if fast.MarathonSecPerKm >= slow.MarathonSecPerKm {
		t.Fatalf("higher VDOT should yield faster marathon pace: slow=%v fast=%v",
			slow.MarathonSecPerKm, fast.MarathonSecPerKm)
	}
}

func TestPredictRaceTimeSeconds_roundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		distanceMeters float64
		timeSeconds    int
		toleranceSec   int
	}{
		{name: "full_marathon_3h", distanceMeters: 42195, timeSeconds: 10800, toleranceSec: 10},
		{name: "10k_40min", distanceMeters: 10000, timeSeconds: 2400, toleranceSec: 5},
		{name: "5k_20min", distanceMeters: 5000, timeSeconds: 1200, toleranceSec: 5},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			vdot := CalculateVDOT(tc.distanceMeters, float64(tc.timeSeconds))
			predicted := PredictRaceTimeSeconds(vdot, tc.distanceMeters)
			if predicted == 0 {
				t.Fatal("PredictRaceTimeSeconds returned 0")
			}
			diff := predicted - tc.timeSeconds
			if diff < 0 {
				diff = -diff
			}
			if diff > tc.toleranceSec {
				t.Fatalf("round-trip diff=%ds (predicted=%d original=%d vdot=%.1f)",
					diff, predicted, tc.timeSeconds, vdot)
			}
		})
	}
}

func TestCalculateRacePredictions_distanceOrdering(t *testing.T) {
	t.Parallel()

	got := CalculateRacePredictions(50.0)
	if !(got.FiveKSeconds < got.TenKSeconds && got.TenKSeconds < got.HalfSeconds && got.HalfSeconds < got.FullSeconds) {
		t.Fatalf("unexpected ordering: %+v", got)
	}
	// 回帰用固定値
	if got.FullSeconds != 11440 || got.HalfSeconds != 5491 || got.TenKSeconds != 2480 || got.FiveKSeconds != 1196 {
		t.Fatalf("golden mismatch: %+v", got)
	}
}

func TestVDOTCalculator_implementsPort(t *testing.T) {
	t.Parallel()

	calc := NewVDOTCalculator()
	if got := calc.CalculateVDOT(10000, 2400); got != 51.9 {
		t.Fatalf("wrapper CalculateVDOT=%v want 51.9", got)
	}
}

func assertFloatEq(t *testing.T, field string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.05 {
		t.Fatalf("%s=%v want %v", field, got, want)
	}
}
