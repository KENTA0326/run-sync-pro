package handler

import (
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/response"
	"github.com/gin-gonic/gin"
)

func writeAPIError(c *gin.Context, status int, code, message string) {
	response.WriteError(c, status, code, message)
}

func codeFromVisible(v *apperrors.Visible) string {
	switch v.HTTP {
	case http.StatusBadRequest:
		return apperrors.CodeInvalidInput
	case http.StatusUnauthorized:
		return apperrors.CodeUnauthorized
	case http.StatusForbidden:
		return apperrors.CodeForbidden
	case http.StatusNotFound:
		return apperrors.CodeNotFound
	case http.StatusConflict:
		return apperrors.CodeConflict
	default:
		return apperrors.CodeInternal
	}
}
