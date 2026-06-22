package domain

import "testing"

func TestResourceIsValid(t *testing.T) {
	t.Parallel()
	if !ResourceShoe.IsValid() {
		t.Fatal("ResourceShoe should be valid")
	}
	if Resource("typo").IsValid() {
		t.Fatal("unknown resource should be invalid")
	}
}

func TestActionIsValid(t *testing.T) {
	t.Parallel()
	if !ActionRead.IsValid() {
		t.Fatal("ActionRead should be valid")
	}
	if Action("delete").IsValid() {
		t.Fatal("unknown action should be invalid")
	}
}
