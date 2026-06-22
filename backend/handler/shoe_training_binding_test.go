package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestShoeIDURI_ShouldBindUri_table(t *testing.T) {
	cases := []struct {
		name    string
		rawPath string // "/shoes/123"
		wantErr bool
		wantID  uint
	}{
		{name: "valid", rawPath: "/shoes/42", wantErr: false, wantID: 42},
		{name: "zero_invalid", rawPath: "/shoes/0", wantErr: true},
		{name: "not_a_number", rawPath: "/shoes/abc", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, tc.rawPath, nil)

			router := gin.New()
			router.DELETE("/shoes/:id", func(c *gin.Context) {
				var uri shoeIDURI
				err := c.ShouldBindUri(&uri)
				if tc.wantErr {
					if err == nil {
						t.Fatal("expected bind error")
					}
					return
				}
				if err != nil {
					t.Fatalf("bind: %v", err)
				}
				if uri.ID != tc.wantID {
					t.Fatalf("ID=%d want %d", uri.ID, tc.wantID)
				}
			})

			router.ServeHTTP(w, req)
			if !tc.wantErr && w.Code != http.StatusOK {
				t.Fatalf("status=%d", w.Code)
			}
		})
	}
}

func TestTrainingLogIDURI_ShouldBindUri_table(t *testing.T) {
	cases := []struct {
		name    string
		rawPath string
		wantErr bool
		wantID  uint
	}{
		{name: "valid", rawPath: "/training-logs/99", wantErr: false, wantID: 99},
		{name: "zero_invalid", rawPath: "/training-logs/0", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.rawPath, nil)

			router := gin.New()
			router.GET("/training-logs/:id", func(c *gin.Context) {
				var uri trainingLogIDURI
				err := c.ShouldBindUri(&uri)
				if tc.wantErr {
					if err == nil {
						t.Fatal("expected bind error")
					}
					return
				}
				if err != nil {
					t.Fatalf("bind: %v", err)
				}
				if uri.ID != tc.wantID {
					t.Fatalf("ID=%d want %d", uri.ID, tc.wantID)
				}
			})

			router.ServeHTTP(w, req)
			if !tc.wantErr && w.Code != http.StatusOK {
				t.Fatalf("status=%d", w.Code)
			}
		})
	}
}

func TestListTrainingLogsQuery_ShouldBindQuery_table(t *testing.T) {
	cases := []struct {
		name    string
		rawQS   string
		wantErr bool
		wantLim int
	}{
		// limit 未指定時は binding 上 0（normalize() で defaultListLimit=50 になる）
		{name: "no_limit", rawQS: "", wantErr: false, wantLim: 0},
		{name: "limit_ok", rawQS: "?limit=10", wantErr: false, wantLim: 10},
		{name: "limit_too_large", rawQS: "?limit=501", wantErr: true},
		// limit=0 は omitempty のため min はかからず 0 のまま（全件扱いと同義）
		{name: "limit_zero_means_unset_semantics", rawQS: "?limit=0", wantErr: false, wantLim: 0},
		{name: "view_formatted_ok", rawQS: "?view=formatted", wantErr: false, wantLim: 0},
		{name: "view_invalid", rawQS: "?view=raw", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/logs"+tc.rawQS, nil)

			router := gin.New()
			router.GET("/logs", func(c *gin.Context) {
				var q listTrainingLogsQuery
				err := c.ShouldBindQuery(&q)
				if tc.wantErr {
					if err == nil {
						t.Fatal("expected bind error")
					}
					return
				}
				if err != nil {
					t.Fatalf("bind: %v", err)
				}
				if q.Limit != tc.wantLim {
					t.Fatalf("Limit=%d want %d", q.Limit, tc.wantLim)
				}
				c.Status(http.StatusOK)
			})

			router.ServeHTTP(w, req)
			if !tc.wantErr && w.Code != http.StatusOK {
				t.Fatalf("status=%d", w.Code)
			}
		})
	}
}
