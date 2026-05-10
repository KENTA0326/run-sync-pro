package service

import (
	"runtime"
	"sort"
	"sync"

	"github.com/KENTA0326/run-sync-pro/internal/timeutil"
	"github.com/KENTA0326/run-sync-pro/model"
)

const (
	groupByMonthParallelThreshold = 512
	groupByMonthMaxWorkers        = 8
)

//
// ── 並行処理メモ（go func と sync.WaitGroup） ───────────────────────────────
//
// • groupLogsByMonth: 大量ログ時にシャードごとに go func を起動し wg.Add/defer wg.Done/wg.Wait で合流する。
//
// • AnalyzeByMonth  : 月キー単位で go func を起動し、同様に WaitGroup で揃えたあとチャネルを閉じて回収する。
//
// ゴルーチン関連で起きやすい問題:
//
// • Add と Done が対応しない → Wait が永遠ブロックしたり早期に抜けたりしてデッドロック・誤結果になる。
// • defer wg.Done() をしない分岐がある →同上。ワーカ入口で defer が無難。
// • go func がループ変数をクロージャでキャプチャ → 競合。ym / slice コピーでそのイテレーションの値だけを渡す。
// • 送信がブロックしたまま誰も受信しない → ゴルーチンが終わらずリークし得る。ここではバッファ len(byMonth) と
//   wg.Wait の後での close/channel drain で「送り側が済むまで親が読む」を満たす。
//
// 月キーは timeutil.MonthKey で JST に換算してから "YYYY-MM" にしている（DB の Asia/Tokyo と語義を合わせる）。

// MonthlyReport は1ヶ月分の集計結果
type MonthlyReport struct {
	YearMonth       string  `json:"year_month"`     // "2026-03"
	TotalDistance   float64 `json:"total_distance"` // km
	TotalDuration   int     `json:"total_duration"` // 秒
	RunCount        int     `json:"run_count"`
	AvgPaceSecPerKm float64 `json:"avg_pace_sec_per_km"` // 秒/km（フロントで "5:30" 表示用）
	AvgVDOT         float64 `json:"avg_vdot"`            // 月内の走行から算出した平均VDOT
	MaxVDOT         float64 `json:"max_vdot"`            // 月内の最高VDOT（走力の目安）
}

// AnalysisResponse は解析APIのレスポンス
type AnalysisResponse struct {
	MonthlyReports []MonthlyReport `json:"monthly_reports"`
	TotalDistance  float64         `json:"total_distance"`
	TotalDuration  int             `json:"total_duration"`
	TotalRunCount  int             `json:"total_run_count"`
}

type analyzerStd struct{}

// NewAnalyzer は本番用の月別解析の具体実装を返す。
func NewAnalyzer() *analyzerStd {
	return &analyzerStd{}
}

func (analyzerStd) AnalyzeByMonth(logs []model.TrainingLog) AnalysisResponse {
	return AnalyzeByMonth(logs)
}

// monthResult はGoroutineからChannelに送る1ヶ月分の集計
type monthResult struct {
	YearMonth       string
	TotalDistance   float64
	TotalDuration   int
	RunCount        int
	AvgPaceSecPerKm float64
	AvgVDOT         float64
	MaxVDOT         float64
}

// groupLogsByMonth はログを年月キーでまとめる。
// 件数が少ないときは単一ゴルーチンで map に触れ、mutex 不要。
// 件数が多いときは go func + sync.WaitGroup（Add / defer Done / Wait）でシャード処理し、
// ワーカーごとにローカル map を作り、共有 map へは sync.Mutex で直列マージする。
// 複数ゴルーチンから同じ map を同時に読み書きするとランタイムが fatal になるため、共有 map への書き込みは必ず保護する。
func groupLogsByMonth(logs []model.TrainingLog) map[string][]model.TrainingLog {
	if len(logs) < groupByMonthParallelThreshold {
		byMonth := make(map[string][]model.TrainingLog)
		for _, log := range logs {
			ym := timeutil.MonthKey(log.TrainingDate)
			byMonth[ym] = append(byMonth[ym], log)
		}
		return byMonth
	}

	workers := min(groupByMonthMaxWorkers, runtime.NumCPU(), len(logs))
	if workers < 2 {
		byMonth := make(map[string][]model.TrainingLog)
		for _, log := range logs {
			ym := timeutil.MonthKey(log.TrainingDate)
			byMonth[ym] = append(byMonth[ym], log)
		}
		return byMonth
	}

	chunk := (len(logs) + workers - 1) / workers
	byMonth := make(map[string][]model.TrainingLog)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		start := w * chunk
		if start >= len(logs) {
			break
		}
		end := start + chunk
		if end > len(logs) {
			end = len(logs)
		}
		part := logs[start:end]
		wg.Add(1) // Done は常にこの go func の defer で呼ぶ（呼び漏れ防止）
		go func(chunk []model.TrainingLog) {
			defer wg.Done()
			local := make(map[string][]model.TrainingLog)
			for _, log := range chunk {
				ym := timeutil.MonthKey(log.TrainingDate)
				local[ym] = append(local[ym], log)
			}
			mu.Lock()
			for ym, list := range local {
				if existing, ok := byMonth[ym]; ok {
					byMonth[ym] = append(existing, list...)
				} else {
					byMonth[ym] = list
				}
			}
			mu.Unlock()
		}(part)
	}
	wg.Wait()
	return byMonth
}

// AnalyzeByMonth は走行ログを月ごとに並列集計する（go func + sync.WaitGroup + バッファチャネル）。
func AnalyzeByMonth(logs []model.TrainingLog) AnalysisResponse {
	if len(logs) == 0 {
		return AnalysisResponse{MonthlyReports: []MonthlyReport{}}
	}

	byMonth := groupLogsByMonth(logs)

	ch := make(chan monthResult, len(byMonth)) // 各ワーカが一度ずつ送信するので詰まりにくくする（送信ブロックでのリーク回避）
	var wg sync.WaitGroup
	wg.Add(len(byMonth))

	for yearMonth, list := range byMonth {
		// ループ変数をgoroutineに渡すためコピー
		ym := yearMonth
		group := make([]model.TrainingLog, len(list))
		copy(group, list)

		go func() {
			defer wg.Done() // パニック時も Done させるために defer
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
				YearMonth:       ym,
				TotalDistance:   dist,
				TotalDuration:   dur,
				RunCount:        len(group),
				AvgPaceSecPerKm: avgPace,
				AvgVDOT:         avgVdot,
				MaxVDOT:         maxVdot,
			}
		}()
	}

	wg.Wait()   // 全ワーカ終了まで待つ（送信完了を保証してからチャネルを閉じる）
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
			YearMonth:       r.YearMonth,
			TotalDistance:   r.TotalDistance,
			TotalDuration:   r.TotalDuration,
			RunCount:        r.RunCount,
			AvgPaceSecPerKm: r.AvgPaceSecPerKm,
			AvgVDOT:         r.AvgVDOT,
			MaxVDOT:         r.MaxVDOT,
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
	now := timeutil.NowJST()
	thisMonth := now.Format("2006-01")
	for _, t := range logs {
		if timeutil.MonthKey(t.TrainingDate) == thisMonth {
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
