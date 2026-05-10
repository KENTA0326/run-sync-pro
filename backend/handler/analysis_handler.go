package handler

import (
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/gin-gonic/gin"
)

// GET /api/v1/analysis/monthly （月別走行レポート・Goroutine並列集計、レガシー: GET /auth/analysis）
func (h *Handlers) MonthlyReport(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	var logs []model.TrainingLog
	if err := h.db.Where("user_id = ?", userID).Order("training_date ASC").Find(&logs).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("走行ログの取得に失敗しました", apperrors.Annotate("db analysis logs", err)))
		return
	}

	res := h.analyzer.AnalyzeByMonth(logs)
	c.JSON(http.StatusOK, res)
}
