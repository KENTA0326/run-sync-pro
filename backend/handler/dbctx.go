package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// バインディングの使い分け（Gin）:
//   - ShouldBindJSON … POST/PUT の JSON（json + binding タグ）
//   - ShouldBindQuery … GET のクエリ（form + binding タグ）
//   - ShouldBindUri    … パスパラメータ（uri + binding タグ）
// 失敗時は respondHTTPError(..., apperrors.BadRequest(...)) で統一する。

// dbCtx はリクエストの context を GORM に載せる（request_id・userID・キャンセル伝播）。
func (h *Handlers) dbCtx(c *gin.Context) *gorm.DB {
	return h.db.WithContext(c.Request.Context())
}
