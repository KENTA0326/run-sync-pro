package handler

import (
	"errors"
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/usecase"
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

	out, err := h.authUC.SignUp(c.Request.Context(), usecase.SignUpInput{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmail):
			respondHTTPError(c, apperrors.BadRequest("メールアドレスの形式が正しくありません", err))
		case errors.Is(err, domain.ErrUserDuplicate):
			respondHTTPError(c, apperrors.ConflictMsg("このメールアドレスは既に登録されています"))
		default:
			respondHTTPError(c, apperrors.InternalMsg("ユーザー作成処理に失敗しました", apperrors.Annotate("signup usecase", err)))
		}
		return
	}
	_ = out
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

	out, err := h.authUC.Login(c.Request.Context(), usecase.LoginInput{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmail):
			respondHTTPError(c, apperrors.BadRequest("メールアドレスの形式が正しくありません", err))
		case errors.Is(err, domain.ErrUserNotFound):
			respondHTTPError(c, apperrors.UnauthorizedMsg("メールアドレスまたはパスワードが正しくありません"))
		default:
			respondHTTPError(c, apperrors.InternalMsg("ログイン処理に失敗しました", apperrors.Annotate("login usecase", err)))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": out.Token})
}

// CurrentUser は JWT からユーザー ID を返す（認証確認・デバッグ用）。
// ルート: GET /api/v1/users/me （レガシー: GET /auth/me）。
func (h *Handlers) CurrentUser(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		respondHTTPError(c, apperrors.UnauthorizedMsg("ユーザー情報を取得できません"))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID.Uint(),
		"message": "認証に成功しています！",
	})
}
