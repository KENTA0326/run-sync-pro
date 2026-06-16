package validation

import (
	"testing"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/timeutil"
	"github.com/go-playground/validator/v10"
)

func TestRegisterCustom_caldate(t *testing.T) {
	v := validator.New()
	if err := registerCustom(v); err != nil {
		t.Fatal(err)
	}

	type row struct {
		D string `validate:"caldate"`
	}

	cases := []struct {
		name string
		in   string
		ok   bool
	}{
		{name: "valid", in: "2026-05-01", ok: true},
		{name: "invalid_month", in: "2026-13-01", ok: false},
		{name: "wrong_layout", in: "01-05-2026", ok: false},
		{name: "empty_skipped_by_rule", in: "", ok: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(row{D: tc.in})
			got := err == nil
			if got != tc.ok {
				t.Fatalf("ok=%v want %v err=%v", got, tc.ok, err)
			}
		})
	}
}

func TestRegisterCustom_caldate_not_future(t *testing.T) {
	v := validator.New()
	if err := registerCustom(v); err != nil {
		t.Fatal(err)
	}

	type row struct {
		D string `validate:"caldate,caldate_not_future"`
	}

	today := timeutil.NowJST()
	okDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, timeutil.JST)
	future := okDay.AddDate(0, 0, 1)

	err := v.Struct(row{D: okDay.Format("2006-01-02")})
	if err != nil {
		t.Fatalf("today should pass: %v", err)
	}
	err = v.Struct(row{D: future.Format("2006-01-02")})
	if err == nil {
		t.Fatal("future date should fail")
	}
}
