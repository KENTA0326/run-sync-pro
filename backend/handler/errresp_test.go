package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/logging"
	"github.com/KENTA0326/run-sync-pro/internal/testutil"
	"github.com/KENTA0326/run-sync-pro/internal/validation"
	"github.com/gin-gonic/gin"
)

func errBody(code, message string) map[string]any {
	return map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	}
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	logging.InitFromEnv()
	validation.RegisterGinBindingValidators()
	os.Exit(m.Run())
}

func TestRespondHTTPError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		err         error
		wantHandled bool
		wantCode    int
		wantJSON    map[string]any
	}{
		{
			name:        "nil_noop",
			err:         nil,
			wantHandled: false,
			wantCode:    http.StatusOK,
			wantJSON:    nil,
		},
		{
			name:        "visible_bad_request",
			err:         apperrors.BadRequest("入力が不正です"),
			wantHandled: true,
			wantCode:    http.StatusBadRequest,
			wantJSON:    errBody(apperrors.CodeInvalidInput, "入力が不正です"),
		},
		{
			name:        "visible_custom_not_found",
			err:         apperrors.NotFoundMsg("該当ログがありません"),
			wantHandled: true,
			wantCode:    http.StatusNotFound,
			wantJSON:    errBody(apperrors.CodeNotFound, "該当ログがありません"),
		},
		{
			name:        "sentinel_not_found_default_body",
			err:         apperrors.ErrNotFound,
			wantHandled: true,
			wantCode:    http.StatusNotFound,
			wantJSON:    errBody(apperrors.CodeNotFound, "リソースが見つかりません"),
		},
		{
			name:        "sentinel_unauthorized_default_body",
			err:         apperrors.ErrUnauthorized,
			wantHandled: true,
			wantCode:    http.StatusUnauthorized,
			wantJSON:    errBody(apperrors.CodeUnauthorized, "認証に失敗しました"),
		},
		{
			name:        "opaque_maps_to_internal",
			err:         errors.New("db exploded"),
			wantHandled: true,
			wantCode:    http.StatusInternalServerError,
			wantJSON:    errBody(apperrors.CodeInternal, "サーバー内部でエラーが発生しました"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			handled := respondHTTPError(c, tc.err)
			if handled != tc.wantHandled {
				t.Fatalf("handled=%v want %v", handled, tc.wantHandled)
			}
			if !tc.wantHandled {
				if rec.Body.Len() != 0 {
					t.Fatalf("unexpected body: %s", rec.Body.String())
				}
				return
			}
			testutil.AssertResponseJSON(t, rec, tc.wantCode, tc.wantJSON)
		})
	}
}

func TestRespondPreferVisible(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		err      error
		fallback string
		wantCode int
		wantJSON map[string]any
	}{
		{
			name:     "uses_visible_when_present",
			err:      apperrors.ConflictMsg("既に登録されています"),
			fallback: "ignored",
			wantCode: http.StatusConflict,
			wantJSON: errBody(apperrors.CodeConflict, "既に登録されています"),
		},
		{
			name:     "wraps_opaque_with_fallback_message",
			err:      errors.New("internal detail"),
			fallback: "処理に失敗しました",
			wantCode: http.StatusInternalServerError,
			wantJSON: errBody(apperrors.CodeInternal, "処理に失敗しました"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			respondPreferVisible(c, tc.err, tc.fallback)
			testutil.AssertResponseJSON(t, rec, tc.wantCode, tc.wantJSON)
		})
	}
}
