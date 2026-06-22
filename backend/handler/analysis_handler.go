package handler

import (
	"log/slog"
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/logging"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/gin-gonic/gin"
)

// GET /api/v1/analysis/monthly （月別走行レポート・Goroutine並列集計、レガシー: GET /auth/analysis）
func (h *Handlers) MonthlyReport(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	ctx := c.Request.Context()
	if cached, hit, err := h.analysisCache.Get(ctx, userID); err != nil {
		logging.FromGin(c).WarnContext(ctx, "analysis_cache_get_failed",
			slog.Uint64("user_id", uint64(userID.Uint())),
			slog.Any("err", err),
		)
	} else if hit {
		c.JSON(http.StatusOK, cached)
		return
	}

	var logs []model.TrainingLog
	if err := h.dbCtx(c).Where("user_id = ?", userID.Uint()).Order("training_date ASC").Find(&logs).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("走行ログの取得に失敗しました", apperrors.Annotate("db analysis logs", err)))
		return
	}

	res := h.analyzer.AnalyzeByMonth(logs)

	if err := h.analysisCache.Set(ctx, userID, res); err != nil {
		logging.FromGin(c).WarnContext(ctx, "analysis_cache_set_failed",
			slog.Uint64("user_id", uint64(userID.Uint())),
			slog.Any("err", err),
		)
	}

	c.JSON(http.StatusOK, res)
}
