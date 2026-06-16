package logging

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type ctxKey int

const ctxKeyRequestID ctxKey = 1

// GinLoggerKey は Gin の c.Set に載せる子 *slog.Logger 用キー。
const GinLoggerKey = "runsync.slog"

// WithRequestID は context にリクエストIDを載せる（DB や下流へ伝播させる用）。
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, id)
}

// RequestIDFromContext は WithRequestID で付与した ID を返す。
func RequestIDFromContext(ctx context.Context) string {
	v := ctx.Value(ctxKeyRequestID)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// FromGin はミドルウェアが載せたリクエスト単位のロガーを返す。無ければ slog.Default()。
func FromGin(c *gin.Context) *slog.Logger {
	if v, ok := c.Get(GinLoggerKey); ok {
		if l, ok := v.(*slog.Logger); ok {
			return l
		}
	}
	return slog.Default()
}
