package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/KENTA0326/run-sync-pro/internal/logging"
	"github.com/gin-gonic/gin"
)

const headerRequestID = "X-Request-ID"

// SlogRequestMiddleware はリクエストID付与・context 伝播・子 slog ロガー・HTTP アクセスログ（JSON slog 前提）を行う。
// gin.Recovery より内側に登録すること（panic 後も defer で status を記録しやすい）。
func SlogRequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := incomingOrNewRequestID(c.GetHeader(headerRequestID))
		c.Writer.Header().Set(headerRequestID, id)

		ctx := logging.WithRequestID(c.Request.Context(), id)
		c.Request = c.Request.WithContext(ctx)

		base := slog.Default().With(
			slog.String("request_id", id),
			slog.String("http_method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)
		c.Set(logging.GinLoggerKey, base)

		start := time.Now()
		defer func() {
			status := c.Writer.Status()
			if status == 0 {
				status = http.StatusOK
			}
			log := base.With(
				slog.Int("status", status),
				slog.Int64("latency_ms", time.Since(start).Milliseconds()),
				slog.String("client_ip", c.ClientIP()),
			)
			const msg = "http_request"
			switch {
			case status >= 500:
				log.ErrorContext(c.Request.Context(), msg)
			case status >= 400:
				log.WarnContext(c.Request.Context(), msg)
			default:
				log.InfoContext(c.Request.Context(), msg)
			}
		}()

		c.Next()
	}
}

func incomingOrNewRequestID(in string) string {
	if s := sanitizeRequestID(in); s != "" {
		return s
	}
	return newRequestID()
}

func sanitizeRequestID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 128 {
		return ""
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			continue
		}
		return ""
	}
	return s
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b[:])
}
