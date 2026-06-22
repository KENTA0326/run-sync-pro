package cache

import (
	"context"
	"log/slog"
	"strings"

	"github.com/KENTA0326/run-sync-pro/internal/config"
	"github.com/redis/go-redis/v9"
)

// ConnectRedisFromEnv は REDIS_URL が設定されていればクライアントを返す。未設定なら nil（キャッシュ無効）。
func ConnectRedisFromEnv() *redis.Client {
	url := strings.TrimSpace(config.ResolveString("", "REDIS_URL", ""))
	if url == "" {
		slog.Info("redis_disabled", slog.String("reason", "REDIS_URL is empty"))
		return nil
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		slog.Error("redis_parse_url_failed", slog.Any("err", err))
		return nil
	}
	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		slog.Error("redis_ping_failed", slog.Any("err", err))
		_ = client.Close()
		return nil
	}
	slog.Info("redis_connected")
	return client
}
