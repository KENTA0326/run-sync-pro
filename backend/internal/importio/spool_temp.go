// Package importio は大容量アップロードの一時ファイル退避とストリーム読み取りのユーティリティを提供する。
package importio

import (
	"fmt"
	"io"
	"os"
)

// DefaultMaxSpoolBytes は 1 リクエストでディスクに退避する上限（サンプル実装用）。
const DefaultMaxSpoolBytes = 32 << 20 // 32 MiB

// SpoolToTempFile は r の内容を一時ファイルへコピーする。
// 呼び出し側は返却された cleanup を defer で必ず実行すること（成功・失敗どちらでも削除）。
func SpoolToTempFile(r io.Reader, maxBytes int64, pattern string) (path string, cleanup func(), err error) {
	if maxBytes <= 0 {
		return "", func() {}, fmt.Errorf("importio: maxBytes must be positive")
	}
	if pattern == "" {
		pattern = "importio-spool-*"
	}

	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", func() {}, fmt.Errorf("importio: create temp: %w", err)
	}
	path = f.Name()
	cleanup = func() { _ = os.Remove(path) }

	limited := io.LimitReader(r, maxBytes+1)
	written, copyErr := io.Copy(f, limited)
	closeErr := f.Close()
	if copyErr != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("importio: spool copy: %w", copyErr)
	}
	if closeErr != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("importio: spool close: %w", closeErr)
	}
	if written > maxBytes {
		cleanup()
		return "", func() {}, fmt.Errorf("importio: payload exceeds %d bytes", maxBytes)
	}
	return path, cleanup, nil
}
