package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/importio"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const trainingLogExportBatchSize = 200

func trainingLogToCSVRecord(log model.TrainingLog) importio.TrainingLogCSVRecord {
	return importio.TrainingLogCSVRecord{
		TrainingDate: log.TrainingDate.Format("2006-01-02"),
		Distance:     log.Distance,
		Duration:     log.Duration,
		Pace:         log.Pace,
		Kind:         log.Kind,
		ShoeID:       log.ShoeID,
		Memo:         log.Memo,
	}
}

// handleExportTrainingLogsCSV は GET /training-logs + Accept: text/csv 用の CSV ストリーム出力。
func (h *Handlers) handleExportTrainingLogsCSV(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	filename := fmt.Sprintf("training_logs_%s.csv", time.Now().Format("20060102"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	ctx := c.Request.Context()
	db := h.dbCtx(c).WithContext(ctx).Model(&model.TrainingLog{}).Where("user_id = ?", userID.Uint())

	written := 0
	maxRows := defaultImportMaxItems
	_, err := importio.StreamWriteTrainingLogCSV(ctx, c.Writer, func(emit func(importio.TrainingLogCSVRecord) error) error {
		var batch []model.TrainingLog
		return db.Order("training_date ASC, id ASC").FindInBatches(&batch, trainingLogExportBatchSize, func(_ *gorm.DB, _ int) error {
			for _, log := range batch {
				if written >= maxRows {
					return apperrors.BadRequest(fmt.Sprintf("エクスポート上限（%d件）を超えています", maxRows))
				}
				if err := emit(trainingLogToCSVRecord(log)); err != nil {
					return err
				}
				written++
			}
			return nil
		}).Error
	})
	if err != nil {
		// ヘッダ送信後は JSON エラーに差し替えられないことがある
		if !c.Writer.Written() {
			respondPreferVisible(c, err, "走行ログの CSV エクスポートに失敗しました")
		}
		return
	}

	c.Status(http.StatusOK)
}

// ExportTrainingLogsCSV はレガシー別名（GET /training-logs + Accept: text/csv と同等）。
func (h *Handlers) ExportTrainingLogsCSV(c *gin.Context) {
	h.handleExportTrainingLogsCSV(c)
}
