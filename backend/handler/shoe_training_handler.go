package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/timeutil"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/KENTA0326/run-sync-pro/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// --- シューズ管理 ---

type listShoesResponse struct {
	Shoes      []model.Shoe   `json:"shoes"`
	Brands     []string       `json:"brands"`
	Pagination PaginationMeta `json:"pagination"`
}

// shoeIDURI は /shoes/:id のパスパラメータ（ShouldBindUri）。
type shoeIDURI struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// trainingLogIDURI は /training-logs/:id のパスパラメータ（ShouldBindUri）。
type trainingLogIDURI struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// listTrainingLogsQuery は一覧 GET の limit / offset / view。
type listTrainingLogsQuery struct {
	listPaginationQuery
	// view=formatted で日付・タイムスタンプを整形した JSON を返す（別 URL ではなく表現の切替）。
	View string `form:"view" binding:"omitempty,oneof=formatted"`
}

type createShoeInput struct {
	Brand        string `json:"brand" binding:"required"`
	Model        string `json:"model" binding:"required"`
	// caldate: JST カレンダー日として解釈できる "YYYY-MM-DD"（カスタム validator: internal/validation）
	PurchaseDate string `json:"purchase_date" binding:"required,caldate"`
}

// POST /api/v1/shoes （レガシー: POST /auth/shoes）
func (h *Handlers) CreateShoe(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	var input createShoeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondHTTPError(c, apperrors.BadRequest("入力内容を確認してください", err))
		return
	}

	purchaseDate, err := timeutil.ParseCalendarDate(input.PurchaseDate)
	if err != nil {
		// caldate タグで通常は弾かれる。ここは防御的。
		respondHTTPError(c, apperrors.BadRequest("purchase_date を確認してください", err))
		return
	}

	shoe := model.Shoe{
		UserID:       userID.Uint(),
		Brand:        input.Brand,
		Model:        input.Model,
		PurchaseDate: purchaseDate,
	}

	if err := h.dbCtx(c).Create(&shoe).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("シューズの登録に失敗しました", apperrors.Annotate("db create shoe", err)))
		return
	}

	c.JSON(http.StatusOK, shoe)
}

// GET /api/v1/shoes （レガシー: GET /auth/shoes）
func (h *Handlers) ListShoes(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	var q listPaginationQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		respondHTTPError(c, apperrors.BadRequest("クエリパラメータが不正です", err))
		return
	}
	limit, offset := q.normalize()

	base := h.dbCtx(c).Model(&model.Shoe{}).
		Where("user_id = ? AND is_active = ?", userID.Uint(), true)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("シューズ一覧の取得に失敗しました", apperrors.Annotate("db count shoes", err)))
		return
	}

	var shoes []model.Shoe
	if err := base.
		Order("purchase_date DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&shoes).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("シューズ一覧の取得に失敗しました", apperrors.Annotate("db list shoes", err)))
		return
	}

	var allForBrands []model.Shoe
	if err := h.dbCtx(c).
		Where("user_id = ? AND is_active = ?", userID.Uint(), true).
		Find(&allForBrands).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("シューズ一覧の取得に失敗しました", apperrors.Annotate("db list shoes brands", err)))
		return
	}

	c.JSON(http.StatusOK, listShoesResponse{
		Shoes:  shoes,
		Brands: service.DistinctShoeBrands(allForBrands),
		Pagination: PaginationMeta{
			Limit:  limit,
			Offset: offset,
			Total:  total,
		},
	})
}

// GET /api/v1/shoes/:id （レガシー: GET /auth/shoes/:id）
func (h *Handlers) GetShoe(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	var uri shoeIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		respondHTTPError(c, apperrors.BadRequest("ID が不正です", err))
		return
	}

	var shoe model.Shoe
	if err := h.dbCtx(c).
		Where("id = ? AND user_id = ? AND is_active = ?", uri.ID, userID.Uint(), true).
		First(&shoe).Error; err != nil {
		mapped := apperrors.FromGORM(err)
		if errors.Is(mapped, apperrors.ErrNotFound) {
			respondHTTPError(c, apperrors.NotFoundMsg("シューズが見つかりません", err))
			return
		}
		respondHTTPError(c, apperrors.InternalMsg("シューズの取得に失敗しました", apperrors.Annotate("db first shoe", mapped)))
		return
	}

	c.JSON(http.StatusOK, shoe)
}

// DELETE /api/v1/shoes/:id （論理削除: is_active=false、レガシー: DELETE /auth/shoes/:id）
func (h *Handlers) DeleteShoe(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	var uri shoeIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		respondHTTPError(c, apperrors.BadRequest("ID が不正です", err))
		return
	}

	var shoe model.Shoe
	if err := h.dbCtx(c).
		Where("id = ? AND user_id = ?", uri.ID, userID.Uint()).
		First(&shoe).Error; err != nil {
		mapped := apperrors.FromGORM(err)
		if errors.Is(mapped, apperrors.ErrNotFound) {
			respondHTTPError(c, apperrors.NotFoundMsg("シューズが見つかりません", err))
			return
		}
		respondHTTPError(c, apperrors.InternalMsg("シューズの取得に失敗しました", apperrors.Annotate("db first shoe", mapped)))
		return
	}

	if !shoe.IsActive {
		c.JSON(http.StatusOK, gin.H{"message": "すでに削除済みです"})
		return
	}

	if err := h.dbCtx(c).Model(&shoe).Update("is_active", false).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("シューズの削除に失敗しました", apperrors.Annotate("db update shoe inactive", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "シューズを削除しました"})
}

// --- 走行ログ管理 ---

type createTrainingLogInput struct {
	// caldate + caldate_not_future: 有効なカレンダー日かつ未来日不可（走行日のビジネスルール）
	TrainingDate string  `json:"training_date" binding:"required,caldate,caldate_not_future"`
	Distance     float64 `json:"distance" binding:"required,gt=0"`
	Duration     int     `json:"duration" binding:"required,gt=0"` // 秒
	Pace         string  `json:"pace" binding:"required"`
	Memo         string  `json:"memo"`
	Kind         int     `json:"kind" binding:"gte=0,lte=3"` // 0:ジョグ, 1:LSD, 2:ペース走, 3:インターバル
	ShoeID       uint    `json:"shoe_id" binding:"required"` // 使用シューズ
}

// POST /api/v1/training-logs （レガシー: POST /auth/training-logs）
// 単件: application/json オブジェクト / CSV: multipart file または text/csv / 一括: JSON 配列
func (h *Handlers) CreateTrainingLog(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	if wantsCSVRequest(c) {
		h.handleImportTrainingLogsCSV(c)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondHTTPError(c, apperrors.BadRequest("リクエスト本文を読み取れませんでした", err))
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	if trimmed := bytes.TrimSpace(body); len(trimmed) > 0 && trimmed[0] == '[' {
		h.handleImportTrainingLogsStream(c)
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	var input createTrainingLogInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondHTTPError(c, apperrors.BadRequest("入力内容が正しくありません。日付・距離・時間・種類・シューズを確認してください。", err))
		return
	}

	trainingDate, err := timeutil.ParseCalendarDate(input.TrainingDate)
	if err != nil {
		respondHTTPError(c, apperrors.BadRequest("training_date を確認してください", err))
		return
	}

	err = h.dbCtx(c).Transaction(func(tx *gorm.DB) error {
		var shoe model.Shoe
		if err := tx.Where("id = ? AND user_id = ?", input.ShoeID, userID.Uint()).First(&shoe).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.NotFoundMsg("シューズが見つかりません", err)
			}
			return apperrors.Annotate("create training log: shoe lookup", err)
		}

		log := model.TrainingLog{
			UserID:       userID.Uint(),
			TrainingDate: trainingDate,
			Distance:     input.Distance,
			Duration:     input.Duration,
			Pace:         input.Pace,
			Memo:         input.Memo,
			Kind:         input.Kind,
			ShoeID:       input.ShoeID,
		}

		if err := tx.Create(&log).Error; err != nil {
			return apperrors.Annotate("create training log: insert log", err)
		}

		if err := tx.Model(&shoe).
			Update("total_distance", gorm.Expr("total_distance + ?", input.Distance)).Error; err != nil {
			return apperrors.Annotate("create training log: update shoe distance", err)
		}

		return nil
	})

	if err != nil {
		respondPreferVisible(c, err, "走行ログの保存に失敗しました")
		return
	}

	h.invalidateMonthlyAnalysisCache(c, userID)
	c.JSON(http.StatusOK, gin.H{"message": "走行ログを保存しました"})
}

type paginatedTrainingLogsResponse struct {
	Items      []model.TrainingLog `json:"items"`
	Pagination PaginationMeta      `json:"pagination"`
}

// GET /api/v1/training-logs （?view=formatted / Accept: text/csv で CSV、レガシー: GET /auth/training-logs）
func (h *Handlers) ListTrainingLogs(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	if wantsCSVResponse(c) {
		h.handleExportTrainingLogsCSV(c)
		return
	}

	var q listTrainingLogsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		respondHTTPError(c, apperrors.BadRequest("クエリパラメータが不正です", err))
		return
	}
	limit, offset := q.normalize()

	base := h.dbCtx(c).Model(&model.TrainingLog{}).Where("user_id = ?", userID.Uint())

	var total int64
	if err := base.Count(&total).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("走行ログの取得に失敗しました", apperrors.Annotate("db count training logs", err)))
		return
	}

	var logs []model.TrainingLog
	if err := model.PreloadTrainingLogAssociations(base).
		Order("training_date DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("走行ログの取得に失敗しました", apperrors.Annotate("db list training logs", err)))
		return
	}

	meta := PaginationMeta{Limit: limit, Offset: offset, Total: total}
	if q.View == "formatted" {
		c.JSON(http.StatusOK, gin.H{
			"items":      service.MapSlice(logs, toTrainingLogFormattedResponse),
			"pagination": meta,
		})
		return
	}

	c.JSON(http.StatusOK, paginatedTrainingLogsResponse{
		Items:      logs,
		Pagination: meta,
	})
}

// GET /api/v1/training-logs/:id （レガシー: GET /auth/training-logs/:id）
func (h *Handlers) GetTrainingLog(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	var uri trainingLogIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		respondHTTPError(c, apperrors.BadRequest("ID が不正です", err))
		return
	}

	var log model.TrainingLog
	if err := model.PreloadTrainingLogAssociations(h.dbCtx(c)).
		Where("id = ? AND user_id = ?", uri.ID, userID.Uint()).
		First(&log).Error; err != nil {
		mapped := apperrors.FromGORM(err)
		if errors.Is(mapped, apperrors.ErrNotFound) {
			respondHTTPError(c, apperrors.NotFoundMsg("走行ログが見つかりません", err))
			return
		}
		respondHTTPError(c, apperrors.InternalMsg("走行ログの取得に失敗しました", apperrors.Annotate("db first training log", mapped)))
		return
	}

	c.JSON(http.StatusOK, log)
}
