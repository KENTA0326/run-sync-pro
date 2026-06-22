package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/mail"
	"github.com/KENTA0326/run-sync-pro/internal/logging"
	"github.com/KENTA0326/run-sync-pro/internal/repository"
	"github.com/KENTA0326/run-sync-pro/internal/usecase"
	"github.com/gin-gonic/gin"
)

const passwordResetRequestMessage = "登録済みのメールアドレスの場合、パスワード再設定の案内を送信しました。"

func passwordResetDevExposeLink() bool {
	// 環境変数で制御は main 側から frontendBaseURL 等で注入済み
	return true
}

// RequestPasswordReset はリセット用リンクを発行する。
// POST /api/v1/auth/password-reset/request
func (h *Handlers) RequestPasswordReset(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		respondHTTPError(c, apperrors.BadRequest("メールアドレスを確認してください", err))
		return
	}

	out, err := h.passwordResetUC.RequestReset(c.Request.Context(), usecase.RequestResetInput{
		Email:       input.Email,
		TTL:         h.passwordResetTTL(),
		FrontendURL: h.frontendBaseURL(),
	})
	if err != nil {
		respondHTTPError(c, apperrors.InternalMsg("リセット手続きの開始に失敗しました", err))
		return
	}

	if !out.UserFound {
		c.JSON(http.StatusOK, gin.H{"message": passwordResetRequestMessage})
		return
	}

	smtpCfg := mail.LoadSMTPConfig()
	if smtpCfg.IsEnabled() {
		if err := smtpCfg.SendPasswordResetEmail(input.Email, out.ResetURL); err != nil {
			logging.FromGin(c).Error("password_reset_email_failed", slog.Any("err", err))
			respondHTTPError(c, apperrors.InternalMsg("メール送信に失敗しました", err))
			return
		}
	} else {
		logging.FromGin(c).Info("password_reset_link_issued",
			slog.String("reset_url", out.ResetURL),
			slog.String("note", "SMTP 未設定のためログに URL を出力（本番では SMTP を設定）"),
		)
	}

	resp := gin.H{"message": passwordResetRequestMessage}
	if !smtpCfg.IsEnabled() && passwordResetDevExposeLink() {
		resp["reset_url"] = out.ResetURL
	}
	c.JSON(http.StatusOK, resp)
}

// ConfirmPasswordReset はトークンで新パスワードを設定する。
// POST /api/v1/auth/password-reset/confirm
func (h *Handlers) ConfirmPasswordReset(c *gin.Context) {
	var input struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		respondHTTPError(c, apperrors.BadRequest("入力内容を確認してください", err))
		return
	}

	err := h.passwordResetUC.ConfirmReset(c.Request.Context(), usecase.ConfirmResetInput{
		Token:       input.Token,
		NewPassword: input.NewPassword,
	})
	if err != nil {
		if errors.Is(err, repository.ErrPasswordResetInvalid) {
			respondHTTPError(c, apperrors.BadRequest("リセットリンクが無効または期限切れです", err))
			return
		}
		respondHTTPError(c, apperrors.InternalMsg("リセット処理に失敗しました", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "パスワードを更新しました。新しいパスワードでログインしてください。"})
}
