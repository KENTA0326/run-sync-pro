package domainservice

import (
	"runtime"
	"testing"
	"time"

	"github.com/KENTA0326/run-sync-pro/model"
)

// ベンチマークと GC の読み方（要約）
//
// • b.ReportAllocs(): 1 演算あたりのアロケーション回数・バイトを出す。ヒープ負荷と GC トリガー頻度の目安になる。
// • b.ResetTimer(): 入力データ構築など「計測したくないセットアップ」をループの外に置いたあと呼ぶ。
//   ループ内で毎回 make([]T, n) するとアロケーション自体が支配的になり、AnalyzeByMonth 本体のコストが見えにくい。
// • Go の GC は並列マーク・短い STW などでヒープを回収する。オールケーションが多いと GC が増え、
//   レイテンシ分布では尾（p99 など）が伸びやすい。ベンチは主にスループットとオールocations/ns を見る一方、
//   実サービスでは trace (/debug/pprof) やレイテンシヒストグラムも併用するとよい。
// • 実行例:
//
//	go test -bench=BenchmarkAnalyzeByMonth -benchmem -count=5 ./service/
//
// • 複数回を統計したい場合は golang.org/x/perf/cmd/benchstat などで比較する。

func makeBenchTrainingLogs(n int) []model.TrainingLog {
	base := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	logs := make([]model.TrainingLog, n)
	for i := range logs {
		logs[i] = model.TrainingLog{
			TrainingDate: base.Add(time.Duration(i) * time.Hour),
			Distance:     5,
			Duration:     1800,
		}
	}
	return logs
}

func BenchmarkAnalyzeByMonth(b *testing.B) {
	sizes := []struct {
		name string
		n    int
	}{
		{"serial_below_threshold_256", 256},
		{"near_threshold_511", 511},
		{"parallel_768", 768},
		{"parallel_large_4096", 4096},
	}
	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			logs := makeBenchTrainingLogs(sz.n)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = AnalyzeByMonth(logs)
			}
		})
	}
}

// BenchmarkAnalyzeByMonth_afterGC は「セットアップ後にヒープを一度きれいにしたあと計測を開始する」例。
// 初回の nursery 盛り上がりを少し抑えたスループットになりやすいが、実運用のレイテンシをそのまま
// 再現するものではない（あくまで比較・傾向確認用）。
func BenchmarkAnalyzeByMonth_afterGC(b *testing.B) {
	logs := makeBenchTrainingLogs(2048)
	runtime.GC()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AnalyzeByMonth(logs)
	}
}
