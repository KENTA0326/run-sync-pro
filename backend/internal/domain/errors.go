package domain

import "errors"

// ドメイン共通エラー。
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserDuplicate      = errors.New("user already exists")
	ErrShoeNotFound       = errors.New("shoe not found")
	ErrTrainingLogNotFound = errors.New("training log not found")
)
