package importio

import (
	"os"
	"strings"
	"testing"
)

func TestSpoolToTempFile_removes_on_success(t *testing.T) {
	t.Parallel()
	path, cleanup, err := SpoolToTempFile(strings.NewReader("hello"), 1024, "test-spool-*")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("temp should exist before cleanup: %v", err)
	}
	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("temp should be removed after cleanup: %v", err)
	}
}

func TestSpoolToTempFile_rejects_oversize(t *testing.T) {
	t.Parallel()
	_, cleanup, err := SpoolToTempFile(strings.NewReader("12345"), 4, "test-spool-*")
	if cleanup != nil {
		cleanup()
	}
	if err == nil {
		t.Fatal("expected oversize error")
	}
}
