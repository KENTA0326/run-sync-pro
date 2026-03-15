package service

import (
	"sort"
	"sync"
	"time"

	"github.com/KENTA0326/run-sync-pro/model"
)

// MonthlyReport は1ヶ月分の集計結果
type MonthlyReport struct {
	YearMonth       string  `json:"year_month"`        // "2026-03"
	TotalDistance   float64 `json:"total_distance"`    // km
	TotalDuration   int     `json:"total_duration"`    // 秒
	RunCount        int     `json:"run_count"`
	AvgPaceSecPerKm float64 `json:"avg_pace_sec_per_km"` // 秒/km（フロントで "5:30" 表示用）
	AvgVDOT         float64 `json:"avg_vdot"`           // 月内の走行から算出した平均VDOT
	MaxVDOT         float64 `json:"max_vdot"`           // 月内の最高VDOT（走力の目安）
}

// AnalysisResponse は解析APIのレスポンス
type AnalysisResponse struct {
	MonthlyReports []MonthlyReport `json:"monthly_reports"`
	TotalDistance  float64         `json:"total_distance"`
	TotalDuration  int             `json:"total_duration"`
	TotalRunCount  int             `json:"total_run_count"`
}

// monthResult はGoroutineからChannelに送る1ヶ月分の集計
type monthResult struct {
	YearMonth        string
	TotalDistance    float64
	TotalDuration    int
	RunCount         int
	AvgPaceSecPerKm  float64
	AvgVDOT          float64
	MaxVDOT          float64
}

// AnalyzeByMonth は走行ログを月ごとに並列集計する（Goroutine + Channel）
func AnalyzeByMonth(logs []model.TrainingLog) AnalysisResponse {
	if len(logs) == 0 {
		return AnalysisResponse{MonthlyReports: []MonthlyReport{}}
	}

	// 月ごとにグループ化（key: "2006-01"）
	byMonth := make(map[string][]model.TrainingLog)
	for _, log := range logs {
		ym := log.TrainingDate.Format("2006-01")
		byMonth[ym] = append(byMonth[ym], log)
	}

	ch := make(chan monthResult, len(byMonth))
	var wg sync.WaitGroup
	wg.Add(len(byMonth))

	for yearMonth, list := range byMonth {
		// ループ変数をgoroutineに渡すためコピー
		ym := yearMonth
		group := make([]model.TrainingLog, len(list))
		copy(group, list)

		go func() {
			defer wg.Done()
			var dist float64
			var dur int
			var vdotSum float64
			var vdotCount int
			var maxVdot float64
			for _, t := range group {
				dist += t.Distance
				dur += t.Duration
				if t.Distance > 0 && t.Duration > 0 {
					v := CalculateVDOT(t.Distance*1000, float64(t.Duration))
					if v > 0 {
						vdotSum += v
						vdotCount++
						if v > maxVdot {
							maxVdot = v
						}
					}
				}
			}
			avgPace := 0.0
			if dist > 0 && dur > 0 {
				avgPace = float64(dur) / dist // 秒/km
			}
			avgVdot := 0.0
			if vdotCount > 0 {
				avgVdot = vdotSum / float64(vdotCount)
			}
			ch <- monthResult{
				YearMonth:        ym,
				TotalDistance:    dist,
				TotalDuration:    dur,
				RunCount:         len(group),
				AvgPaceSecPerKm:  avgPace,
				AvgVDOT:          avgVdot,
				MaxVDOT:          maxVdot,
			}
		}()
	}

	// 全goroutineの完了を待機してから結果を回収
	wg.Wait()
	close(ch)
	results := make([]monthResult, 0, len(byMonth))
	for r := range ch {
		results = append(results, r)
	}

	// 月順にソート
	sort.Slice(results, func(i, j int) bool {
		return results[i].YearMonth < results[j].YearMonth
	})

	// レスポンス構築
	reports := make([]MonthlyReport, len(results))
	var totalDist float64
	var totalDur int
	var totalCount int
	for i, r := range results {
		reports[i] = MonthlyReport{
			YearMonth:        r.YearMonth,
			TotalDistance:    r.TotalDistance,
			TotalDuration:    r.TotalDuration,
			RunCount:         r.RunCount,
			AvgPaceSecPerKm:  r.AvgPaceSecPerKm,
			AvgVDOT:          r.AvgVDOT,
			MaxVDOT:          r.MaxVDOT,
		}
		totalDist += r.TotalDistance
		totalDur += r.TotalDuration
		totalCount += r.RunCount
	}

	return AnalysisResponse{
		MonthlyReports: reports,
		TotalDistance:  totalDist,
		TotalDuration:  totalDur,
		TotalRunCount:  totalCount,
	}
}

// ThisMonthSummary は「今月」の集計だけを返す（ダッシュボード用）
func ThisMonthSummary(logs []model.TrainingLog) (distance float64, duration int, count int, avgPaceSecPerKm float64) {
	now := time.Now()
	thisMonth := now.Format("2006-01")
	for _, t := range logs {
		if t.TrainingDate.Format("2006-01") == thisMonth {
			distance += t.Distance
			duration += t.Duration
			count++
		}
	}
	if distance > 0 && duration > 0 {
		avgPaceSecPerKm = float64(duration) / distance
	}
	return
}
