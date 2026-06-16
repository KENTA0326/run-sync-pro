package importio

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestStreamWriteTrainingLogCSV(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	n, err := StreamWriteTrainingLogCSV(context.Background(), &buf, func(emit func(TrainingLogCSVRecord) error) error {
		return emit(TrainingLogCSVRecord{
			TrainingDate: "2026-04-22",
			Distance:     10,
			Duration:     3600,
			Pace:         "6:00",
			Kind:         0,
			ShoeID:       1,
			Memo:         "test",
		})
	})
	if err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "training_date,distance,duration,pace,kind,shoe_id,memo\n") {
		t.Fatalf("unexpected header: %q", out)
	}
	if !strings.Contains(out, "2026-04-22,10,3600,6:00,0,1,test") {
		t.Fatalf("unexpected body: %q", out)
	}
}
