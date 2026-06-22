package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/logging"
	"github.com/gin-gonic/gin"
)

// respondHTTPError は err を HTTP 応答に写す。nil なら何もせず false。
func respondHTTPError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	var v *apperrors.Visible
	if errors.As(err, &v) {
		writeAPIError(c, v.HTTP, codeFromVisible(v), v.Msg)
		return true
	}
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		writeAPIError(c, http.StatusNotFound, apperrors.CodeNotFound, "リソースが見つかりません")
	case errors.Is(err, apperrors.ErrUnauthorized):
		writeAPIError(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "認証に失敗しました")
	case errors.Is(err, apperrors.ErrConflict):
		writeAPIError(c, http.StatusConflict, apperrors.CodeConflict, "競合が発生しました")
	default:
		logCtx := context.Background()
		if c.Request != nil {
			logCtx = c.Request.Context()
		}
		logging.FromGin(c).ErrorContext(logCtx, "handler_unhandled_error", slog.Any("err", err))
		writeAPIError(c, http.StatusInternalServerError, apperrors.CodeInternal, "サーバー内部でエラーが発生しました")
	}
	return true
}

// respondPreferVisible は *Visible があればそのまま応答し、なければ fallback と Internal で包む。
func respondPreferVisible(c *gin.Context, err error, fallbackUserMsg string) {
	if err == nil {
		return
	}
	var v *apperrors.Visible
	if errors.As(err, &v) {
		respondHTTPError(c, err)
		return
	}
	respondHTTPError(c, apperrors.InternalMsg(fallbackUserMsg, err))
}
