//go:build !unix

package importio

import "os"

// 非 Unix ではファイルロックをスキップ（開発用サンプル。本番 Linux では flock が有効）。
func flockShared(*os.File) error { return nil }

func funlock(*os.File) error { return nil }
