package response

import "github.com/gin-gonic/gin"

// ErrorBody は統一エラーレスポンス `{ "error": { "code", "message" } }`。
type ErrorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// WriteError は統一形式で JSON エラーを返す。
func WriteError(c *gin.Context, status int, code, message string) {
	var body ErrorBody
	body.Error.Code = code
	body.Error.Message = message
	c.JSON(status, body)
}
