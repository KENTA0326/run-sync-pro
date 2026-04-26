package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/KENTA0326/run-sync-pro/database"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// --- 共通: コンテキストから userID を取り出すヘルパ ---

func getUserIDFromContext(c *gin.Context) (uint, bool) {
	raw, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	// JWTのMapClaimsから来るので float64 想定
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

// POST /auth/shoes
func CreateShoe(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できません"})
		return
	}

	var input createShoeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	purchaseDate, err := time.Parse("2006-01-02", input.PurchaseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "purchase_date は YYYY-MM-DD 形式で指定してください"})
		return
	}

	shoe := model.Shoe{
		UserID:       userID,
		Brand:        input.Brand,
		Model:        input.Model,
		PurchaseDate: purchaseDate,
	}

	if err := database.DB.Create(&shoe).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "シューズの登録に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, shoe)
}

// GET /auth/shoes
func ListShoes(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できません"})
		return
	}

	var shoes []model.Shoe
	if err := database.DB.
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("purchase_date DESC, id DESC").
		Find(&shoes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "シューズ一覧の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, shoes)
}

// DELETE /auth/shoes/:id （論理削除: is_active=false）
func DeleteShoe(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できません"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不正なIDです"})
		return
	}

	var shoe model.Shoe
	if err := database.DB.
		Where("id = ? AND user_id = ?", id, userID).
		First(&shoe).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "シューズが見つかりません"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "シューズの取得に失敗しました"})
		return
	}

	if !shoe.IsActive {
		// すでに非アクティブなら 200 を返してもよい
		c.JSON(http.StatusOK, gin.H{"message": "すでに削除済みです"})
		return
	}

	if err := database.DB.Model(&shoe).Update("is_active", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "シューズの削除に失敗しました"})
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

// POST /auth/training-logs
func CreateTrainingLog(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できません"})
		return
	}

	var input createTrainingLogInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "入力内容が正しくありません。日付・距離・時間・種類・シューズを確認してください。",
		})
		return
	}

	trainingDate, err := time.Parse("2006-01-02", input.TrainingDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "training_date は YYYY-MM-DD 形式で指定してください"})
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// シューズの存在と所有者チェック
		var shoe model.Shoe
		if err := tx.Where("id = ? AND user_id = ?", input.ShoeID, userID).First(&shoe).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return gin.Error{Err: err, Type: gin.ErrorTypePublic, Meta: "シューズが見つかりません"}
			}
			return err
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
			return err
		}

		// 累計距離を加算
		if err := tx.Model(&shoe).
			Update("total_distance", gorm.Expr("total_distance + ?", input.Distance)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if ginErr, ok := err.(gin.Error); ok && ginErr.Type == gin.ErrorTypePublic {
			if msg, ok := ginErr.Meta.(string); ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "走行ログの保存に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "走行ログを保存しました"})
}

// GET /auth/training-logs
func ListTrainingLogs(c *gin.Context) {
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

	c.JSON(http.StatusOK, logs)
}
