package service

import "math"

type SplitRow struct {
	// Km は通過距離（1kmごと。最後だけ 42.195km を Finish として返す）
	Km float64 `json:"km"`
	// Label は表示用ラベル（例: "1km", "Finish"）
	Label string `json:"label"`
	// CumulativeSeconds はスタートからの累積秒
	CumulativeSeconds int `json:"cumulative_seconds"`
}

// GenerateFullMarathonSplits は 1km ごとの通過タイム + Finish(42.195km) を生成する
// paceSecPerKm は 1km あたりの秒数
func GenerateFullMarathonSplits(paceSecPerKm float64) []SplitRow {
	if paceSecPerKm <= 0 {
		return nil
	}

	rows := make([]SplitRow, 0, 43)

	// 1km〜42km
	for km := 1; km <= 42; km++ {
		sec := int(math.Round(float64(km) * paceSecPerKm))
		rows = append(rows, SplitRow{
			Km:                float64(km),
			Label:             formatKmLabel(float64(km)),
			CumulativeSeconds: sec,
		})
	}

	// Finish: 42.195km
	finishKm := 42.195
	finishSec := int(math.Round(finishKm * paceSecPerKm))
	rows = append(rows, SplitRow{
		Km:                finishKm,
		Label:             "Finish",
		CumulativeSeconds: finishSec,
	})

	return rows
}

func formatKmLabel(km float64) string {
	if km == math.Trunc(km) {
		return fmtInt(int(km)) + "km"
	}
	// 小数は現状使わないが一応
	return fmtFloat(km) + "km"
}

func fmtInt(v int) string {
	// strconv.Itoa をラップして依存を増やさない
	if v == 0 {
		return "0"
	}
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	buf := make([]byte, 0, 12)
	for v > 0 {
		buf = append(buf, byte('0'+v%10))
		v /= 10
	}
	// reverse
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return sign + string(buf)
}

func fmtFloat(v float64) string {
	// 最小限: 3桁小数まで
	iv := int(math.Trunc(v))
	frac := v - float64(iv)
	if frac < 0 {
		frac = -frac
	}
	milli := int(math.Round(frac * 1000))
	if milli == 0 {
		return fmtInt(iv)
	}
	// 末尾0を落とす
	s := fmtInt(milli)
	for len(s) < 3 {
		s = "0" + s
	}
	for len(s) > 0 && s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}
	return fmtInt(iv) + "." + s
}
