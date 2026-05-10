// Package strconcat は文字列結合の実装パターンとコストの違いをベンチマークで比較するためのパッケージである。
//
// 大量結合では strings.Builder（必要なら Grow で容量を先取り）か strings.Join を選ぶのが無難。
// + や fmt.Sprintf をループで繰り返すと、都度新しい文字列が割り当てられコピーも累積しやすい。
//
// ベンチ実行例:
//
//	go test -bench=. -benchmem ./internal/strconcat/
package strconcat
