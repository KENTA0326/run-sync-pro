package domainservice

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

type analyzerStd struct{}

// NewAnalyzer は本番用の月別解析の具体実装を返す。
func NewAnalyzer() *analyzerStd {
	return &analyzerStd{}
}

func (analyzerStd) AnalyzeByMonth(logs []model.TrainingLog) model.AnalysisResponse {
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
func AnalyzeByMonth(logs []model.TrainingLog) model.AnalysisResponse {
	if len(logs) == 0 {
		return model.AnalysisResponse{MonthlyReports: []model.MonthlyReport{}}
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

	wg.Wait() // 全ワーカ終了まで待つ（送信完了を保証してからチャネルを閉じる）
	close(ch)
	results := make([]monthResult, 0, len(byMonth))
	for r := range ch {
		results = append(results, r)
	}

	// 月順にソート
	sort.Slice(results, func(i, j int) bool {
		return results[i].YearMonth < results[j].YearMonth
	})

	// レスポンス構築（monthResult → MonthlyReport は型変換のみなので MapSlice + any 制約）
	reports := MapSlice(results, func(r monthResult) model.MonthlyReport { return model.MonthlyReport(r) })
	var totalDist float64
	var totalDur int
	var totalCount int
	for _, r := range results {
		totalDist += r.TotalDistance
		totalDur += r.TotalDuration
		totalCount += r.RunCount
	}

	return model.AnalysisResponse{
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
