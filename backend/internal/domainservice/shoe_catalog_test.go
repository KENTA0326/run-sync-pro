package domainservice

import (
	"testing"

	"github.com/KENTA0326/run-sync-pro/model"
)

func TestDistinctShoeBrands(t *testing.T) {
	t.Parallel()
	shoes := []model.Shoe{
		{Brand: "Asics"},
		{Brand: "  Nike "},
		{Brand: "Asics"},
		{Brand: ""},
		{Brand: "Nike"},
	}
	got := DistinctShoeBrands(shoes)
	want := []string{"Asics", "Nike"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
