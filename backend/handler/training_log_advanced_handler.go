package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/KENTA0326/run-sync-pro/database"
	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/timeutil"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	apiDateLayout = "2006-01-02"
	apiTimeLayout = "2006-01-02 15:04:05"

	defaultImportMaxItems        = 1000
	defaultImportAtomic          = true
	defaultImportDisallowUnknown = true
)

// APIDate は JSON 上で YYYY-MM-DD のみを扱う。
type APIDate struct {
	time.Time
}

func (d APIDate) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time.Format(apiDateLayout))
}

func (d *APIDate) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		d.Time = time.Time{}
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("APIDate: %w", apperrors.BadRequest("日付は文字列で指定してください", err))
	}

	t, err := timeutil.ParseCalendarDate(s)
	if err != nil {
		return fmt.Errorf("APIDate: %w", apperrors.BadRequest("日付は YYYY-MM-DD 形式で指定してください", err))
	}
	d.Time = t
	return nil
}

// APITimestamp は JSON 上で YYYY-MM-DD HH:mm:ss に整形する。
type APITimestamp struct {
	time.Time
}

func (t APITimestamp) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Time.Format(apiTimeLayout))
}

type trainingLogFormattedResponse struct {
	ID           uint         `json:"id"`
	TrainingDate APIDate      `json:"training_date"`
	Distance     float64      `json:"distance"`
	Duration     int          `json:"duration"`
	Pace         string       `json:"pace"`
	Memo         string       `json:"memo,omitempty"`
	Kind         int          `json:"kind"`
	ShoeID       uint         `json:"shoe_id"`
	Shoe         *shoeSummary `json:"shoe,omitempty"`
	CreatedAt    APITimestamp `json:"created_at"`
	UpdatedAt    APITimestamp `json:"updated_at"`
}

type shoeSummary struct {
	ID           uint    `json:"id"`
	Brand        string  `json:"brand"`
	Model        string  `json:"model"`
	PurchaseDate APIDate `json:"purchase_date"`
}

func toTrainingLogFormattedResponse(log model.TrainingLog) trainingLogFormattedResponse {
	var shoe *shoeSummary
	if log.Shoe.ID != 0 {
		shoe = &shoeSummary{
			ID:           log.Shoe.ID,
			Brand:        log.Shoe.Brand,
			Model:        log.Shoe.Model,
			PurchaseDate: APIDate{Time: log.Shoe.PurchaseDate},
		}
	}

	return trainingLogFormattedResponse{
		ID:           log.ID,
		TrainingDate: APIDate{Time: log.TrainingDate},
		Distance:     log.Distance,
		Duration:     log.Duration,
		Pace:         log.Pace,
		Memo:         log.Memo,
		Kind:         log.Kind,
		ShoeID:       log.ShoeID,
		Shoe:         shoe,
		CreatedAt:    APITimestamp{Time: log.CreatedAt},
		UpdatedAt:    APITimestamp{Time: log.UpdatedAt},
	}
}

// ListTrainingLogsFormatted はレガシー別名（GET /training-logs?view=formatted と同等）。
func (h *Handlers) ListTrainingLogsFormatted(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Set("view", "formatted")
	c.Request.URL.RawQuery = q.Encode()
	h.ListTrainingLogs(c)
}

type streamTrainingLogInput struct {
	TrainingDate APIDate `json:"training_date"`
	Distance     float64 `json:"distance"`
	Duration     int     `json:"duration"`
	Pace         string  `json:"pace"`
	Memo         string  `json:"memo,omitempty"`
	Kind         int     `json:"kind"`
	ShoeID       uint    `json:"shoe_id"`
}

// TrainingLogImportOptions はストリーム取り込み時のオプション引数。
// フィールドは同一パッケージの Option からのみ設定し、外部からはゲッター経由で読む。
type TrainingLogImportOptions struct {
	maxItems             int
	atomic               bool
	disallowUnknownField bool
}

func (o TrainingLogImportOptions) MaxItems() int              { return o.maxItems }
func (o TrainingLogImportOptions) Atomic() bool               { return o.atomic }
func (o TrainingLogImportOptions) DisallowUnknownField() bool { return o.disallowUnknownField }

func (o TrainingLogImportOptions) validate() error {
	if o.maxItems <= 0 {
		return fmt.Errorf("max items invalid: %w", apperrors.ErrInvalidInput)
	}
	return nil
}

// TrainingLogImportOption は TrainingLogImportOptions を変更する関数型。
type TrainingLogImportOption func(*TrainingLogImportOptions)

// WithImportMaxItems は 1 リクエストで受け取る上限件数を設定する。
func WithImportMaxItems(max int) TrainingLogImportOption {
	return func(o *TrainingLogImportOptions) {
		o.maxItems = max
	}
}

// WithImportAtomic は取り込みを 1 トランザクションで行うかを設定する。
func WithImportAtomic(atomic bool) TrainingLogImportOption {
	return func(o *TrainingLogImportOptions) {
		o.atomic = atomic
	}
}

// WithImportDisallowUnknownField は未知フィールドを拒否するかを設定する。
func WithImportDisallowUnknownField(disallow bool) TrainingLogImportOption {
	return func(o *TrainingLogImportOptions) {
		o.disallowUnknownField = disallow
	}
}

// NewTrainingLogImportOptions はデフォルト値を入れた上でオプションを適用し、最終検証する。
func NewTrainingLogImportOptions(opts ...TrainingLogImportOption) (TrainingLogImportOptions, error) {
	o := TrainingLogImportOptions{
		maxItems:             defaultImportMaxItems,
		atomic:               defaultImportAtomic,
		disallowUnknownField: defaultImportDisallowUnknown,
	}
	for _, opt := range opts {
		opt(&o)
	}
	if err := o.validate(); err != nil {
		return TrainingLogImportOptions{}, apperrors.Annotate("NewTrainingLogImportOptions", err)
	}
	return o, nil
}

func validateStreamTrainingLogInput(in streamTrainingLogInput) error {
	if in.TrainingDate.Time.IsZero() {
		return apperrors.BadRequest("training_date は YYYY-MM-DD 形式で指定してください")
	}
	if in.Distance <= 0 {
		return apperrors.BadRequest("distance は 0 より大きい値で指定してください")
	}
	if in.Duration <= 0 {
		return apperrors.BadRequest("duration は 0 より大きい値で指定してください")
	}
	if in.Pace == "" {
		return apperrors.BadRequest("pace は必須です")
	}
	if in.Kind < 0 || in.Kind > 3 {
		return apperrors.BadRequest("kind は 0-3 の範囲で指定してください")
	}
	if in.ShoeID == 0 {
		return apperrors.BadRequest("shoe_id は必須です")
	}
	return nil
}

func importStreamTrainingLogs(
	ctx context.Context,
	db *gorm.DB,
	userID domain.UserID,
	decoder *json.Decoder,
	options TrainingLogImportOptions,
) (int, error) {
	count := 0
	for decoder.More() {
		if count >= options.MaxItems() {
			return count, apperrors.BadRequest(fmt.Sprintf("1回の取り込み上限（%d件）を超えています", options.MaxItems()))
		}

		var in streamTrainingLogInput
		if err := decoder.Decode(&in); err != nil {
			return count, apperrors.BadRequest("JSON 配列の要素を読み取れませんでした", err)
		}
		if err := importOneTrainingLog(ctx, db, userID, in); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// handleImportTrainingLogsStream は POST /training-logs + JSON 配列本文用。
// 形式: [{"training_date":"2026-04-22","distance":10.0,...}, ...]
func (h *Handlers) handleImportTrainingLogsStream(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	options, err := NewTrainingLogImportOptions(
		// デフォルトを明示。将来の要件ではこの呼び出しに Option を追加するだけで拡張できる。
		WithImportMaxItems(defaultImportMaxItems),
		WithImportAtomic(defaultImportAtomic),
		WithImportDisallowUnknownField(defaultImportDisallowUnknown),
	)
	if err != nil {
		respondHTTPError(c, apperrors.InternalMsg("取り込み設定が不正です", err))
		return
	}

	decoder := json.NewDecoder(c.Request.Body)
	if options.DisallowUnknownField() {
		decoder.DisallowUnknownFields()
	}

	token, err := decoder.Token()
	if err != nil {
		respondHTTPError(c, apperrors.BadRequest("JSON の読み取りに失敗しました", err))
		return
	}
	delim, ok := token.(json.Delim)
	if !ok || delim != '[' {
		respondHTTPError(c, apperrors.BadRequest("トップレベルは JSON 配列で送信してください"))
		return
	}

	createdCount := 0
	if options.Atomic() {
		err = database.WithTx(c.Request.Context(), h.dbCtx(c), func(tx *gorm.DB) error {
			var importErr error
			createdCount, importErr = importStreamTrainingLogs(c.Request.Context(), tx, userID, decoder, options)
			return importErr
		})
	} else {
		createdCount, err = importStreamTrainingLogs(c.Request.Context(), h.dbCtx(c), userID, decoder, options)
	}
	if err != nil {
		respondPreferVisible(c, err, "走行ログの一括取込に失敗しました")
		return
	}

	endToken, err := decoder.Token()
	if err != nil {
		respondHTTPError(c, apperrors.BadRequest("JSON 配列の終端が不正です", err))
		return
	}
	endDelim, ok := endToken.(json.Delim)
	if !ok || endDelim != ']' {
		respondHTTPError(c, apperrors.BadRequest("JSON 配列の終端が不正です"))
		return
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != nil && !errors.Is(err, io.EOF) {
		respondHTTPError(c, apperrors.BadRequest("JSON の末尾に不正なデータがあります", err))
		return
	}
	if len(extra) > 0 {
		respondHTTPError(c, apperrors.BadRequest("JSON の末尾に不正なデータがあります"))
		return
	}

	h.invalidateMonthlyAnalysisCache(c, userID)
	c.JSON(http.StatusOK, gin.H{
		"message":       "走行ログを一括取り込みしました",
		"created_count": createdCount,
	})
}

// ImportTrainingLogsStream はレガシー別名（POST /training-logs + JSON 配列と同等）。
func (h *Handlers) ImportTrainingLogsStream(c *gin.Context) {
	h.handleImportTrainingLogsStream(c)
}
