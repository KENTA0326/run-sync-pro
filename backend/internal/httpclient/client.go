// Package httpclient は外部 REST API 向けの HTTP クライアント（タイムアウト・5xx リトライ）を提供する。
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout    = 10 * time.Second
	defaultMaxRetry = 3
	defaultRetryBase  = 200 * time.Millisecond
)

// Config は Client 生成時の設定。
type Config struct {
	BaseURL    string
	APIKey     string        // 非空なら Authorization: Bearer を付与
	Timeout    time.Duration // 0 なら defaultTimeout
	MaxRetries int           // 失敗後の再試行上限（0 なら defaultMaxRetry）。初回を除く
	HTTPClient *http.Client  // 省略時は Timeout 付きで生成
}

// Client はベース URL・認証・http.Client を保持する外部 API クライアント。
type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
	maxRetries int
}

// New は Config から Client を構築する。
func New(cfg Config) (*Client, error) {
	base := strings.TrimSpace(cfg.BaseURL)
	if base == "" {
		return nil, fmt.Errorf("httpclient: BaseURL is required")
	}
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("httpclient: invalid BaseURL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("httpclient: BaseURL must include scheme and host")
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	hc := cfg.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: timeout}
	} else if hc.Timeout == 0 {
		hc.Timeout = timeout
	}

	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	if maxRetries == 0 {
		maxRetries = defaultMaxRetry
	}

	return &Client{
		baseURL:    u,
		apiKey:     strings.TrimSpace(cfg.APIKey),
		httpClient: hc,
		maxRetries: maxRetries,
	}, nil
}

// Get は path（先頭 / 可）へ GET する。
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil, "")
}

// PostJSON は JSON ボディで POST する。
func (c *Client) PostJSON(ctx context.Context, path string, payload any) (*http.Response, error) {
	var body io.Reader
	contentType := ""
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("httpclient: marshal json: %w", err)
		}
		body = bytes.NewReader(b)
		contentType = "application/json"
	}
	return c.Do(ctx, http.MethodPost, path, body, contentType)
}

// Do は絶対 URL を組み立ててリクエストを送る（リトライ込み）。
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, error) {
	reqURL, err := c.resolveURL(path)
	if err != nil {
		return nil, err
	}

	var bodyBytes []byte
	if body != nil {
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("httpclient: read body: %w", err)
		}
	}

	var lastErr error
	var lastStatus int

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			if err := waitRetry(ctx, attempt); err != nil {
				return nil, err
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL, bytesReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		if c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < c.maxRetries && isRetryableNetErr(err) {
				continue
			}
			return nil, fmt.Errorf("httpclient: request failed: %w", err)
		}

		if !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		lastStatus = resp.StatusCode
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		if attempt >= c.maxRetries {
			return nil, fmt.Errorf("httpclient: max retries exceeded, last status %d", lastStatus)
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("httpclient: max retries exceeded, last status %d", lastStatus)
}

func (c *Client) resolveURL(path string) (string, error) {
	ref, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("httpclient: invalid path: %w", err)
	}
	return c.baseURL.ResolveReference(ref).String(), nil
}

func bytesReader(b []byte) io.Reader {
	if len(b) == 0 {
		return nil
	}
	return bytes.NewReader(b)
}

func waitRetry(ctx context.Context, attempt int) error {
	delay := defaultRetryBase * time.Duration(1<<(attempt-1))
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
