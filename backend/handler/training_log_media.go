package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// wantsCSVResponse は Accept が text/csv（または application/csv）を要求しているか。
func wantsCSVResponse(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	if accept == "" {
		return false
	}
	for _, part := range strings.Split(accept, ",") {
		media := strings.TrimSpace(strings.Split(part, ";")[0])
		if media == "text/csv" || media == "application/csv" {
			return true
		}
	}
	return false
}

// wantsCSVRequest は CSV 取込リクエストか（multipart file または Content-Type: text/csv）。
func wantsCSVRequest(c *gin.Context) bool {
	if file, _ := c.FormFile("file"); file != nil {
		return true
	}
	return isCSVContentType(c.GetHeader("Content-Type"))
}
