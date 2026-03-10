package main

import (
	"time" // これが必要になります

	"github.com/KENTA0326/run-sync-pro/database"
	"github.com/KENTA0326/run-sync-pro/handler"
	"github.com/KENTA0326/run-sync-pro/middleware"
	"github.com/gin-contrib/cors" // これを追加
	"github.com/gin-gonic/gin"
)

func main() {
	database.Connect()
	r := gin.Default()

	// CORS設定: 公式ライブラリで一括設定（これが一番確実です）
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3001"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	
	r.POST("/signup", handler.SignUp)
	r.POST("/login", handler.Login)

	// VDOT計算（認証不要で利用可能）
	r.POST("/vdot/calculate", handler.VDOTCalculate)
	// スプリット計算（フルマラソン・10km単位ページング）
	r.POST("/splits/fullmarathon", handler.FullMarathonSplits)

	// --- ここから追加 ---
	authGroup := r.Group("/auth")
	authGroup.Use(middleware.AuthMiddleware()) // 関所を設置
	{
		// 認証チェック用のシンプルなエンドポイント
		authGroup.GET("/me", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			c.JSON(200, gin.H{
				"user_id": userID,
				"message": "認証に成功しています！",
			})
		})

		// シューズ管理
		authGroup.POST("/shoes", handler.CreateShoe)
		authGroup.GET("/shoes", handler.ListShoes)
		authGroup.DELETE("/shoes/:id", handler.DeleteShoe)

		// 走行ログ管理
		authGroup.POST("/training-logs", handler.CreateTrainingLog)
		authGroup.GET("/training-logs", handler.ListTrainingLogs)
	}
	// --- ここまで ---

	r.Run(":8080")
}