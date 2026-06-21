package domainservice

import (
	"strconv"
	"testing"
)

func TestMapSlice(t *testing.T) {
	t.Parallel()
	got := MapSlice([]int{1, 2, 3}, func(n int) string { return strconv.Itoa(n * 10) })
	want := []string{"10", "20", "30"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if empty := MapSlice([]int(nil), func(n int) int { return n }); len(empty) != 0 {
		t.Fatalf("nil slice should yield empty slice, got len %d", len(empty))
	}
}

func TestUnique(t *testing.T) {
	t.Parallel()
	got := Unique([]string{"a", "b", "a", "c", "b"})
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
