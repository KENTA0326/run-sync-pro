package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/KENTA0326/run-sync-pro/database"
	"github.com/KENTA0326/run-sync-pro/handler"
	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/cache"
	"github.com/KENTA0326/run-sync-pro/internal/config"
	"github.com/KENTA0326/run-sync-pro/internal/grpcserver"
	"github.com/KENTA0326/run-sync-pro/internal/httpserver"
	"github.com/KENTA0326/run-sync-pro/internal/logging"
	"github.com/KENTA0326/run-sync-pro/internal/repository"
	"github.com/KENTA0326/run-sync-pro/internal/response"
	"github.com/KENTA0326/run-sync-pro/internal/usecase"
	"github.com/KENTA0326/run-sync-pro/internal/validation"
	"github.com/KENTA0326/run-sync-pro/middleware"
	"github.com/KENTA0326/run-sync-pro/internal/domainservice"
	"github.com/KENTA0326/run-sync-pro/internal/infrastructure/auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	logging.InitFromEnv()
	database.Connect()
	redisClient := cache.ConnectRedisFromEnv()
	var analysisCache handler.MonthlyAnalysisCache = cache.NoopMonthlyAnalysis{}
	if redisClient != nil {
		analysisCache = cache.NewMonthlyAnalysis(redisClient)
	}

	// 組み立て: domain.Repository → usecase → handler（Interface Adapters）
	// 計算・JWT 等の具象は domainservice / infrastructure/auth から注入する。
	userRepo := repository.NewUserRepository(database.DB)
	shoeRepo := repository.NewShoeRepository(database.DB)
	trainingLogRepo := repository.NewTrainingLogRepository(database.DB)
	passwordResetRepo := repository.NewPasswordResetRepository(database.DB)

	authInfra := auth.NewAuth()
	authUC := usecase.NewAuthUseCase(userRepo, authInfra)
	shoeUC := usecase.NewShoeUseCase(shoeRepo)
	trainingLogUC := usecase.NewTrainingLogUseCase(trainingLogRepo, shoeRepo)
	passwordResetUC := usecase.NewPasswordResetUseCase(userRepo, passwordResetRepo, authInfra)

	h := handler.NewHandlers(
		database.DB,
		authInfra,
		authUC,
		shoeUC,
		trainingLogUC,
		passwordResetUC,
		domainservice.NewAnalyzer(),
		analysisCache,
		domainservice.NewVDOTCalculator(),
		domainservice.NewMarathonSplits(),
		func() time.Duration {
			minStr := config.ResolveString("", "PASSWORD_RESET_TOKEN_TTL_MINUTES", "60")
			min, err := strconv.Atoi(minStr)
			if err != nil || min <= 0 {
				min = 60
			}
			return time.Duration(min) * time.Minute
		},
		func() string {
			return config.ResolveString("", "FRONTEND_BASE_URL", "http://localhost:3001")
		},
	)

	r := gin.New()
	validation.RegisterGinBindingValidators()

	// ミドルウェア順（外→内）: Recovery → CORS → SlogRequest → ルート（Auth はグループ単位）
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logging.FromGin(c).Error("panic_recovered", slog.Any("panic", recovered))
		response.WriteError(c, http.StatusInternalServerError, apperrors.CodeInternal, "サーバー内部でエラーが発生しました")
	}))

	// CORS を Logger より外側: OPTIONS プリフライトはここで返し、アクセスログのノイズを減らす
	r.Use(cors.New(cors.Config{
		// Docker  compose は 3001→コンテナ3000。ホストで nuxt dev 単体は 3000 が多い。
		AllowOrigins:     []string{"http://localhost:3001", "http://localhost:3000"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(middleware.SlogRequestMiddleware())

	httpserver.RegisterRoutes(r, h)
	if grpcPort := strings.TrimSpace(os.Getenv("GRPC_PORT")); grpcPort != "" {
		go func() {
			if err := grpcserver.Start(":" + grpcPort); err != nil {
				slog.Error("grpc_server_failed", slog.Any("err", err))
			}
		}()
	}
	slog.Info("httpserver_routes_registered",
		slog.String("api_prefix", httpserver.PathAPIv1),
		slog.String("note", "このログが無いとコンテナは古いバイナリの可能性"))

	// Air がビルド失敗のまま古い tmp/main を動かすと /api/v1 が 404 のままになる。起動ログで確認できるようにする。
	if gin.IsDebugging() {
		for _, ri := range r.Routes() {
			if strings.Contains(ri.Path, "/api/v1/") && (strings.Contains(ri.Path, "signup") || strings.Contains(ri.Path, "health")) {
				slog.Debug("route_registered", slog.String("method", ri.Method), slog.String("path", ri.Path))
			}
		}
	}

	_ = r.Run(":8080")
}
