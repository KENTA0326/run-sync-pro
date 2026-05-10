package handler

import (
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/gin-gonic/gin"
)

// VDOTCalculateInput リクエストボディ
type VDOTCalculateInput struct {
	DistanceMeters float64 `json:"distance_meters" binding:"required,gt=0"`
	TimeSeconds    float64 `json:"time_seconds" binding:"required,gt=0"`
	// リーゲル公式の指数（任意）。指定がなければ 1.08（標準的な市民ランナー）扱い
	RiegelExponent float64 `json:"riegel_exponent" binding:"omitempty,gt=1,lt=2"`
}

// VDOTCalculate は距離(m)とタイム(秒)を受け取り、VDOT値とトレーニングペースを返す。
// POST /api/v1/vdot/calculate （レガシー: POST /vdot/calculate）
func (h *Handlers) VDOTCalculate(c *gin.Context) {
	var input VDOTCalculateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondHTTPError(c, apperrors.BadRequest("distance_meters と time_seconds を正しく指定してください", err))
		return
	}

	vdot := h.vdot.CalculateVDOT(input.DistanceMeters, input.TimeSeconds)
	if vdot <= 0 {
		respondHTTPError(c, apperrors.BadRequest("VDOTを算出できません。距離とタイムを確認してください（3km以上推奨）"))
		return
	}

	paces := h.vdot.CalculateTrainingPaces(vdot)
	exponent := input.RiegelExponent
	if exponent == 0 {
		exponent = 1.08
	}
	riegel := h.vdot.CalculateRiegelPredictions(input.DistanceMeters, input.TimeSeconds, exponent)

	c.JSON(http.StatusOK, gin.H{
		"vdot":               vdot,
		"paces":              paces,
		"riegel_exponent":    exponent,
		"riegel_predictions": riegel,
	})
}
