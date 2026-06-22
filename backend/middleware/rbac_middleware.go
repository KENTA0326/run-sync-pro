package middleware

import (
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type permissionMatch struct {
	Allowed bool `gorm:"column:allowed"`
}

// RequirePermission は user_roles -> role_permissions -> permissions で認可を判定する。
func RequirePermission(db *gorm.DB, resource domain.Resource, action domain.Action) gin.HandlerFunc {
	if !resource.IsValid() || !action.IsValid() {
		panic("middleware: invalid RBAC resource or action")
	}
	return func(c *gin.Context) {
		userID, ok := domain.UserIDFromRequest(c.Request)
		if !ok {
			response.WriteError(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "ユーザー情報を取得できません")
			c.Abort()
			return
		}

		var m permissionMatch
		err := db.WithContext(c.Request.Context()).Raw(`
			SELECT EXISTS (
				SELECT 1
				FROM user_roles ur
				JOIN role_permissions rp ON rp.role_id = ur.role_id
				JOIN permissions p ON p.id = rp.permission_id
				WHERE ur.user_id = ?
				  AND (
					(p.resource = ? AND p.action = ?)
					OR (p.resource = '*' AND p.action = '*')
				  )
			) AS allowed
		`, userID.Uint(), resource.String(), action.String()).Scan(&m).Error
		if err != nil {
			response.WriteError(c, http.StatusInternalServerError, apperrors.CodeInternal, "権限確認に失敗しました")
			c.Abort()
			return
		}
		if !m.Allowed {
			response.WriteError(c, http.StatusForbidden, apperrors.CodeForbidden, "この操作を実行する権限がありません")
			c.Abort()
			return
		}

		c.Next()
	}
}
