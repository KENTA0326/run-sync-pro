package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.WriteError(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "認証トークンが必要です")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("your_secret_key"), nil
		})

		if err != nil || !token.Valid {
			response.WriteError(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "無効なトークンです")
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID, ok := domain.ParseUserID(claims["user_id"])
		if !ok || !userID.IsValid() {
			response.WriteError(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "無効なトークンです")
			c.Abort()
			return
		}
		ctx := domain.WithUserID(c.Request.Context(), userID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
