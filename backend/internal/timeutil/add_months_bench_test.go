package timeutil

import (
	"testing"
	"time"
)

// 軽量関数のベンチ例。b.ResetTimer より前に入力時刻だけ固定しておき、
// 「同一入力を繰り返し処理したときの ns/op」とアロケーションを読む。

func BenchmarkAddMonthsClampEndOfMonth(b *testing.B) {
	in := time.Date(2025, 1, 31, 9, 30, 0, 0, JST)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AddMonthsClampEndOfMonth(in, 1)
	}
}
