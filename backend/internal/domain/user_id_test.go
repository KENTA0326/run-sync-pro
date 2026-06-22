package domain

import (
	"context"
	"net/http"
	"testing"
)

func TestParseUserID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		raw   any
		want  UserID
		ok    bool
		valid bool
	}{
		{"uint", uint(42), 42, true, true},
		{"UserID", UserID(7), 7, true, true},
		{"float64", float64(3), 3, true, true},
		{"negative int", int(-1), 0, false, false},
		{"negative float64", float64(-1), 0, false, false},
		{"zero uint invalid", uint(0), 0, true, false},
		{"string", "1", 0, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := ParseUserID(tt.raw)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("id = %v, want %v", got, tt.want)
			}
			if got.IsValid() != tt.valid {
				t.Fatalf("IsValid() = %v, want %v", got.IsValid(), tt.valid)
			}
		})
	}
}

func TestUserIDFromStdContext(t *testing.T) {
	t.Parallel()
	ctx := WithUserID(context.Background(), 9)
	id, ok := UserIDFromStdContext(ctx)
	if !ok || id != 9 {
		t.Fatalf("got (%v, %v), want (9, true)", id, ok)
	}
	_, ok = UserIDFromStdContext(context.Background())
	if ok {
		t.Fatal("expected false for empty context")
	}
}

func TestUserIDFromRequest(t *testing.T) {
	t.Parallel()
	req := &http.Request{}
	req = req.WithContext(WithUserID(context.Background(), 12))
	id, ok := UserIDFromRequest(req)
	if !ok || id != 12 {
		t.Fatalf("got (%v, %v), want (12, true)", id, ok)
	}
}

type fakeContext struct {
	val any
	ok  bool
}

func (f fakeContext) Get(string) (any, bool) { return f.val, f.ok }

func TestUserIDFromContext_legacyKeyStore(t *testing.T) {
	t.Parallel()
	id, ok := UserIDFromContext(fakeContext{val: UserID(9), ok: true})
	if !ok || id != 9 {
		t.Fatalf("got (%v, %v), want (9, true)", id, ok)
	}
	_, ok = UserIDFromContext(fakeContext{ok: false})
	if ok {
		t.Fatal("expected false for missing context value")
	}
}
