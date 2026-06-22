package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/config"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/redis/go-redis/v9"
)

const monthlyAnalysisKeyPrefix = "analysis:monthly:user:"

// MonthlyAnalysis は月別分析 API レスポンスのキャッシュ。
type MonthlyAnalysis struct {
	client *redis.Client
	ttl    time.Duration
}

// NewMonthlyAnalysis は Redis クライアントからキャッシュを構築する（client は非 nil であること）。
func NewMonthlyAnalysis(client *redis.Client) *MonthlyAnalysis {
	ttlSec, err := strconv.Atoi(config.ResolveString("", "ANALYSIS_CACHE_TTL_SEC", "300"))
	if err != nil || ttlSec <= 0 {
		ttlSec = 300
	}
	return &MonthlyAnalysis{
		client: client,
		ttl:    time.Duration(ttlSec) * time.Second,
	}
}

// NoopMonthlyAnalysis は Redis 未使用時の no-op 実装。
type NoopMonthlyAnalysis struct{}

func (NoopMonthlyAnalysis) Get(context.Context, domain.UserID) (model.AnalysisResponse, bool, error) {
	return model.AnalysisResponse{}, false, nil
}

func (NoopMonthlyAnalysis) Set(context.Context, domain.UserID, model.AnalysisResponse) error {
	return nil
}

func (NoopMonthlyAnalysis) Invalidate(context.Context, domain.UserID) error {
	return nil
}

func monthlyAnalysisKey(userID domain.UserID) string {
	return monthlyAnalysisKeyPrefix + strconv.FormatUint(uint64(userID.Uint()), 10)
}

func (c *MonthlyAnalysis) Get(ctx context.Context, userID domain.UserID) (model.AnalysisResponse, bool, error) {
	raw, err := c.client.Get(ctx, monthlyAnalysisKey(userID)).Bytes()
	if err == redis.Nil {
		return model.AnalysisResponse{}, false, nil
	}
	if err != nil {
		return model.AnalysisResponse{}, false, fmt.Errorf("monthly analysis cache get: %w", err)
	}
	var res model.AnalysisResponse
	if err := json.Unmarshal(raw, &res); err != nil {
		_ = c.client.Del(ctx, monthlyAnalysisKey(userID)).Err()
		return model.AnalysisResponse{}, false, fmt.Errorf("monthly analysis cache decode: %w", err)
	}
	return res, true, nil
}

func (c *MonthlyAnalysis) Set(ctx context.Context, userID domain.UserID, res model.AnalysisResponse) error {
	raw, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("monthly analysis cache encode: %w", err)
	}
	if err := c.client.Set(ctx, monthlyAnalysisKey(userID), raw, c.ttl).Err(); err != nil {
		return fmt.Errorf("monthly analysis cache set: %w", err)
	}
	return nil
}

func (c *MonthlyAnalysis) Invalidate(ctx context.Context, userID domain.UserID) error {
	if err := c.client.Del(ctx, monthlyAnalysisKey(userID)).Err(); err != nil {
		return fmt.Errorf("monthly analysis cache invalidate: %w", err)
	}
	return nil
}

// NewMonthlyAnalysisForTest はテスト用に TTL を指定して構築する。
func NewMonthlyAnalysisForTest(client *redis.Client, ttl time.Duration) *MonthlyAnalysis {
	return &MonthlyAnalysis{client: client, ttl: ttl}
}
