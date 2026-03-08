package main

import (
	"github.com/KENTA0326/run-sync-pro/database"
	"github.com/KENTA0326/run-sync-pro/handler"
	"github.com/KENTA0326/run-sync-pro/middleware" // 1. これが必要！
	"github.com/gin-gonic/gin"
)

func main() {
	database.Connect()
	r := gin.Default()

	r.POST("/signup", handler.SignUp)
	r.POST("/login", handler.Login)

	// --- ここから追加 ---
	authGroup := r.Group("/auth")
	authGroup.Use(middleware.AuthMiddleware()) // 関所を設置
	{
		authGroup.GET("/me", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			c.JSON(200, gin.H{
				"user_id": userID,
				"message": "認証に成功しています！",
			})
		})
	}
	// --- ここまで ---

	r.Run(":8080")
}