package httpclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/config"
)

// ErrNotConfigured は EXTERNAL_API_BASE_URL が未設定のとき。
var ErrNotConfigured = errors.New("httpclient: EXTERNAL_API_BASE_URL is not configured")

// FromEnv は環境変数から Client を構築する。
//
//	EXTERNAL_API_BASE_URL  … 必須（例: https://api.example.com）
//	EXTERNAL_API_API_KEY   … 任意（Bearer）
//	EXTERNAL_API_TIMEOUT_SEC … 既定 10
//	EXTERNAL_API_MAX_RETRIES   … 既定 3（初回を除く再試行回数）
func FromEnv() (*Client, error) {
	base := config.ResolveString("", "EXTERNAL_API_BASE_URL", "")
	if base == "" {
		return nil, ErrNotConfigured
	}

	timeoutSec, err := strconv.Atoi(config.ResolveString("", "EXTERNAL_API_TIMEOUT_SEC", "10"))
	if err != nil || timeoutSec <= 0 {
		timeoutSec = 10
	}
	maxRetries, err := strconv.Atoi(config.ResolveString("", "EXTERNAL_API_MAX_RETRIES", "3"))
	if err != nil || maxRetries < 0 {
		maxRetries = defaultMaxRetry
	}

	return New(Config{
		BaseURL:    base,
		APIKey:     config.ResolveString("", "EXTERNAL_API_API_KEY", ""),
		Timeout:    time.Duration(timeoutSec) * time.Second,
		MaxRetries: maxRetries,
	})
}

// DecodeJSON は 2xx レスポンス本文を JSON デコードする。呼び出し側は resp.Body を Close すること。
func DecodeJSON(resp *http.Response, dest any) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("httpclient: unexpected status %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("httpclient: decode json: %w", err)
	}
	return nil
}
