package apperrors

// 機械可読な API エラーコード（HTTP ステータスと独立）。
const (
	CodeInvalidInput   = "INVALID_INPUT"
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeForbidden      = "FORBIDDEN"
	CodeNotFound       = "NOT_FOUND"
	CodeConflict       = "CONFLICT"
	CodeInternal       = "INTERNAL_ERROR"
)
