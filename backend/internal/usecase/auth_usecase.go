package usecase

import (
	"context"
	"errors"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
)

// AuthTokenGenerator はトークン生成の抽象（infrastructure 層で実装）。
type AuthTokenGenerator interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
	GenerateToken(userID uint) (string, error)
}

// SignUpInput はユーザー登録の入力。
type SignUpInput struct {
	Name     string
	Email    string
	Password string
}

// SignUpOutput はユーザー登録の出力。
type SignUpOutput struct {
	UserID uint
}

// LoginInput はログインの入力。
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput はログインの出力。
type LoginOutput struct {
	Token string
}

// AuthUseCase は認証関連のユースケース。
type AuthUseCase struct {
	repo domain.UserRepository
	auth AuthTokenGenerator
}

// NewAuthUseCase は AuthUseCase を生成する。
func NewAuthUseCase(repo domain.UserRepository, auth AuthTokenGenerator) *AuthUseCase {
	return &AuthUseCase{repo: repo, auth: auth}
}

// SignUp はユーザー新規登録を行う。
func (uc *AuthUseCase) SignUp(ctx context.Context, input SignUpInput) (*SignUpOutput, error) {
	email, err := domain.NewEmail(input.Email)
	if err != nil {
		return nil, domain.ErrInvalidEmail
	}

	exists, err := uc.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrUserDuplicate
	}

	hashed, err := uc.auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:     input.Name,
		Email:    email,
		Password: hashed,
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &SignUpOutput{UserID: user.ID}, nil
}

// Login はログイン処理を行う。
func (uc *AuthUseCase) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	email, err := domain.NewEmail(input.Email)
	if err != nil {
		return nil, domain.ErrInvalidEmail
	}

	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	if !uc.auth.CheckPassword(input.Password, user.Password) {
		return nil, domain.ErrUserNotFound
	}

	token, err := uc.auth.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{Token: token}, nil
}
