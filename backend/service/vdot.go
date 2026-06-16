package service

import (
	"math"

	"github.com/KENTA0326/run-sync-pro/model"
)

// VDOT計算式（Jack Daniels' Running Formula に基づく）
// 参考文献: VDOT = (−4.60 + 0.182258×V + 0.000104×V²) / (0.8 + 0.1894393×e^−0.012778×T + 0.2989558×e^−0.1932605×T)
// V = 速度 (m/min), T = 時間 (min)

func calculateVDOTRaw(distanceMeters, timeSeconds float64) float64 {
	if distanceMeters <= 0 || timeSeconds <= 0 {
		return 0
	}
	timeMin := timeSeconds / 60.0
	velocity := distanceMeters / timeMin // m/min

	num := -4.60 + 0.182258*velocity + 0.000104*velocity*velocity
	denom := 0.8 + 0.1894393*math.Exp(-0.012778*timeMin) + 0.2989558*math.Exp(-0.1932605*timeMin)
	if denom <= 0 {
		return 0
	}
	vdot := num / denom
	if vdot < 0 {
		return 0
	}
	return vdot
}

// CalculateVDOT はレース距離(m)とタイム(秒)からVDOT値を算出する（小数1桁に丸め）
func CalculateVDOT(distanceMeters, timeSeconds float64) float64 {
	vdot := calculateVDOTRaw(distanceMeters, timeSeconds)
	if vdot <= 0 {
		return 0
	}
	return math.Round(vdot*10) / 10
}

// velocityFromVDOT は VDOT と持続時間T(分)・強度係数(0-1)から速度(m/min)を逆算する
// VDOT_eff = VDOT * fraction となる V を求める（2次方程式の解）
func velocityFromVDOT(vdot, fraction float64, timeMin float64) float64 {
	if vdot <= 0 || fraction <= 0 {
		return 0
	}
	denom := 0.8 + 0.1894393*math.Exp(-0.012778*timeMin) + 0.2989558*math.Exp(-0.1932605*timeMin)
	target := vdot * fraction * denom
	// num(V) = target  =>  0.000104*V² + 0.182258*V + (-4.60 - target) = 0
	a := 0.000104
	b := 0.182258
	c := -4.60 - target
	disc := b*b - 4*a*c
	if disc < 0 {
		return 0
	}
	v := (-b + math.Sqrt(disc)) / (2 * a)
	if v <= 0 {
		return 0
	}
	return v
}

// velocityToSecPerKm は速度(m/min)を 1kmあたりの秒数 に変換
func velocityToSecPerKm(velocityMperMin float64) float64 {
	if velocityMperMin <= 0 {
		return 0
	}
	return 60000.0 / velocityMperMin // 1000m / (velocity/60) = 60000/velocity
}

// CalculateTrainingPaces はVDOT値から各強度の推奨ペース(秒/km)を算出する
// Jack Daniels の強度目安: E 約72%, M 約82%, T 約90%, I 約97.5%, R 約102%
func CalculateTrainingPaces(vdot float64) model.TrainingPaces {
	// 持続時間の目安(分): Easy長め, M 180, T 20, I 5, R 2
	const (
		tEasy = 60
		tM    = 180
		tT    = 20
		tI    = 5
		tR    = 2
	)
	vM := velocityFromVDOT(vdot, 0.82, tM)
	vT := velocityFromVDOT(vdot, 0.90, tT)
	vI := velocityFromVDOT(vdot, 0.975, tI)
	vR := velocityFromVDOT(vdot, 1.02, tR)

	secM := velocityToSecPerKm(vM)
	secT := velocityToSecPerKm(vT)
	secI := velocityToSecPerKm(vI)
	secR := velocityToSecPerKm(vR)

	// Easy は範囲表示（約 65%〜78% に相当する幅）
	vEHigh := velocityFromVDOT(vdot, 0.65, tEasy)
	vELow := velocityFromVDOT(vdot, 0.78, tEasy)
	secEMax := velocityToSecPerKm(vEHigh) // ゆっくり = 秒数大
	secEMin := velocityToSecPerKm(vELow)  // 速い = 秒数小

	return model.TrainingPaces{
		EasyMinSecPerKm:    math.Round(secEMin*10) / 10,
		EasyMaxSecPerKm:    math.Round(secEMax*10) / 10,
		MarathonSecPerKm:   math.Round(secM*10) / 10,
		ThresholdSecPerKm:  math.Round(secT*10) / 10,
		IntervalSecPerKm:   math.Round(secI*10) / 10,
		RepetitionSecPerKm: math.Round(secR*10) / 10,
	}
}

// PredictRiegelSeconds はリーゲル公式で距離換算したタイム（秒）を返す
// T2 = T1 * (D2/D1)^exponent
func PredictRiegelSeconds(timeSeconds float64, baseDistanceMeters float64, targetDistanceMeters float64, exponent float64) int {
	if timeSeconds <= 0 || baseDistanceMeters <= 0 || targetDistanceMeters <= 0 {
		return 0
	}
	if exponent <= 1.0 || exponent >= 2.0 {
		return 0
	}
	ratio := targetDistanceMeters / baseDistanceMeters
	if ratio <= 0 {
		return 0
	}
	sec := timeSeconds * math.Pow(ratio, exponent)
	if math.IsNaN(sec) || math.IsInf(sec, 0) || sec <= 0 {
		return 0
	}
	return int(math.Round(sec))
}

// CalculateRiegelPredictions は入力(距離・タイム)を起点に、リーゲル公式で主要距離の予想タイムを返す
func CalculateRiegelPredictions(distanceMeters float64, timeSeconds float64, exponent float64) model.RacePredictions {
	return model.RacePredictions{
		FullSeconds:  PredictRiegelSeconds(timeSeconds, distanceMeters, 42195, exponent),
		HalfSeconds:  PredictRiegelSeconds(timeSeconds, distanceMeters, 21097.5, exponent),
		TenKSeconds:  PredictRiegelSeconds(timeSeconds, distanceMeters, 10000, exponent),
		FiveKSeconds: PredictRiegelSeconds(timeSeconds, distanceMeters, 5000, exponent),
	}
}

// PredictRaceTimeSeconds は「指定距離でVDOTが同等になるタイム(秒)」を数値的に逆算する
// 2分探索で解く。計算は内部のraw VDOTを使用する（丸めの影響を避ける）。
func PredictRaceTimeSeconds(vdot float64, distanceMeters float64) int {
	if vdot <= 0 || distanceMeters <= 0 {
		return 0
	}

	// 探索範囲（秒）。上限は距離に応じてざっくり長めに取る
	lo := 60.0               // 1分
	hi := distanceMeters * 3 // 1m=3sec(=5:00/km)相当を起点
	if hi < 600 {            // 最低10分
		hi = 600
	}
	if distanceMeters >= 42195 {
		hi = 8 * 60 * 60 // フルは最大8時間
	} else if distanceMeters >= 21097.5 {
		hi = 5 * 60 * 60 // ハーフは最大5時間
	} else if distanceMeters >= 10000 {
		hi = 3 * 60 * 60 // 10kmは最大3時間
	}

	// hi が十分小さくてVDOTが高すぎる場合があるので、目標VDOTを下回るまで広げる
	for i := 0; i < 20; i++ {
		cur := calculateVDOTRaw(distanceMeters, hi)
		// cur==0 は「遅すぎて式の分子が負になった」などのケース（十分に低い）なので停止してOK
		if cur < vdot {
			break
		}
		hi *= 1.5
		if hi > 12*60*60 { // 安全弁: 12時間
			hi = 12 * 60 * 60
			break
		}
	}

	// 2分探索（vdot は time が増えると下がる前提）
	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2
		cur := calculateVDOTRaw(distanceMeters, mid)
		// cur==0 は「遅すぎる（VDOTが十分低い）」扱いにして、時間を短くする方向へ寄せる
		if cur <= 0 {
			hi = mid
			continue
		}
		if cur > vdot {
			// 速すぎ（VDOT高すぎ）→ 時間を増やす
			lo = mid
		} else {
			// 遅すぎ（VDOT低すぎ）→ 時間を減らす
			hi = mid
		}
	}
	sec := int(math.Round(hi))
	if sec < 1 {
		return 0
	}
	return sec
}

// CalculateRacePredictions は主要距離の予想タイムをまとめて返す
func CalculateRacePredictions(vdot float64) model.RacePredictions {
	return model.RacePredictions{
		FullSeconds:  PredictRaceTimeSeconds(vdot, 42195),
		HalfSeconds:  PredictRaceTimeSeconds(vdot, 21097.5),
		TenKSeconds:  PredictRaceTimeSeconds(vdot, 10000),
		FiveKSeconds: PredictRaceTimeSeconds(vdot, 5000),
	}
}

type vdotStd struct{}

// NewVDOTCalculator は本番用の VDOT まわりの具体実装を返す。
func NewVDOTCalculator() *vdotStd {
	return &vdotStd{}
}

func (vdotStd) CalculateVDOT(distanceMeters, timeSeconds float64) float64 {
	return CalculateVDOT(distanceMeters, timeSeconds)
}

func (vdotStd) CalculateTrainingPaces(vdot float64) model.TrainingPaces {
	return CalculateTrainingPaces(vdot)
}

func (vdotStd) CalculateRiegelPredictions(distanceMeters, timeSeconds, exponent float64) model.RacePredictions {
	return CalculateRiegelPredictions(distanceMeters, timeSeconds, exponent)
}
