package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/config"
	"github.com/KENTA0326/run-sync-pro/internal/mail"
	"github.com/KENTA0326/run-sync-pro/internal/logging"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/KENTA0326/run-sync-pro/service"
	"github.com/gin-gonic/gin"
)

const passwordResetRequestMessage = "登録済みのメールアドレスの場合、パスワード再設定の案内を送信しました。"

func passwordResetTTL() time.Duration {
	minStr := config.ResolveString("", "PASSWORD_RESET_TOKEN_TTL_MINUTES", "60")
	min, err := strconv.Atoi(minStr)
	if err != nil || min <= 0 {
		min = 60
	}
	return time.Duration(min) * time.Minute
}

func frontendBaseURL() string {
	return config.ResolveString("", "FRONTEND_BASE_URL", "http://localhost:3001")
}

func passwordResetDevExposeLink() bool {
	return config.ResolveString("", "PASSWORD_RESET_DEV_EXPOSE_LINK", "true") == "true"
}

// RequestPasswordReset はリセット用リンクを発行する。
// POST /api/v1/auth/password-reset/request （レガシー: POST /auth/password-reset/request）
func (h *Handlers) RequestPasswordReset(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		respondHTTPError(c, apperrors.BadRequest("メールアドレスを確認してください", err))
		return
	}

	email, err := model.NewEmail(input.Email)
	if err != nil {
		// 存在有無を漏らさないため成功レスポンスと同形
		c.JSON(http.StatusOK, gin.H{"message": passwordResetRequestMessage})
		return
	}

	var user model.User
	if err := h.dbCtx(c).Where("email = ?", email.String()).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": passwordResetRequestMessage})
		return
	}

	plainToken, err := service.IssuePasswordResetToken(h.dbCtx(c), user.ID, passwordResetTTL())
	if err != nil {
		respondHTTPError(c, apperrors.InternalMsg("リセット手続きの開始に失敗しました", err))
		return
	}

	resetURL := service.BuildPasswordResetURL(frontendBaseURL(), plainToken)
	smtpCfg := mail.LoadSMTPConfig()

	if smtpCfg.IsEnabled() {
		if err := smtpCfg.SendPasswordResetEmail(email.String(), resetURL); err != nil {
			logging.FromGin(c).Error("password_reset_email_failed", slog.Any("err", err))
			respondHTTPError(c, apperrors.InternalMsg("メール送信に失敗しました", err))
			return
		}
	} else {
		logging.FromGin(c).Info("password_reset_link_issued",
			slog.Uint64("user_id", uint64(user.ID)),
			slog.String("reset_url", resetURL),
			slog.String("note", "SMTP 未設定のためログに URL を出力（本番では SMTP を設定）"),
		)
	}

	resp := gin.H{"message": passwordResetRequestMessage}
	if !smtpCfg.IsEnabled() && passwordResetDevExposeLink() {
		resp["reset_url"] = resetURL
	}
	c.JSON(http.StatusOK, resp)
}

// ConfirmPasswordReset はトークンで新パスワードを設定する。
// POST /api/v1/auth/password-reset/confirm （レガシー: POST /auth/password-reset/confirm）
func (h *Handlers) ConfirmPasswordReset(c *gin.Context) {
	var input struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		respondHTTPError(c, apperrors.BadRequest("入力内容を確認してください", err))
		return
	}

	userID, err := service.ConsumePasswordResetToken(h.dbCtx(c), input.Token)
	if err != nil {
		if errors.Is(err, service.ErrPasswordResetInvalid) {
			respondHTTPError(c, apperrors.BadRequest("リセットリンクが無効または期限切れです", err))
			return
		}
		respondHTTPError(c, apperrors.InternalMsg("リセット処理に失敗しました", err))
		return
	}

	hashed, err := h.auth.HashPassword(input.NewPassword)
	if err != nil {
		respondHTTPError(c, apperrors.InternalMsg("パスワードの更新に失敗しました", err))
		return
	}

	if err := h.dbCtx(c).Model(&model.User{}).Where("id = ?", userID).Update("password", hashed).Error; err != nil {
		respondHTTPError(c, apperrors.InternalMsg("パスワードの更新に失敗しました", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "パスワードを更新しました。新しいパスワードでログインしてください。"})
}
