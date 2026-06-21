package domain

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

// ErrInvalidEmail はメールアドレスが不正な場合のドメインエラー。
var ErrInvalidEmail = errors.New("invalid email")

// Email は検証済みメールアドレスの値オブジェクト。
type Email struct {
	value string
}

// NewEmail は文字列を検証し Email を生成する。
func NewEmail(raw string) (Email, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return Email{}, ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return Email{}, ErrInvalidEmail
	}
	normalized := strings.TrimSpace(addr.Address)
	if normalized == "" {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: normalized}, nil
}

// String はメールアドレス文字列を返す。
func (e Email) String() string { return e.value }

// User はユーザーのドメインエンティティ（永続化非依存）。
type User struct {
	ID        uint
	Name      string
	Email     Email
	Password  string
	Role      int
	CreatedAt time.Time
	UpdatedAt time.Time
}
