package domain

import (
	"context"
	"time"
)

// PasswordResetRepository はパスワードリセットトークンの永続化インターフェース。
type PasswordResetRepository interface {
	Issue(ctx context.Context, userID uint, ttl time.Duration) (plainToken string, err error)
	Consume(ctx context.Context, plainToken string) (userID uint, err error)
}
