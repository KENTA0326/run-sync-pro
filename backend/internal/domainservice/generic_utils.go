package domainservice

// MapSlice は入力スライス in を要素ごとに変換し、新しいスライスを返す。
//
// 型制約に any を使う理由:
//   - T は「入力要素の型」、R は「出力要素の型」で、どちらも比較可能である必要はない。
//   - 必要なのは mapper が T -> R の変換関数として呼べることだけなので、最小制約の any を使う。
func MapSlice[T any, R any](in []T, mapper func(T) R) []R {
	if len(in) == 0 {
		return []R{}
	}

	out := make([]R, 0, len(in))
	for _, v := range in {
		out = append(out, mapper(v))
	}
	return out
}

// Unique は重複を除いた要素を、最初に出現した順序で返す。
//
// 型制約に comparable を使う理由:
//   - 重複判定に map のキーとして T を利用するため、T は比較可能である必要がある。
//   - そのため any ではなく comparable を明示する。
//
// 単一ゴルーチンからのみ map に触れるため mutex は不要。存在確認は v, ok := m[key] を使う。
func Unique[T comparable](in []T) []T {
	if len(in) == 0 {
		return []T{}
	}

	seen := make(map[T]struct{}, len(in))
	out := make([]T, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
