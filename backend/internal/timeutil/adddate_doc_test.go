package timeutil_test

import (
	"testing"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/timeutil"
)

// Go の AddDate は「暦の年月日」をそれぞれ加算したうえで正規化する。
// 1/31 に 1 ヶ月足しても「2 月末へクリップ」にはならず、存在しない日へ進んだぶん翌月へ繰り上がる。
// 「月末の翌月も末端でそろえる」要件には、そのまま AddDate を使わず別設計が必要になる。
func TestAddDateJan31PlusOneMonthGoSemantics(t *testing.T) {
	t.Parallel()
	jst := timeutil.JST

	cases := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{
			name: "2025 non-leap Feb has 28 days: Jan31 + 1mo overflows into March",
			in:   time.Date(2025, 1, 31, 0, 0, 0, 0, jst),
			want: time.Date(2025, 3, 3, 0, 0, 0, 0, jst),
		},
		{
			name: "2024 leap Feb has 29 days: Jan31 + 1mo overflows by 2 days",
			in:   time.Date(2024, 1, 31, 0, 0, 0, 0, jst),
			want: time.Date(2024, 3, 2, 0, 0, 0, 0, jst),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.in.AddDate(0, 1, 0)
			if !got.Equal(tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestAddMonthsClampEndOfMonth(t *testing.T) {
	t.Parallel()
	jst := timeutil.JST

	cases := []struct {
		name   string
		in     time.Time
		months int
		want   time.Time
	}{
		{
			name:   "2025-01-31 +1m => 2025-02-28",
			in:     time.Date(2025, 1, 31, 9, 30, 0, 0, jst),
			months: 1,
			want:   time.Date(2025, 2, 28, 9, 30, 0, 0, jst),
		},
		{
			name:   "2024-01-31 +1m => 2024-02-29 (leap year)",
			in:     time.Date(2024, 1, 31, 9, 30, 0, 0, jst),
			months: 1,
			want:   time.Date(2024, 2, 29, 9, 30, 0, 0, jst),
		},
		{
			name:   "2025-03-30 -1m => 2025-02-28",
			in:     time.Date(2025, 3, 30, 9, 30, 0, 0, jst),
			months: -1,
			want:   time.Date(2025, 2, 28, 9, 30, 0, 0, jst),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := timeutil.AddMonthsClampEndOfMonth(tc.in, tc.months)
			if !got.Equal(tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
