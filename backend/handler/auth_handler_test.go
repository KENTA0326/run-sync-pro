package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/testutil"
	"github.com/KENTA0326/run-sync-pro/internal/usecase"
	"github.com/gin-gonic/gin"
)

func jsonPOST(path string, body any) *http.Request {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestSignUp_success(t *testing.T) {
	t.Parallel()

	var gotInput usecase.SignUpInput
	h := newTestHandlers(&fakeAuthUC{
		signUp: func(_ context.Context, input usecase.SignUpInput) (*usecase.SignUpOutput, error) {
			gotInput = input
			return &usecase.SignUpOutput{UserID: 7}, nil
		},
	}, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/signup", map[string]string{
		"name":     "太郎",
		"email":    "taro@example.com",
		"password": "secret1",
	})

	h.SignUp(c)

	if gotInput.Name != "太郎" || gotInput.Email != "taro@example.com" || gotInput.Password != "secret1" {
		t.Fatalf("unexpected usecase input: %+v", gotInput)
	}
	testutil.AssertResponseJSON(t, rec, http.StatusOK, map[string]any{
		"message": "登録完了！",
	})
}

func TestSignUp_duplicateEmail(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(&fakeAuthUC{
		signUp: func(_ context.Context, _ usecase.SignUpInput) (*usecase.SignUpOutput, error) {
			return nil, domain.ErrUserDuplicate
		},
	}, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/signup", map[string]string{
		"name":     "太郎",
		"email":    "dup@example.com",
		"password": "secret1",
	})

	h.SignUp(c)

	testutil.AssertResponseJSON(t, rec, http.StatusConflict, map[string]any{
		"error": map[string]any{
			"code":    apperrors.CodeConflict,
			"message": "このメールアドレスは既に登録されています",
		},
	})
}

func TestSignUp_invalidEmail_fromUsecase(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(&fakeAuthUC{
		signUp: func(_ context.Context, _ usecase.SignUpInput) (*usecase.SignUpOutput, error) {
			return nil, domain.ErrInvalidEmail
		},
	}, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/signup", map[string]string{
		"name":     "太郎",
		"email":    "taro@example.com",
		"password": "secret1",
	})

	h.SignUp(c)

	testutil.AssertResponseJSON(t, rec, http.StatusBadRequest, map[string]any{
		"error": map[string]any{
			"code":    apperrors.CodeInvalidInput,
			"message": "メールアドレスの形式が正しくありません",
		},
	})
}

func TestSignUp_validationError(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(&fakeAuthUC{}, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/signup", map[string]string{
		"name":  "太郎",
		"email": "taro@example.com",
	})

	h.SignUp(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_success(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(&fakeAuthUC{
		login: func(_ context.Context, input usecase.LoginInput) (*usecase.LoginOutput, error) {
			if input.Email != "taro@example.com" || input.Password != "secret1" {
				t.Fatalf("unexpected input: %+v", input)
			}
			return &usecase.LoginOutput{Token: "jwt-abc"}, nil
		},
	}, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/login", map[string]string{
		"email":    "taro@example.com",
		"password": "secret1",
	})

	h.Login(c)

	testutil.AssertResponseJSON(t, rec, http.StatusOK, map[string]any{
		"token": "jwt-abc",
	})
}

func TestLogin_wrongCredentials(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(&fakeAuthUC{
		login: func(_ context.Context, _ usecase.LoginInput) (*usecase.LoginOutput, error) {
			return nil, domain.ErrUserNotFound
		},
	}, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/login", map[string]string{
		"email":    "taro@example.com",
		"password": "wrong",
	})

	h.Login(c)

	testutil.AssertResponseJSON(t, rec, http.StatusUnauthorized, map[string]any{
		"error": map[string]any{
			"code":    apperrors.CodeUnauthorized,
			"message": "メールアドレスまたはパスワードが正しくありません",
		},
	})
}

func TestLogin_internalError(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(&fakeAuthUC{
		login: func(_ context.Context, _ usecase.LoginInput) (*usecase.LoginOutput, error) {
			return nil, errors.New("db down")
		},
	}, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = jsonPOST("/login", map[string]string{
		"email":    "taro@example.com",
		"password": "secret1",
	})

	h.Login(c)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCurrentUser_authenticated(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(&fakeAuthUC{}, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req = req.WithContext(domain.WithUserID(req.Context(), domain.UserID(42)))
	c.Request = req

	h.CurrentUser(c)

	testutil.AssertResponseJSON(t, rec, http.StatusOK, map[string]any{
		"user_id": float64(42),
		"message": "認証に成功しています！",
	})
}

func TestCurrentUser_unauthorized(t *testing.T) {
	t.Parallel()

	h := newTestHandlers(&fakeAuthUC{}, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/me", nil)

	h.CurrentUser(c)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}
