package main

import (
	"github.com/KENTA0326/run-sync-pro/database"
	"github.com/gin-gonic/gin"
)

func main() {
	// DB接続
	database.Connect()

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, RunSync Pro API with Gin and GORM!",
		})
	})

	r.Run(":8080")
}