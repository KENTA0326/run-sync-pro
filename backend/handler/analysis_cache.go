package handler

import (
	"log/slog"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/logging"
	"github.com/gin-gonic/gin"
)

func (h *Handlers) invalidateMonthlyAnalysisCache(c *gin.Context, userID domain.UserID) {
	if err := h.analysisCache.Invalidate(c.Request.Context(), userID); err != nil {
		logging.FromGin(c).WarnContext(c.Request.Context(), "analysis_cache_invalidate_failed",
			slog.Uint64("user_id", uint64(userID.Uint())),
			slog.Any("err", err),
		)
	}
}
