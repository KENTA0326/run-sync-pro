package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/repository"
)

// ErrPasswordResetInvalid はトークン無効・期限切れのユースケースエラー。
var ErrPasswordResetInvalid = repository.ErrPasswordResetInvalid

// PasswordResetUseCase はパスワードリセットのユースケース。
type PasswordResetUseCase struct {
	userRepo  domain.UserRepository
	resetRepo domain.PasswordResetRepository
	auth      AuthTokenGenerator
}

func NewPasswordResetUseCase(
	userRepo domain.UserRepository,
	resetRepo domain.PasswordResetRepository,
	auth AuthTokenGenerator,
) *PasswordResetUseCase {
	return &PasswordResetUseCase{
		userRepo:  userRepo,
		resetRepo: resetRepo,
		auth:      auth,
	}
}

// RequestResetInput はリセット要求の入力。
type RequestResetInput struct {
	Email       string
	TTL         time.Duration
	FrontendURL string
}

// RequestResetOutput はリセット要求の出力。
type RequestResetOutput struct {
	ResetURL  string
	UserFound bool
}

// RequestReset はパスワードリセットトークンを発行する。
func (uc *PasswordResetUseCase) RequestReset(ctx context.Context, input RequestResetInput) (*RequestResetOutput, error) {
	email, err := domain.NewEmail(input.Email)
	if err != nil {
		return &RequestResetOutput{UserFound: false}, nil
	}

	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return &RequestResetOutput{UserFound: false}, nil
		}
		return nil, err
	}

	plainToken, err := uc.resetRepo.Issue(ctx, user.ID, input.TTL)
	if err != nil {
		return nil, err
	}

	resetURL := buildPasswordResetURL(input.FrontendURL, plainToken)
	return &RequestResetOutput{ResetURL: resetURL, UserFound: true}, nil
}

// ConfirmResetInput はリセット確定の入力。
type ConfirmResetInput struct {
	Token       string
	NewPassword string
}

// ConfirmReset はトークンを検証しパスワードを更新する。
func (uc *PasswordResetUseCase) ConfirmReset(ctx context.Context, input ConfirmResetInput) error {
	userID, err := uc.resetRepo.Consume(ctx, input.Token)
	if err != nil {
		return err
	}

	hashed, err := uc.auth.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	return uc.userRepo.UpdatePassword(ctx, userID, hashed)
}

func buildPasswordResetURL(frontendBase, plainToken string) string {
	return fmt.Sprintf("%s/reset-password?token=%s", trimTrailingSlash(frontendBase), plainToken)
}

func trimTrailingSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
