package httpserver

import (
	"github.com/KENTA0326/run-sync-pro/handler"
	"github.com/KENTA0326/run-sync-pro/middleware"
	"github.com/gin-gonic/gin"
)

// PathAPIv1 は REST API の現行バージョンprefix。
const PathAPIv1 = "/api/v1"

// RegisterRoutes はバージョン付き API と、後方互換用レガシー別名を登録する。
func RegisterRoutes(r *gin.Engine, h *handler.Handlers) {
	registerAPIv1(r, h)
	registerLegacyRoutes(r, h)
}

func registerAPIv1(r *gin.Engine, h *handler.Handlers) {
	v1 := r.Group(PathAPIv1)

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ── 認証不要 ─────────────────────────────────────────────────────────────

	v1.POST("/auth/signup", h.SignUp)
	v1.POST("/auth/login", h.Login)

	v1.POST("/vdot/calculate", h.VDOTCalculate)
	v1.POST("/splits/full-marathon", h.FullMarathonSplits)

	// ── 認証必須（JWT）。パスに /auth を嵌めずリソース中心にまとめる ───────────────

	authz := v1.Group("")
	authz.Use(middleware.AuthMiddleware())
	{
		authz.GET("/users/me", h.CurrentUser)

		authz.POST("/shoes", h.CreateShoe)
		authz.GET("/shoes", h.ListShoes)
		authz.DELETE("/shoes/:id", h.DeleteShoe)

		authz.POST("/training-logs", h.CreateTrainingLog)
		authz.GET("/training-logs", h.ListTrainingLogs)
		authz.GET("/training-logs/formatted", h.ListTrainingLogsFormatted)
		authz.POST("/training-logs/import/stream", h.ImportTrainingLogsStream)

		authz.GET("/analysis/monthly", h.MonthlyReport)
	}
}

// registerLegacyRoutes は既存クライアント向け。新規クライアントは /api/v1 を利用すること。
func registerLegacyRoutes(r *gin.Engine, h *handler.Handlers) {
	r.POST("/signup", h.SignUp)
	r.POST("/login", h.Login)

	r.POST("/vdot/calculate", h.VDOTCalculate)
	r.POST("/splits/fullmarathon", h.FullMarathonSplits)

	legacyAuth := r.Group("/auth")
	legacyAuth.Use(middleware.AuthMiddleware())
	{
		legacyAuth.GET("/me", h.CurrentUser)

		legacyAuth.POST("/shoes", h.CreateShoe)
		legacyAuth.GET("/shoes", h.ListShoes)
		legacyAuth.DELETE("/shoes/:id", h.DeleteShoe)

		legacyAuth.POST("/training-logs", h.CreateTrainingLog)
		legacyAuth.GET("/training-logs", h.ListTrainingLogs)
		legacyAuth.GET("/training-logs/formatted", h.ListTrainingLogsFormatted)
		legacyAuth.POST("/training-logs/stream", h.ImportTrainingLogsStream)

		legacyAuth.GET("/analysis", h.MonthlyReport)
	}
}
