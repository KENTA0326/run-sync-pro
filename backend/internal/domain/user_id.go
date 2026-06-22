package domain

import (
	"context"
	"net/http"
)

// UserID は認証済みユーザーの DB 主キー（uint）を表す値オブジェクト。
type UserID uint

type ctxKeyUserID struct{}

// IsValid は正の ID かどうかを返す。
func (id UserID) IsValid() bool {
	return id > 0
}

// Uint は GORM / model 層向けに primitive へ変換する。
func (id UserID) Uint() uint {
	return uint(id)
}

// WithUserID は認証済み userID を context.Context に載せる（イミュータブル）。
// AuthMiddleware が c.Request = c.Request.WithContext(...) で差し替える。
func WithUserID(ctx context.Context, id UserID) context.Context {
	return context.WithValue(ctx, ctxKeyUserID{}, id)
}

// UserIDFromStdContext は context.Context から userID を返す。
func UserIDFromStdContext(ctx context.Context) (UserID, bool) {
	if ctx == nil {
		return 0, false
	}
	raw := ctx.Value(ctxKeyUserID{})
	id, ok := ParseUserID(raw)
	if !ok || !id.IsValid() {
		return 0, false
	}
	return id, true
}

// UserIDFromRequest は *http.Request の context から userID を返す。
func UserIDFromRequest(req *http.Request) (UserID, bool) {
	if req == nil {
		return 0, false
	}
	return UserIDFromStdContext(req.Context())
}

// contextGetter はレガシー互換用のキー値ストア（テスト等）。
type contextGetter interface {
	Get(string) (any, bool)
}

// UserIDFromContext は Gin の c.Get("userID") 等レガシー経路向け。新規コードは UserIDFromRequest を使う。
func UserIDFromContext(c contextGetter) (UserID, bool) {
	raw, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	id, ok := ParseUserID(raw)
	if !ok || !id.IsValid() {
		return 0, false
	}
	return id, true
}

// ParseUserID は JWT claims 等の生値を UserID に変換する。
func ParseUserID(raw any) (UserID, bool) {
	switch v := raw.(type) {
	case UserID:
		return v, true
	case uint:
		return UserID(v), true
	case int:
		if v < 0 {
			return 0, false
		}
		return UserID(v), true
	case float64:
		if v < 0 {
			return 0, false
		}
		return UserID(v), true
	default:
		return 0, false
	}
}
