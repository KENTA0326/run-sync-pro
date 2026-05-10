package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// fixtureDir はこのファイルから見た testdata ディレクトリ。
func fixtureDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("testutil.fixtureDir: runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "testdata")
}

// LoadFixture は internal/testutil/testdata から JSON を読み、dst にデコードする。
// rel は "training_logs/two_months.json" のような testdata からの相対パス。
func LoadFixture(t *testing.T, rel string, dst any) {
	t.Helper()
	path := filepath.Join(fixtureDir(), filepath.FromSlash(rel))
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		t.Fatalf("decode fixture %s: %v", path, err)
	}
}
