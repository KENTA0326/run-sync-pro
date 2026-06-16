package service

import "testing"

func TestHashResetTokenDeterministic(t *testing.T) {
	t.Parallel()
	a := hashResetToken("abc")
	b := hashResetToken("abc")
	if a != b || len(a) != 64 {
		t.Fatalf("unexpected hash: %q", a)
	}
}

func TestBuildPasswordResetURL(t *testing.T) {
	t.Parallel()
	got := BuildPasswordResetURL("http://localhost:3001/", "tok123")
	want := "http://localhost:3001/reset-password?token=tok123"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
