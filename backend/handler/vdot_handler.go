package handler

import (
	"net/http"

	"github.com/KENTA0326/run-sync-pro/service"
	"github.com/gin-gonic/gin"
)

// VDOTCalculateInput リクエストボディ
type VDOTCalculateInput struct {
	DistanceMeters float64 `json:"distance_meters" binding:"required,gt=0"`
	TimeSeconds   float64 `json:"time_seconds" binding:"required,gt=0"`
	// リーゲル公式の指数（任意）。指定がなければ 1.08（標準的な市民ランナー）扱い
	RiegelExponent float64 `json:"riegel_exponent" binding:"omitempty,gt=1,lt=2"`
}

// VDOTCalculate 距離(m)とタイム(秒)を受け取り、VDOT値とトレーニングペースを返す
func VDOTCalculate(c *gin.Context) {
	var input VDOTCalculateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "distance_meters と time_seconds を正しく指定してください"})
		return
	}

	vdot := service.CalculateVDOT(input.DistanceMeters, input.TimeSeconds)
	if vdot <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VDOTを算出できません。距離とタイムを確認してください（3km以上推奨）"})
		return
	}

	paces := service.CalculateTrainingPaces(vdot)
	exponent := input.RiegelExponent
	if exponent == 0 {
		exponent = 1.08
	}
	riegel := service.CalculateRiegelPredictions(input.DistanceMeters, input.TimeSeconds, exponent)

	c.JSON(http.StatusOK, gin.H{
		"vdot":        vdot,
		"paces":       paces,
		"riegel_exponent": exponent,
		"riegel_predictions": riegel,
	})
}
