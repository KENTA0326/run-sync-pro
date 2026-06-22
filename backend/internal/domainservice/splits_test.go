package domainservice

import (
	"testing"
)

func TestGenerateFullMarathonSplits_knownPace(t *testing.T) {
	t.Parallel()

	rows := GenerateFullMarathonSplits(300) // 5:00/km
	if len(rows) != 43 {
		t.Fatalf("len=%d want 43 (1-42km + Finish)", len(rows))
	}
	if rows[0].Km != 1 || rows[0].Label != "1km" || rows[0].CumulativeSeconds != 300 {
		t.Fatalf("km1: %+v", rows[0])
	}
	if rows[41].Km != 42 || rows[41].CumulativeSeconds != 12600 {
		t.Fatalf("km42: %+v", rows[41])
	}
	finish := rows[42]
	if finish.Km != 42.195 || finish.Label != "Finish" || finish.CumulativeSeconds != 12659 {
		t.Fatalf("finish: %+v", finish)
	}
}

func TestGenerateFullMarathonSplits_invalidPace(t *testing.T) {
	t.Parallel()

	if got := GenerateFullMarathonSplits(0); got != nil {
		t.Fatalf("expected nil, got len=%d", len(got))
	}
	if got := GenerateFullMarathonSplits(-1); got != nil {
		t.Fatalf("expected nil, got len=%d", len(got))
	}
}

func TestGenerateFullMarathonSplits_cumulativeIncreases(t *testing.T) {
	t.Parallel()

	rows := GenerateFullMarathonSplits(270)
	prev := 0
	for i, row := range rows {
		if row.CumulativeSeconds < prev {
			t.Fatalf("row[%d] cumulative decreased: %d -> %d", i, prev, row.CumulativeSeconds)
		}
		prev = row.CumulativeSeconds
	}
}

func TestMarathonSplitsWrapper(t *testing.T) {
	t.Parallel()

	s := NewMarathonSplits()
	rows := s.GenerateFullMarathonSplits(300)
	if len(rows) != 43 {
		t.Fatalf("len=%d want 43", len(rows))
	}
}
