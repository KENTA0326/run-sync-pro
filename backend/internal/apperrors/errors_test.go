package apperrors_test

import (
	"errors"
	"testing"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
)

func TestVisibleExtract_table(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		err      error
		wantHTTP int
		wantMsg  string
	}{
		{
			name:     "bad_request",
			err:      apperrors.BadRequest("bad"),
			wantHTTP: 400,
			wantMsg:  "bad",
		},
		{
			name:     "unauthorized_msg",
			err:      apperrors.UnauthorizedMsg("ログインしてください"),
			wantHTTP: 401,
			wantMsg:  "ログインしてください",
		},
		{
			name:     "not_found_msg",
			err:      apperrors.NotFoundMsg("無い"),
			wantHTTP: 404,
			wantMsg:  "無い",
		},
		{
			name:     "conflict_msg",
			err:      apperrors.ConflictMsg("競合"),
			wantHTTP: 409,
			wantMsg:  "競合",
		},
		{
			name:     "internal_msg",
			err:      apperrors.InternalMsg("内部エラー", errors.New("cause")),
			wantHTTP: 500,
			wantMsg:  "内部エラー",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var v *apperrors.Visible
			if !errors.As(tc.err, &v) {
				t.Fatal("errors.As Visible failed")
			}
			if v.HTTP != tc.wantHTTP {
				t.Fatalf("HTTP=%d want %d", v.HTTP, tc.wantHTTP)
			}
			if v.Msg != tc.wantMsg {
				t.Fatalf("Msg=%q want %q", v.Msg, tc.wantMsg)
			}
		})
	}
}

func TestSentinelIs_table(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		target error
		want bool
	}{
		{
			name:   "not_found_msg_wraps_sentinel",
			err:    apperrors.NotFoundMsg("x"),
			target: apperrors.ErrNotFound,
			want:   true,
		},
		{
			name:   "bad_request_wraps_invalid_input",
			err:    apperrors.BadRequest("x"),
			target: apperrors.ErrInvalidInput,
			want:   true,
		},
		{
			name:   "unauthorized_msg_wraps_sentinel",
			err:    apperrors.UnauthorizedMsg("x"),
			target: apperrors.ErrUnauthorized,
			want:   true,
		},
		{
			name:   "internal_wraps_internal_sentinel",
			err:    apperrors.InternalMsg("x", errors.New("y")),
			target: apperrors.ErrInternal,
			want:   true,
		},
		{
			name:   "opaque_not_found",
			err:    errors.New("plain"),
			target: apperrors.ErrNotFound,
			want:   false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := errors.Is(tc.err, tc.target)
			if got != tc.want {
				t.Fatalf("errors.Is(%v, %v)=%v want %v", tc.err, tc.target, got, tc.want)
			}
		})
	}
}
