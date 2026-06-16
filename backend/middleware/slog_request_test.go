package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSlogRequestMiddleware_propagates_request_id(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SlogRequestMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(headerRequestID, "abc-123")
	r.ServeHTTP(w, req)

	if got := w.Header().Get(headerRequestID); got != "abc-123" {
		t.Fatalf("response X-Request-ID=%q want abc-123", got)
	}
}

func TestSlogRequestMiddleware_generates_id(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SlogRequestMiddleware())
	r.GET("/p", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/p", nil))
	id := w.Header().Get(headerRequestID)
	if len(id) < 8 {
		t.Fatalf("expected generated id, got %q", id)
	}
}
