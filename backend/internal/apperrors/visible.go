package apperrors

import (
	"fmt"
	"net/http"
)

// Visible はユーザー向け HTTP 応答（文言とステータス）を errors.As で取り出す。
type Visible struct {
	Msg   string
	HTTP  int
	cause error
}

func (v *Visible) Error() string {
	if v.Msg != "" {
		return v.Msg
	}
	if v.cause != nil {
		return v.cause.Error()
	}
	return "error"
}

func (v *Visible) Unwrap() error {
	return v.cause
}

func newVisible(status int, msg string, inner error) error {
	return fmt.Errorf("%w", &Visible{Msg: msg, HTTP: status, cause: inner})
}

// BadRequest は 400。InvalidInput を連鎖に載せる。
func BadRequest(msg string, cause ...error) error {
	var inner error = ErrInvalidInput
	if len(cause) > 0 && cause[0] != nil {
		inner = fmt.Errorf("%w: %w", ErrInvalidInput, cause[0])
	}
	return newVisible(http.StatusBadRequest, msg, inner)
}

// UnauthorizedMsg は 401。
func UnauthorizedMsg(msg string) error {
	return newVisible(http.StatusUnauthorized, msg, ErrUnauthorized)
}

// NotFoundMsg は 404。NotFound センチネルを連鎖に載せる。
func NotFoundMsg(msg string, cause ...error) error {
	var inner error = ErrNotFound
	if len(cause) > 0 && cause[0] != nil {
		inner = fmt.Errorf("%w: %w", ErrNotFound, cause[0])
	}
	return newVisible(http.StatusNotFound, msg, inner)
}

// ConflictMsg は 409。
func ConflictMsg(msg string, cause ...error) error {
	var inner error = ErrConflict
	if len(cause) > 0 && cause[0] != nil {
		inner = fmt.Errorf("%w: %w", ErrConflict, cause[0])
	}
	return newVisible(http.StatusConflict, msg, inner)
}

// InternalMsg は 500。Internal センチネルでラップする。
func InternalMsg(msg string, cause error) error {
	if cause == nil {
		cause = ErrInternal
	}
	inner := fmt.Errorf("%w: %w", ErrInternal, cause)
	return newVisible(http.StatusInternalServerError, msg, inner)
}
