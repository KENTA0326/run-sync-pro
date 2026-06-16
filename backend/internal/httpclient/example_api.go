package httpclient

import (
	"context"
	"fmt"
)

// ExampleAPI は外部 API をメソッドでラップする例（スマートウォッチ連携などの雛形）。
type ExampleAPI struct {
	client *Client
}

// NewExampleAPI は既存 Client をラップする。
func NewExampleAPI(c *Client) *ExampleAPI {
	return &ExampleAPI{client: c}
}

// NewExampleAPIFromEnv は環境変数から ExampleAPI を構築する。
func NewExampleAPIFromEnv() (*ExampleAPI, error) {
	c, err := FromEnv()
	if err != nil {
		return nil, err
	}
	return NewExampleAPI(c), nil
}

type syncActivitiesRequest struct {
	UserID string `json:"user_id"`
}

type syncActivitiesResponse struct {
	Imported int `json:"imported"`
}

// SyncActivities は POST /v1/activities/sync の呼び出し例。
func (a *ExampleAPI) SyncActivities(ctx context.Context, userID string) (int, error) {
	resp, err := a.client.PostJSON(ctx, "/v1/activities/sync", syncActivitiesRequest{UserID: userID})
	if err != nil {
		return 0, err
	}

	var out syncActivitiesResponse
	if err := DecodeJSON(resp, &out); err != nil {
		return 0, fmt.Errorf("example api sync activities: %w", err)
	}
	return out.Imported, nil
}
