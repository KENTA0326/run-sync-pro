package main

import (
	"log"
	"strings"
	"time" // これが必要になります

	"github.com/KENTA0326/run-sync-pro/database"
	"github.com/KENTA0326/run-sync-pro/handler"
	"github.com/KENTA0326/run-sync-pro/internal/httpserver"
	"github.com/gin-contrib/cors" // これを追加
	"github.com/gin-gonic/gin"
)

func main() {
	database.Connect()
	h := handler.NewDefaultHandlers(database.DB)

	r := gin.Default()

	// CORS設定: 公式ライブラリで一括設定（これが一番確実です）
	r.Use(cors.New(cors.Config{
		// Docker  compose は 3001→コンテナ3000。ホストで nuxt dev 単体は 3000 が多い。
		AllowOrigins:     []string{"http://localhost:3001", "http://localhost:3000"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	httpserver.RegisterRoutes(r, h)

	// Air がビルド失敗のまま古い tmp/main を動かすと /api/v1 が 404 のままになる。起動ログで確認できるようにする。
	if gin.IsDebugging() {
		for _, ri := range r.Routes() {
			if strings.Contains(ri.Path, "/api/v1/") && (strings.Contains(ri.Path, "signup") || strings.Contains(ri.Path, "health")) {
				log.Printf("route %s %s", ri.Method, ri.Path)
			}
		}
	}

	_ = r.Run(":8080")
}
