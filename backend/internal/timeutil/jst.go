// Package timeutil はカレンダー日・「日本の今月」判定など、アプリ全体でタイムゾーンを揃える。
//
// Go の time.Parse はゾーン無し日付を UTC の深夜として解釈するため、
// 「走った日」という語義では JST の midnight とずれることがある。ParseCalendarDate で統一する。
package timeutil

import "time"

// JST は DB の TimeZone=Asia/Tokyo と揃えたアプリ標準タイムゾーン。
// IANA データが無い環境では UTC+9 の固定オフセットにフォールバックする。
var JST = mustJST()

func mustJST() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.FixedZone("JST", 9*3600)
	}
	return loc
}

// ParseCalendarDate は "YYYY-MM-DD" を JST のその日 00:00:00 として解釈する。
// time.Parse の UTC 深夜既定による「カレンダー日がずれる」リスクを避ける。
func ParseCalendarDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, JST)
}

// MonthKey は瞬間 t を JST に換算したうえで "YYYY-MM" を返す（月別集計のキー用）。
func MonthKey(t time.Time) string {
	return t.In(JST).Format("2006-01")
}

// NowJST は現在時刻を JST で表した値（「日本の今日」「今月」の判定用）。
func NowJST() time.Time {
	return time.Now().In(JST)
}
