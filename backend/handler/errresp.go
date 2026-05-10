package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/gin-gonic/gin"
)

// respondHTTPError は err を HTTP 応答に写す。nil なら何もせず false。
func respondHTTPError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	var v *apperrors.Visible
	if errors.As(err, &v) {
		c.JSON(v.HTTP, gin.H{"error": v.Msg})
		return true
	}
	if errors.Is(err, apperrors.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "リソースが見つかりません"})
		return true
	}
	if errors.Is(err, apperrors.ErrUnauthorized) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証に失敗しました"})
		return true
	}
	log.Printf("handler error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバー内部でエラーが発生しました"})
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
