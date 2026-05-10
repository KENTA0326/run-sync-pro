package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/KENTA0326/run-sync-pro/internal/timeutil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// --- 共通: コンテキストから userID を取り出すヘルパ ---

func getUserIDFromContext(c *gin.Context) (uint, bool) {
	raw, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		return uint(v), true
	case uint:
		return v, true
	case int:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	default:
		return 0, false
	}
}

// --- シューズ管理 ---

type createShoeInput struct {
	Brand        string `json:"brand" binding:"required"`
	Model        string `json:"model" binding:"required"`
	PurchaseDate string `json:"purchase_date" binding:"required"` // "2006-01-02" 形式を想定
}

// POST /api/v1/shoes （レガシー: POST /auth/shoes）
func (h *Handlers) CreateShoe(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
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
		respondHTTPError(c, apperrors.BadRequest("purchase_date は YYYY-MM-DD 形式で指定してください", err))
		return
	}

	shoe := model.Shoe{
		UserID:       userID,
		Brand:        input.Brand,
		Model:        input.Model,
		PurchaseDate: purchaseDate,
	}

	if err := h.db.Create(&shoe).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("シューズの登録に失敗しました", apperrors.Annotate("db create shoe", err)))
		return
	}

	c.JSON(http.StatusOK, shoe)
}

// GET /api/v1/shoes （レガシー: GET /auth/shoes）
func (h *Handlers) ListShoes(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	var shoes []model.Shoe
	if err := h.db.
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("purchase_date DESC, id DESC").
		Find(&shoes).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("シューズ一覧の取得に失敗しました", apperrors.Annotate("db list shoes", err)))
		return
	}

	c.JSON(http.StatusOK, shoes)
}

// DELETE /api/v1/shoes/:id （論理削除: is_active=false、レガシー: DELETE /auth/shoes/:id）
func (h *Handlers) DeleteShoe(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		respondHTTPError(c, apperrors.BadRequest("不正なIDです", err))
		return
	}

	var shoe model.Shoe
	if err := h.db.
		Where("id = ? AND user_id = ?", id, userID).
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

	if err := h.db.Model(&shoe).Update("is_active", false).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("シューズの削除に失敗しました", apperrors.Annotate("db update shoe inactive", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "シューズを削除しました"})
}

// --- 走行ログ管理 ---

type createTrainingLogInput struct {
	TrainingDate string  `json:"training_date" binding:"required"` // "2006-01-02"
	Distance     float64 `json:"distance" binding:"required,gt=0"`
	Duration     int     `json:"duration" binding:"required,gt=0"` // 秒
	Pace         string  `json:"pace" binding:"required"`
	Memo         string  `json:"memo"`
	Kind         int     `json:"kind" binding:"gte=0,lte=3"` // 0:ジョグ, 1:LSD, 2:ペース走, 3:インターバル
	ShoeID       uint    `json:"shoe_id" binding:"required"` // 使用シューズ
}

// POST /api/v1/training-logs （レガシー: POST /auth/training-logs）
func (h *Handlers) CreateTrainingLog(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	var input createTrainingLogInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondHTTPError(c, apperrors.BadRequest("入力内容が正しくありません。日付・距離・時間・種類・シューズを確認してください。", err))
		return
	}

	trainingDate, err := timeutil.ParseCalendarDate(input.TrainingDate)
	if err != nil {
		respondHTTPError(c, apperrors.BadRequest("training_date は YYYY-MM-DD 形式で指定してください", err))
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		var shoe model.Shoe
		if err := tx.Where("id = ? AND user_id = ?", input.ShoeID, userID).First(&shoe).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.NotFoundMsg("シューズが見つかりません", err)
			}
			return apperrors.Annotate("create training log: shoe lookup", err)
		}

		log := model.TrainingLog{
			UserID:       userID,
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

	c.JSON(http.StatusOK, gin.H{"message": "走行ログを保存しました"})
}

// GET /api/v1/training-logs （レガシー: GET /auth/training-logs）
func (h *Handlers) ListTrainingLogs(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}

	var logs []model.TrainingLog
	if err := h.db.
		Where("user_id = ?", userID).
		Preload("Shoe").
		Order("training_date DESC, id DESC").
		Find(&logs).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("走行ログの取得に失敗しました", apperrors.Annotate("db list training logs", err)))
		return
	}

	c.JSON(http.StatusOK, logs)
}
