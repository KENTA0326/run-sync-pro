package httpserver

import (
	"github.com/KENTA0326/run-sync-pro/handler"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
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

	// ── 認証不要（動詞の方が分かりやすい操作はパスに含める）────────────────────

	v1.POST("/auth/signup", h.SignUp)
	v1.POST("/auth/login", h.Login)
	v1.POST("/auth/password-reset/request", h.RequestPasswordReset)
	v1.POST("/auth/password-reset/confirm", h.ConfirmPasswordReset)

	v1.POST("/vdot/calculate", h.VDOTCalculate)
	v1.POST("/splits/full-marathon", h.FullMarathonSplits)

	// ── 認証必須（JWT）────────────────────────────────────────────────────────

	authz := v1.Group("")
	authz.Use(middleware.AuthMiddleware())
	{
		authz.GET("/users/me", middleware.RequirePermission(h.DB(), domain.ResourceUserProfile, domain.ActionRead), h.CurrentUser)

		// シューズ: コレクション + :id で個体を特定
		authz.GET("/shoes", middleware.RequirePermission(h.DB(), domain.ResourceShoe, domain.ActionRead), h.ListShoes)
		authz.POST("/shoes", middleware.RequirePermission(h.DB(), domain.ResourceShoe, domain.ActionWrite), h.CreateShoe)
		authz.GET("/shoes/:id", middleware.RequirePermission(h.DB(), domain.ResourceShoe, domain.ActionRead), h.GetShoe)
		authz.DELETE("/shoes/:id", middleware.RequirePermission(h.DB(), domain.ResourceShoe, domain.ActionWrite), h.DeleteShoe)

		// 走行ログ: export/import は GET/POST /training-logs の Content negotiation（レガシー別名は /auth 側）
		authz.GET("/training-logs", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionRead), h.ListTrainingLogs)
		authz.POST("/training-logs", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionWrite), h.CreateTrainingLog)
		authz.GET("/training-logs/:id", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionRead), h.GetTrainingLog)

		authz.GET("/analysis/monthly", middleware.RequirePermission(h.DB(), domain.ResourceAnalysis, domain.ActionRead), h.MonthlyReport)

		authz.POST("/graphql", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionRead), h.GraphQL)
		authz.GET("/ws/training-logs", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionRead), h.TrainingLogsWebSocket)
	}
}

// registerLegacyRoutes は既存クライアント向け。新規クライアントは /api/v1 を利用すること。
func registerLegacyRoutes(r *gin.Engine, h *handler.Handlers) {
	r.POST("/signup", h.SignUp)
	r.POST("/login", h.Login)
	r.POST("/auth/password-reset/request", h.RequestPasswordReset)
	r.POST("/auth/password-reset/confirm", h.ConfirmPasswordReset)

	r.POST("/vdot/calculate", h.VDOTCalculate)
	r.POST("/splits/fullmarathon", h.FullMarathonSplits)

	legacyAuth := r.Group("/auth")
	legacyAuth.Use(middleware.AuthMiddleware())
	{
		legacyAuth.GET("/me", middleware.RequirePermission(h.DB(), domain.ResourceUserProfile, domain.ActionRead), h.CurrentUser)

		legacyAuth.GET("/shoes", middleware.RequirePermission(h.DB(), domain.ResourceShoe, domain.ActionRead), h.ListShoes)
		legacyAuth.POST("/shoes", middleware.RequirePermission(h.DB(), domain.ResourceShoe, domain.ActionWrite), h.CreateShoe)
		legacyAuth.GET("/shoes/:id", middleware.RequirePermission(h.DB(), domain.ResourceShoe, domain.ActionRead), h.GetShoe)
		legacyAuth.DELETE("/shoes/:id", middleware.RequirePermission(h.DB(), domain.ResourceShoe, domain.ActionWrite), h.DeleteShoe)

		legacyAuth.GET("/training-logs", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionRead), h.ListTrainingLogs)
		legacyAuth.GET("/training-logs/export/csv", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionRead), h.ExportTrainingLogsCSV)
		legacyAuth.GET("/training-logs/formatted", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionRead), h.ListTrainingLogsFormatted)
		legacyAuth.POST("/training-logs", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionWrite), h.CreateTrainingLog)
		legacyAuth.GET("/training-logs/:id", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionRead), h.GetTrainingLog)
		legacyAuth.POST("/training-logs/stream", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionWrite), h.ImportTrainingLogsStream)
		legacyAuth.POST("/training-logs/import/csv", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionWrite), h.ImportTrainingLogsCSV)

		legacyAuth.GET("/analysis", middleware.RequirePermission(h.DB(), domain.ResourceAnalysis, domain.ActionRead), h.MonthlyReport)

		legacyAuth.POST("/graphql", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionRead), h.GraphQL)
		legacyAuth.GET("/ws/training-logs", middleware.RequirePermission(h.DB(), domain.ResourceTrainingLog, domain.ActionRead), h.TrainingLogsWebSocket)
	}
}
