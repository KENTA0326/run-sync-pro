package timeutil

import "time"

// AddMonthsClampEndOfMonth は「月を加算した結果の日が存在しない場合、対象月の末日に丸める」。
// 例: 2025-01-31 + 1 month => 2025-02-28, 2024-01-31 + 1 month => 2024-02-29
//
// AddDate(0, months, 0) のように翌月へ繰り上がる挙動が要件に合わない場面
// （請求日・締め日など）で使う。
func AddMonthsClampEndOfMonth(t time.Time, months int) time.Time {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	nsec := t.Nanosecond()
	loc := t.Location()

	// 対象月の1日へ移動
	base := time.Date(year, month, 1, hour, min, sec, nsec, loc).AddDate(0, months, 0)
	targetYear, targetMonth, _ := base.Date()

	lastDay := daysInMonth(targetYear, targetMonth, loc)
	if day > lastDay {
		day = lastDay
	}
	return time.Date(targetYear, targetMonth, day, hour, min, sec, nsec, loc)
}

func daysInMonth(year int, month time.Month, loc *time.Location) int {
	// 翌月0日 = 当月末日
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}
