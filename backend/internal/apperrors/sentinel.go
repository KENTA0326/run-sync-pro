package apperrors

import "errors"

// Sentinel は errors.Is で種別を判定するために使う。
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInternal     = errors.New("internal")
)
