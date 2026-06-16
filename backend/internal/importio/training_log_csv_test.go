package importio

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestStreamTrainingLogCSV(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "logs.csv")
	content := "training_date,distance,duration,pace,kind,shoe_id,memo\n" +
		"2026-04-22,10.5,3600,6:00,0,1,朝ラン\n" +
		"2026-04-23,5,1800,6:30,1,1,\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	var rows []TrainingLogCSVRecord
	n, err := StreamTrainingLogCSV(context.Background(), path, 100, func(r TrainingLogCSVRecord) error {
		rows = append(rows, r)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 || len(rows) != 2 {
		t.Fatalf("got n=%d rows=%d", n, len(rows))
	}
	if rows[0].Memo != "朝ラン" || rows[0].Distance != 10.5 {
		t.Fatalf("unexpected first row: %+v", rows[0])
	}
}
