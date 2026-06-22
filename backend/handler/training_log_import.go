package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/KENTA0326/run-sync-pro/database"
	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/importio"
	"github.com/KENTA0326/run-sync-pro/internal/timeutil"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func csvRecordToInput(rec importio.TrainingLogCSVRecord) (streamTrainingLogInput, error) {
	t, err := timeutil.ParseCalendarDate(rec.TrainingDate)
	if err != nil {
		return streamTrainingLogInput{}, apperrors.BadRequest("training_date は YYYY-MM-DD 形式で指定してください", err)
	}
	return streamTrainingLogInput{
		TrainingDate: APIDate{Time: t},
		Distance:     rec.Distance,
		Duration:     rec.Duration,
		Pace:         rec.Pace,
		Memo:         rec.Memo,
		Kind:         rec.Kind,
		ShoeID:       rec.ShoeID,
	}, nil
}

func importOneTrainingLog(ctx context.Context, db *gorm.DB, userID domain.UserID, in streamTrainingLogInput) error {
	if err := validateStreamTrainingLogInput(in); err != nil {
		return err
	}

	var shoe model.Shoe
	if err := db.WithContext(ctx).
		Where("id = ? AND user_id = ?", in.ShoeID, userID.Uint()).
		First(&shoe).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NotFoundMsg("指定した shoe_id のシューズが見つかりません", err)
		}
		return apperrors.Annotate("import training log shoe", err)
	}

	log := model.TrainingLog{
		UserID:       userID.Uint(),
		TrainingDate: in.TrainingDate.Time,
		Distance:     in.Distance,
		Duration:     in.Duration,
		Pace:         in.Pace,
		Memo:         in.Memo,
		Kind:         in.Kind,
		ShoeID:       in.ShoeID,
	}
	if err := db.WithContext(ctx).Create(&log).Error; err != nil {
		return apperrors.Annotate("import training log insert", err)
	}
	if err := db.WithContext(ctx).Model(&shoe).
		Update("total_distance", gorm.Expr("total_distance + ?", in.Distance)).Error; err != nil {
		return apperrors.Annotate("import training log update shoe km", err)
	}
	return nil
}

func importCSVTrainingLogs(
	ctx context.Context,
	db *gorm.DB,
	userID domain.UserID,
	csvPath string,
	options TrainingLogImportOptions,
) (int, error) {
	return importio.StreamTrainingLogCSV(ctx, csvPath, options.MaxItems(), func(rec importio.TrainingLogCSVRecord) error {
		in, err := csvRecordToInput(rec)
		if err != nil {
			return err
		}
		return importOneTrainingLog(ctx, db, userID, in)
	})
}

func importBodyReader(c *gin.Context) (io.Reader, func(), error) {
	if file, err := c.FormFile("file"); err == nil {
		src, err := file.Open()
		if err != nil {
			return nil, func() {}, apperrors.BadRequest("アップロードファイルを開けませんでした", err)
		}
		return src, func() { _ = src.Close() }, nil
	}
	return c.Request.Body, func() {}, nil
}

func isCSVContentType(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
	return ct == "text/csv" || ct == "application/csv"
}

// handleImportTrainingLogsCSV は POST /training-logs + Content-Type: text/csv（または multipart file）用。
// CSV 例（1 行目ヘッダー必須）:
// training_date,distance,duration,pace,kind,shoe_id,memo
func (h *Handlers) handleImportTrainingLogsCSV(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	if file, _ := c.FormFile("file"); file == nil && !isCSVContentType(c.GetHeader("Content-Type")) {
		respondHTTPError(c, apperrors.BadRequest("multipart の file または Content-Type: text/csv で送信してください"))
		return
	}

	options, err := NewTrainingLogImportOptions(
		WithImportMaxItems(defaultImportMaxItems),
		WithImportAtomic(defaultImportAtomic),
	)
	if err != nil {
		respondHTTPError(c, apperrors.InternalMsg("取り込み設定が不正です", err))
		return
	}

	body, closeBody, err := importBodyReader(c)
	if err != nil {
		respondHTTPError(c, err)
		return
	}
	defer closeBody()

	path, removeTemp, err := importio.SpoolToTempFile(body, importio.DefaultMaxSpoolBytes, "training-logs-*.csv")
	if err != nil {
		if strings.Contains(err.Error(), "exceeds") {
			respondHTTPError(c, apperrors.BadRequest("CSV ファイルが大きすぎます", err))
			return
		}
		respondHTTPError(c, apperrors.InternalMsg("一時ファイルの作成に失敗しました", err))
		return
	}
	defer removeTemp()

	var createdCount int
	if options.Atomic() {
		err = database.WithTx(c.Request.Context(), h.dbCtx(c), func(tx *gorm.DB) error {
			var importErr error
			createdCount, importErr = importCSVTrainingLogs(c.Request.Context(), tx, userID, path, options)
			return importErr
		})
	} else {
		createdCount, err = importCSVTrainingLogs(c.Request.Context(), h.dbCtx(c), userID, path, options)
	}
	if err != nil {
		respondPreferVisible(c, err, "走行ログの CSV 取込に失敗しました")
		return
	}

	h.invalidateMonthlyAnalysisCache(c, userID)
	c.JSON(http.StatusOK, gin.H{
		"message":       "走行ログを CSV から取り込みました",
		"created_count": createdCount,
	})
	BroadcastTrainingLogEvent(userID, "training_logs_imported", map[string]any{
		"created_count": createdCount,
	})
}

// ImportTrainingLogsCSV はレガシー別名（POST /training-logs + CSV 本文と同等）。
func (h *Handlers) ImportTrainingLogsCSV(c *gin.Context) {
	h.handleImportTrainingLogsCSV(c)
}
