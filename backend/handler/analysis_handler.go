package handler

import (
	"net/http"

	"github.com/KENTA0326/run-sync-pro/database"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/KENTA0326/run-sync-pro/service"
	"github.com/gin-gonic/gin"
)

// GET /auth/analysis （月別走行レポート・Goroutine並列集計）
func MonthlyReport(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できません"})
		return
	}

	var logs []model.TrainingLog
	if err := database.DB.Where("user_id = ?", userID).Order("training_date ASC").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "走行ログの取得に失敗しました"})
		return
	}

	res := service.AnalyzeByMonth(logs)
	c.JSON(http.StatusOK, res)
}
