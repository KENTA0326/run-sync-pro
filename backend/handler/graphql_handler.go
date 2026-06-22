package handler

import (
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/graphqlapi"
	"github.com/KENTA0326/run-sync-pro/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
)

// GraphQL POST /api/v1/graphql （レガシー: POST /graphql）
// Body: { "query": "...", "variables": {} }
func (h *Handlers) GraphQL(c *gin.Context) {
	if _, ok := domain.UserIDFromRequest(c.Request); !ok {
		response.WriteError(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "認証が必要です")
		return
	}

	var body struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Query == "" {
		response.WriteError(c, http.StatusBadRequest, apperrors.CodeInvalidInput, "query が必要です")
		return
	}

	gql := graphqlapi.New(h.dbCtx(c))
	schema, err := gql.Schema()
	if err != nil {
		response.WriteError(c, http.StatusInternalServerError, apperrors.CodeInternal, "GraphQL スキーマの初期化に失敗しました")
		return
	}

	result := graphql.Do(graphql.Params{
		Schema:         schema,
		RequestString:  body.Query,
		VariableValues: body.Variables,
		Context:        c.Request.Context(),
	})
	if len(result.Errors) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"errors": result.Errors,
			"data":   result.Data,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result.Data})
}
