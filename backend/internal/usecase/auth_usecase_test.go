package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
)

type fakeUserRepo struct {
	existsByEmail func(ctx context.Context, email domain.Email) (bool, error)
	findByEmail   func(ctx context.Context, email domain.Email) (*domain.User, error)
	create        func(ctx context.Context, user *domain.User) error
}

func (f *fakeUserRepo) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	if f.findByEmail != nil {
		return f.findByEmail(ctx, email)
	}
	return nil, domain.ErrUserNotFound
}

func (f *fakeUserRepo) ExistsByEmail(ctx context.Context, email domain.Email) (bool, error) {
	if f.existsByEmail != nil {
		return f.existsByEmail(ctx, email)
	}
	return false, nil
}

func (f *fakeUserRepo) Create(ctx context.Context, user *domain.User) error {
	if f.create != nil {
		return f.create(ctx, user)
	}
	user.ID = 99
	return nil
}

func (f *fakeUserRepo) UpdatePassword(ctx context.Context, userID uint, hashedPassword string) error {
	return nil
}

type fakeAuthToken struct {
	hashPassword    func(password string) (string, error)
	checkPassword   func(password, hash string) bool
	generateToken   func(userID uint) (string, error)
}

func (f *fakeAuthToken) HashPassword(password string) (string, error) {
	if f.hashPassword != nil {
		return f.hashPassword(password)
	}
	return "hashed:" + password, nil
}

func (f *fakeAuthToken) CheckPassword(password, hash string) bool {
	if f.checkPassword != nil {
		return f.checkPassword(password, hash)
	}
	return hash == "hashed:"+password
}

func (f *fakeAuthToken) GenerateToken(userID uint) (string, error) {
	if f.generateToken != nil {
		return f.generateToken(userID)
	}
	return "token-for-" + string(rune(userID)), nil
}

func TestAuthUseCase_SignUp_success(t *testing.T) {
	t.Parallel()

	var created *domain.User
	uc := NewAuthUseCase(&fakeUserRepo{
		create: func(_ context.Context, user *domain.User) error {
			created = user
			user.ID = 42
			return nil
		},
	}, &fakeAuthToken{})

	out, err := uc.SignUp(context.Background(), SignUpInput{
		Name:     "花子",
		Email:    "hanako@example.com",
		Password: "pass123",
	})
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}
	if out.UserID != 42 {
		t.Fatalf("UserID=%d want 42", out.UserID)
	}
	if created == nil || created.Name != "花子" || created.Password != "hashed:pass123" {
		t.Fatalf("unexpected created user: %+v", created)
	}
	if created.Email.String() != "hanako@example.com" {
		t.Fatalf("email=%q", created.Email.String())
	}
}

func TestAuthUseCase_SignUp_duplicate(t *testing.T) {
	t.Parallel()

	uc := NewAuthUseCase(&fakeUserRepo{
		existsByEmail: func(_ context.Context, _ domain.Email) (bool, error) {
			return true, nil
		},
	}, &fakeAuthToken{})

	_, err := uc.SignUp(context.Background(), SignUpInput{
		Name:     "花子",
		Email:    "dup@example.com",
		Password: "pass123",
	})
	if !errors.Is(err, domain.ErrUserDuplicate) {
		t.Fatalf("err=%v want ErrUserDuplicate", err)
	}
}

func TestAuthUseCase_SignUp_invalidEmail(t *testing.T) {
	t.Parallel()

	uc := NewAuthUseCase(&fakeUserRepo{}, &fakeAuthToken{})

	_, err := uc.SignUp(context.Background(), SignUpInput{
		Name:     "花子",
		Email:    "not-an-email",
		Password: "pass123",
	})
	if !errors.Is(err, domain.ErrInvalidEmail) {
		t.Fatalf("err=%v want ErrInvalidEmail", err)
	}
}

func TestAuthUseCase_Login_success(t *testing.T) {
	t.Parallel()

	email, _ := domain.NewEmail("taro@example.com")
	uc := NewAuthUseCase(&fakeUserRepo{
		findByEmail: func(_ context.Context, _ domain.Email) (*domain.User, error) {
			return &domain.User{ID: 7, Email: email, Password: "hashed:secret"}, nil
		},
	}, &fakeAuthToken{
		generateToken: func(userID uint) (string, error) {
			if userID != 7 {
				t.Fatalf("userID=%d", userID)
			}
			return "jwt-xyz", nil
		},
	})

	out, err := uc.Login(context.Background(), LoginInput{
		Email:    "taro@example.com",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if out.Token != "jwt-xyz" {
		t.Fatalf("token=%q", out.Token)
	}
}

func TestAuthUseCase_Login_wrongPassword(t *testing.T) {
	t.Parallel()

	email, _ := domain.NewEmail("taro@example.com")
	uc := NewAuthUseCase(&fakeUserRepo{
		findByEmail: func(_ context.Context, _ domain.Email) (*domain.User, error) {
			return &domain.User{ID: 7, Email: email, Password: "hashed:secret"}, nil
		},
	}, &fakeAuthToken{
		checkPassword: func(_, _ string) bool { return false },
	})

	_, err := uc.Login(context.Background(), LoginInput{
		Email:    "taro@example.com",
		Password: "wrong",
	})
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("err=%v want ErrUserNotFound", err)
	}
}
