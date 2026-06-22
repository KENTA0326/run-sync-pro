package model

// MonthlyReport は1ヶ月分の走行集計（API レスポンス DTO）。
type MonthlyReport struct {
	YearMonth       string  `json:"year_month"`
	TotalDistance   float64 `json:"total_distance"`
	TotalDuration   int     `json:"total_duration"`
	RunCount        int     `json:"run_count"`
	AvgPaceSecPerKm float64 `json:"avg_pace_sec_per_km"`
	AvgVDOT         float64 `json:"avg_vdot"`
	MaxVDOT         float64 `json:"max_vdot"`
}

// AnalysisResponse は月別解析 API のレスポンス DTO。
type AnalysisResponse struct {
	MonthlyReports []MonthlyReport `json:"monthly_reports"`
	TotalDistance  float64         `json:"total_distance"`
	TotalDuration  int             `json:"total_duration"`
	TotalRunCount  int             `json:"total_run_count"`
}

// TrainingPaces は VDOT に基づく各強度の推奨ペース（1km あたりの秒数）DTO。
type TrainingPaces struct {
	EasyMinSecPerKm    float64 `json:"easy_min_sec_per_km"`
	EasyMaxSecPerKm    float64 `json:"easy_max_sec_per_km"`
	MarathonSecPerKm   float64 `json:"marathon_sec_per_km"`
	ThresholdSecPerKm  float64 `json:"threshold_sec_per_km"`
	IntervalSecPerKm   float64 `json:"interval_sec_per_km"`
	RepetitionSecPerKm float64 `json:"repetition_sec_per_km"`
}

// RacePredictions は VDOT / リーゲルから推定した各距離のタイム（秒）DTO。
type RacePredictions struct {
	FullSeconds  int `json:"full_seconds"`
	HalfSeconds  int `json:"half_seconds"`
	TenKSeconds  int `json:"ten_k_seconds"`
	FiveKSeconds int `json:"five_k_seconds"`
}

// SplitRow はフルマラソンスプリット表の1行 DTO。
type SplitRow struct {
	Km                float64 `json:"km"`
	Label             string  `json:"label"`
	CumulativeSeconds int     `json:"cumulative_seconds"`
}
