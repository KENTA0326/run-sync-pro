package strconcat

import (
	"fmt"
	"strings"
	"testing"
)

const benchNumParts = 512

const benchPart = "abcdefghij"

func benchParts() []string {
	ps := make([]string, benchNumParts)
	for i := range ps {
		ps[i] = benchPart
	}
	return ps
}

// BenchmarkConcatPlusLoop: += はループごとに新しいバッファを確保し既存内容をコピーするため、
// 文字列が長くなるほど 1 回あたりのコピー量が増え、だいたい二乗に近い総コピーになりやすい。
func BenchmarkConcatPlusLoop(b *testing.B) {
	ps := benchParts()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s string
		for _, p := range ps {
			s += p
		}
		_ = s
	}
}

// BenchmarkConcatSprintfLoop: fmt は書式解析と中間表現のコストがあり、ループ内では特に重くなりやすい。
func BenchmarkConcatSprintfLoop(b *testing.B) {
	ps := benchParts()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := ""
		for _, p := range ps {
			s = fmt.Sprintf("%s%s", s, p)
		}
		_ = s
	}
}

// BenchmarkConcatStringsBuilder: 単一バッファへ追記するのでコピーが線形に抑えられやすい。
func BenchmarkConcatStringsBuilder(b *testing.B) {
	ps := benchParts()
	wantLen := benchNumParts * len(benchPart)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sb strings.Builder
		sb.Grow(wantLen)
		for _, p := range ps {
			sb.WriteString(p)
		}
		_ = sb.String()
	}
}

// BenchmarkConcatStringsBuilderNoGrow: Grow なし版。必要長が分かるときは Grow でリアロケーションを減らせる。
func BenchmarkConcatStringsBuilderNoGrow(b *testing.B) {
	ps := benchParts()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sb strings.Builder
		for _, p := range ps {
			sb.WriteString(p)
		}
		_ = sb.String()
	}
}

// BenchmarkConcatStringsJoin: スライスが既にある場合の標準ライブラリ実装（単一結合に強い）。
func BenchmarkConcatStringsJoin(b *testing.B) {
	ps := benchParts()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Join(ps, "")
	}
}
