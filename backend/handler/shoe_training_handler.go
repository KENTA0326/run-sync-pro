package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/timeutil"
	"github.com/KENTA0326/run-sync-pro/internal/usecase"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/KENTA0326/run-sync-pro/internal/domainservice"
	"github.com/gin-gonic/gin"
)

// --- シューズ管理 ---

type shoeResponse struct {
	ID            uint    `json:"id"`
	UserID        uint    `json:"user_id"`
	Brand         string  `json:"brand"`
	Model         string  `json:"model"`
	PurchaseDate  string  `json:"purchase_date"`
	TotalDistance float64 `json:"total_distance"`
	IsActive      bool    `json:"is_active"`
}

func toShoeResponse(s *domain.Shoe) shoeResponse {
	return shoeResponse{
		ID:            s.ID,
		UserID:        s.UserID,
		Brand:         s.Brand,
		Model:         s.Model,
		PurchaseDate:  s.PurchaseDate.Format("2006-01-02"),
		TotalDistance: s.TotalDistance,
		IsActive:      s.IsActive,
	}
}

type listShoesResponse struct {
	Shoes      []shoeResponse `json:"shoes"`
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
	View string `form:"view" binding:"omitempty,oneof=formatted"`
}

type updateTrainingLogKindInput struct {
	Kind int `json:"kind" binding:"required,min=0,max=4"`
}

type createShoeInput struct {
	Brand        string `json:"brand" binding:"required"`
	Model        string `json:"model" binding:"required"`
	PurchaseDate string `json:"purchase_date" binding:"required,caldate"`
}

// POST /api/v1/shoes
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
		respondHTTPError(c, apperrors.BadRequest("purchase_date を確認してください", err))
		return
	}

	shoe, err := h.shoeUC.Create(c.Request.Context(), usecase.CreateShoeInput{
		UserID:       userID.Uint(),
		Brand:        input.Brand,
		Model:        input.Model,
		PurchaseDate: purchaseDate,
	})
	if err != nil {
		respondHTTPError(c, apperrors.InternalMsg("シューズの登録に失敗しました", apperrors.Annotate("shoe usecase create", err)))
		return
	}

	c.JSON(http.StatusOK, toShoeResponse(shoe))
}

// GET /api/v1/shoes
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

	out, err := h.shoeUC.List(c.Request.Context(), userID.Uint(), limit, offset)
	if err != nil {
		respondHTTPError(c, apperrors.InternalMsg("シューズ一覧の取得に失敗しました", apperrors.Annotate("shoe usecase list", err)))
		return
	}

	shoes := make([]shoeResponse, len(out.Shoes))
	for i, s := range out.Shoes {
		shoes[i] = toShoeResponse(s)
	}

	c.JSON(http.StatusOK, listShoesResponse{
		Shoes:  shoes,
		Brands: out.Brands,
		Pagination: PaginationMeta{
			Limit:  limit,
			Offset: offset,
			Total:  out.Total,
		},
	})
}

// GET /api/v1/shoes/:id
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

	shoe, err := h.shoeUC.GetByID(c.Request.Context(), uri.ID, userID.Uint())
	if err != nil {
		if errors.Is(err, domain.ErrShoeNotFound) {
			respondHTTPError(c, apperrors.NotFoundMsg("シューズが見つかりません", err))
			return
		}
		respondHTTPError(c, apperrors.InternalMsg("シューズの取得に失敗しました", apperrors.Annotate("shoe usecase get", err)))
		return
	}

	c.JSON(http.StatusOK, toShoeResponse(shoe))
}

// DELETE /api/v1/shoes/:id
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

	if err := h.shoeUC.Delete(c.Request.Context(), uri.ID, userID.Uint()); err != nil {
		if errors.Is(err, domain.ErrShoeNotFound) {
			respondHTTPError(c, apperrors.NotFoundMsg("シューズが見つかりません", err))
			return
		}
		respondHTTPError(c, apperrors.InternalMsg("シューズの削除に失敗しました", apperrors.Annotate("shoe usecase delete", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "シューズを削除しました"})
}

// --- 走行ログ管理 ---

type createTrainingLogInput struct {
	TrainingDate string  `json:"training_date" binding:"required,caldate,caldate_not_future"`
	Distance     float64 `json:"distance" binding:"required,gt=0"`
	Duration     int     `json:"duration" binding:"required,gt=0"`
	Pace         string  `json:"pace" binding:"required"`
	Memo         string  `json:"memo"`
	Kind         int     `json:"kind" binding:"gte=0,lte=4"`
	ShoeID       uint    `json:"shoe_id" binding:"required"`
}

// POST /api/v1/training-logs
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

	if err := h.trainingLogUC.Create(c.Request.Context(), usecase.CreateTrainingLogInput{
		UserID:       userID.Uint(),
		TrainingDate: trainingDate,
		Distance:     input.Distance,
		Duration:     input.Duration,
		Pace:         input.Pace,
		Memo:         input.Memo,
		Kind:         input.Kind,
		ShoeID:       input.ShoeID,
	}); err != nil {
		if errors.Is(err, domain.ErrShoeNotFound) {
			respondHTTPError(c, apperrors.NotFoundMsg("シューズが見つかりません", err))
			return
		}
		respondHTTPError(c, apperrors.InternalMsg("走行ログの保存に失敗しました", apperrors.Annotate("training log usecase create", err)))
		return
	}

	h.invalidateMonthlyAnalysisCache(c, userID)
	c.JSON(http.StatusOK, gin.H{"message": "走行ログを保存しました"})
}

type paginatedTrainingLogsResponse struct {
	Items      []model.TrainingLog `json:"items"`
	Pagination PaginationMeta      `json:"pagination"`
}

// GET /api/v1/training-logs
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

	out, err := h.trainingLogUC.List(c.Request.Context(), userID.Uint(), limit, offset)
	if err != nil {
		respondHTTPError(c, apperrors.InternalMsg("走行ログの取得に失敗しました", apperrors.Annotate("training log usecase list", err)))
		return
	}

	// 既存 model.TrainingLog 互換のレスポンスを維持（DB→model レガシー経由）
	// TODO: 将来的にドメインモデルベースのレスポンスに統一
	base := h.dbCtx(c).Model(&model.TrainingLog{}).Where("user_id = ?", userID.Uint())
	var logs []model.TrainingLog
	if err := model.PreloadTrainingLogAssociations(base).
		Order("training_date DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("走行ログの取得に失敗しました", apperrors.Annotate("db list training logs", err)))
		return
	}

	meta := PaginationMeta{Limit: limit, Offset: offset, Total: out.Total}
	if q.View == "formatted" {
		c.JSON(http.StatusOK, gin.H{
			"items":      domainservice.MapSlice(logs, toTrainingLogFormattedResponse),
			"pagination": meta,
		})
		return
	}

	c.JSON(http.StatusOK, paginatedTrainingLogsResponse{
		Items:      logs,
		Pagination: meta,
	})
}

// GET /api/v1/training-logs/:id
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

// PATCH /api/v1/training-logs/:id/kind
func (h *Handlers) UpdateTrainingLogKind(c *gin.Context) {
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

	var input updateTrainingLogKindInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondHTTPError(c, apperrors.BadRequest("種類(kind)は 0-4 の範囲で指定してください", err))
		return
	}

	res := h.dbCtx(c).
		Model(&model.TrainingLog{}).
		Where("id = ? AND user_id = ?", uri.ID, userID.Uint()).
		Update("kind", input.Kind)
	if res.Error != nil {
		respondHTTPError(c, apperrors.InternalMsg("走行ログの種類更新に失敗しました", apperrors.Annotate("db update training log kind", res.Error)))
		return
	}
	if res.RowsAffected == 0 {
		respondHTTPError(c, apperrors.NotFoundMsg("走行ログが見つかりません"))
		return
	}

	h.invalidateMonthlyAnalysisCache(c, userID)
	c.JSON(http.StatusOK, gin.H{"message": "走行ログの種類を更新しました"})
}
