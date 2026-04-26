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
		return fmt.Errorf("date must be a string: %w", err)
	}

	t, err := time.Parse(apiDateLayout, s)
	if err != nil {
		return fmt.Errorf("date must be in YYYY-MM-DD format")
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

// GET /auth/training-logs/formatted
func ListTrainingLogsFormatted(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できません"})
		return
	}

	var logs []model.TrainingLog
	if err := database.DB.
		Where("user_id = ?", userID).
		Preload("Shoe").
		Order("training_date DESC, id DESC").
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "走行ログの取得に失敗しました"})
		return
	}

	res := make([]trainingLogFormattedResponse, 0, len(logs))
	for _, log := range logs {
		res = append(res, toTrainingLogFormattedResponse(log))
	}

	c.JSON(http.StatusOK, res)
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

type publicError struct {
	message string
}

func (e publicError) Error() string {
	return e.message
}

// TrainingLogImportOptions はストリーム取り込み時のオプション引数。
// Functional Options Pattern の適用先として使う。
type TrainingLogImportOptions struct {
	MaxItems             int
	Atomic               bool
	DisallowUnknownField bool
}

func (o TrainingLogImportOptions) validate() error {
	if o.MaxItems <= 0 {
		return fmt.Errorf("max items must be greater than 0")
	}
	return nil
}

// TrainingLogImportOption は TrainingLogImportOptions を変更する関数型。
type TrainingLogImportOption func(*TrainingLogImportOptions)

// WithImportMaxItems は 1 リクエストで受け取る上限件数を設定する。
func WithImportMaxItems(max int) TrainingLogImportOption {
	return func(o *TrainingLogImportOptions) {
		o.MaxItems = max
	}
}

// WithImportAtomic は取り込みを 1 トランザクションで行うかを設定する。
func WithImportAtomic(atomic bool) TrainingLogImportOption {
	return func(o *TrainingLogImportOptions) {
		o.Atomic = atomic
	}
}

// WithImportDisallowUnknownField は未知フィールドを拒否するかを設定する。
func WithImportDisallowUnknownField(disallow bool) TrainingLogImportOption {
	return func(o *TrainingLogImportOptions) {
		o.DisallowUnknownField = disallow
	}
}

// NewTrainingLogImportOptions はデフォルト値を入れた上でオプションを適用し、最終検証する。
func NewTrainingLogImportOptions(opts ...TrainingLogImportOption) (TrainingLogImportOptions, error) {
	o := TrainingLogImportOptions{
		MaxItems:             defaultImportMaxItems,
		Atomic:               defaultImportAtomic,
		DisallowUnknownField: defaultImportDisallowUnknown,
	}
	for _, opt := range opts {
		opt(&o)
	}
	if err := o.validate(); err != nil {
		return TrainingLogImportOptions{}, err
	}
	return o, nil
}

func validateStreamTrainingLogInput(in streamTrainingLogInput) error {
	if in.TrainingDate.Time.IsZero() {
		return publicError{message: "training_date は YYYY-MM-DD 形式で指定してください"}
	}
	if in.Distance <= 0 {
		return publicError{message: "distance は 0 より大きい値で指定してください"}
	}
	if in.Duration <= 0 {
		return publicError{message: "duration は 0 より大きい値で指定してください"}
	}
	if in.Pace == "" {
		return publicError{message: "pace は必須です"}
	}
	if in.Kind < 0 || in.Kind > 3 {
		return publicError{message: "kind は 0-3 の範囲で指定してください"}
	}
	if in.ShoeID == 0 {
		return publicError{message: "shoe_id は必須です"}
	}
	return nil
}

func importStreamTrainingLogs(
	ctx context.Context,
	db *gorm.DB,
	userID uint,
	decoder *json.Decoder,
	options TrainingLogImportOptions,
) (int, error) {
	count := 0
	for decoder.More() {
		if count >= options.MaxItems {
			return count, publicError{
				message: fmt.Sprintf("1回の取り込み上限（%d件）を超えています", options.MaxItems),
			}
		}

		var in streamTrainingLogInput
		if err := decoder.Decode(&in); err != nil {
			return count, publicError{message: "JSON 配列の要素を読み取れませんでした"}
		}
		if err := validateStreamTrainingLogInput(in); err != nil {
			return count, err
		}

		var shoe model.Shoe
		if err := db.WithContext(ctx).
			Where("id = ? AND user_id = ?", in.ShoeID, userID).
			First(&shoe).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return count, publicError{message: "指定した shoe_id のシューズが見つかりません"}
			}
			return count, err
		}

		log := model.TrainingLog{
			UserID:       userID,
			TrainingDate: in.TrainingDate.Time,
			Distance:     in.Distance,
			Duration:     in.Duration,
			Pace:         in.Pace,
			Memo:         in.Memo,
			Kind:         in.Kind,
			ShoeID:       in.ShoeID,
		}
		if err := db.WithContext(ctx).Create(&log).Error; err != nil {
			return count, err
		}
		if err := db.WithContext(ctx).Model(&shoe).
			Update("total_distance", gorm.Expr("total_distance + ?", in.Distance)).Error; err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// POST /auth/training-logs/stream
// 形式: [{"training_date":"2026-04-22","distance":10.0,"duration":3600,"pace":"6:00","kind":0,"shoe_id":1}, ...]
func ImportTrainingLogsStream(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できません"})
		return
	}

	options, err := NewTrainingLogImportOptions(
		// デフォルトを明示。将来の要件ではこの呼び出しに Option を追加するだけで拡張できる。
		WithImportMaxItems(defaultImportMaxItems),
		WithImportAtomic(defaultImportAtomic),
		WithImportDisallowUnknownField(defaultImportDisallowUnknown),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取り込み設定が不正です"})
		return
	}

	decoder := json.NewDecoder(c.Request.Body)
	if options.DisallowUnknownField {
		decoder.DisallowUnknownFields()
	}

	token, err := decoder.Token()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON の読み取りに失敗しました"})
		return
	}
	delim, ok := token.(json.Delim)
	if !ok || delim != '[' {
		c.JSON(http.StatusBadRequest, gin.H{"error": "トップレベルは JSON 配列で送信してください"})
		return
	}

	createdCount := 0
	if options.Atomic {
		err = database.WithGlobalTx(c.Request.Context(), func(tx *gorm.DB) error {
			var importErr error
			createdCount, importErr = importStreamTrainingLogs(c.Request.Context(), tx, userID, decoder, options)
			return importErr
		})
	} else {
		createdCount, err = importStreamTrainingLogs(c.Request.Context(), database.DB, userID, decoder, options)
	}
	if err != nil {
		var pubErr publicError
		if errors.As(err, &pubErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": pubErr.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "走行ログの一括取込に失敗しました"})
		return
	}

	endToken, err := decoder.Token()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON 配列の終端が不正です"})
		return
	}
	endDelim, ok := endToken.(json.Delim)
	if !ok || endDelim != ']' {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON 配列の終端が不正です"})
		return
	}

	// 終端 "]" のあとに余計な JSON がないか確認する。
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON の末尾に不正なデータがあります"})
		return
	}
	if len(extra) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON の末尾に不正なデータがあります"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "走行ログを一括取り込みしました",
		"created_count": createdCount,
	})
}
