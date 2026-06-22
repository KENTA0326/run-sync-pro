package importio

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStreamTrainingLogCSV_respectsContextCancel(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "logs.csv")
	lines := "training_date,distance,duration,pace,kind,shoe_id\n"
	for i := 0; i < 50; i++ {
		lines += "2026-04-22,10,3600,6:00,0,1\n"
	}
	if err := os.WriteFile(path, []byte(lines), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := StreamTrainingLogCSV(ctx, path, 100, func(TrainingLogCSVRecord) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
}

func TestStreamTrainingLogCSV_nilContextUsesBackground(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "one.csv")
	content := "training_date,distance,duration,pace,kind,shoe_id\n2026-04-22,5,1800,6:00,0,1\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	n, err := StreamTrainingLogCSV(ctx, path, 10, func(TrainingLogCSVRecord) error { return nil })
	if err != nil || n != 1 {
		t.Fatalf("got n=%d err=%v", n, err)
	}
}
