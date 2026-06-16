package handler

import "testing"

func TestListPaginationQuery_normalize(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		q          listPaginationQuery
		wantLimit  int
		wantOffset int
	}{
		{name: "defaults", q: listPaginationQuery{}, wantLimit: 50, wantOffset: 0},
		{name: "explicit_limit", q: listPaginationQuery{Limit: 10}, wantLimit: 10, wantOffset: 0},
		{name: "explicit_offset", q: listPaginationQuery{Offset: 20}, wantLimit: 50, wantOffset: 20},
		{name: "over_max_clamped", q: listPaginationQuery{Limit: 999}, wantLimit: 500, wantOffset: 0},
		{name: "negative_offset_zeroed", q: listPaginationQuery{Offset: -1}, wantLimit: 50, wantOffset: 0},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			limit, offset := tc.q.normalize()
			if limit != tc.wantLimit || offset != tc.wantOffset {
				t.Fatalf("normalize() = (%d, %d) want (%d, %d)", limit, offset, tc.wantLimit, tc.wantOffset)
			}
		})
	}
}
