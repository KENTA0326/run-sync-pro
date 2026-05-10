package handler

import (
	"errors"
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/gin-gonic/gin"
)

// SignUp はユーザー新規登録。
// POST /api/v1/auth/signup （レガシー: POST /signup）
func (h *Handlers) SignUp(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		respondHTTPError(c, apperrors.BadRequest("入力内容を確認してください", err))
		return
	}

	email, err := model.NewEmail(input.Email)
	if err != nil {
		if errors.Is(err, model.ErrInvalidEmail) {
			respondHTTPError(c, apperrors.BadRequest("メールアドレスの形式が正しくありません", err))
			return
		}
		respondHTTPError(c, apperrors.BadRequest("メールアドレスの入力を確認してください", err))
		return
	}

	var dupCount int64
	if err := h.db.Model(&model.User{}).Where("email = ?", email.String()).Count(&dupCount).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("ユーザー登録処理に失敗しました", apperrors.Annotate("count user by email", err)))
		return
	}
	if dupCount > 0 {
		respondHTTPError(c, apperrors.ConflictMsg("このメールアドレスは既に登録されています"))
		return
	}

	hashedPassword, err := h.auth.HashPassword(input.Password)
	if err != nil {
		respondHTTPError(c, apperrors.InternalMsg("ユーザー作成処理に失敗しました", apperrors.Annotate("hash password", err)))
		return
	}

	user := model.User{
		Name:     input.Name,
		Email:    email,
		Password: hashedPassword,
	}

	if err := h.db.Create(&user).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("ユーザー作成に失敗しました", apperrors.Annotate("db create user", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "登録完了！"})
}

// Login はログイン。
// POST /api/v1/auth/login （レガシー: POST /login）
func (h *Handlers) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		respondHTTPError(c, apperrors.BadRequest("入力内容を確認してください", err))
		return
	}

	email, err := model.NewEmail(input.Email)
	if err != nil {
		if errors.Is(err, model.ErrInvalidEmail) {
			respondHTTPError(c, apperrors.BadRequest("メールアドレスの形式が正しくありません", err))
			return
		}
		respondHTTPError(c, apperrors.BadRequest("メールアドレスの入力を確認してください", err))
		return
	}

	var user model.User
	if err := h.db.Where("email = ?", email.String()).First(&user).Error; err != nil {
		mapped := apperrors.FromGORM(err)
		if errors.Is(mapped, apperrors.ErrNotFound) {
			respondHTTPError(c, apperrors.UnauthorizedMsg("メールアドレスまたはパスワードが正しくありません"))
			return
		}
		respondHTTPError(c, apperrors.InternalMsg("ログイン処理に失敗しました", apperrors.Annotate("db find user", mapped)))
		return
	}

	if !h.auth.CheckPassword(input.Password, user.Password) {
		respondHTTPError(c, apperrors.UnauthorizedMsg("メールアドレスまたはパスワードが正しくありません"))
		return
	}

	token, err := h.auth.GenerateToken(user.ID)
	if err != nil {
		respondHTTPError(c, apperrors.InternalMsg("トークン発行に失敗しました", apperrors.Annotate("generate token", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// CurrentUser は JWT からユーザー ID を返す（認証確認・デバッグ用）。
// ルート: GET /api/v1/users/me （レガシー: GET /auth/me）。
func (h *Handlers) CurrentUser(c *gin.Context) {
	userID, _ := c.Get("userID")
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"message": "認証に成功しています！",
	})
}
