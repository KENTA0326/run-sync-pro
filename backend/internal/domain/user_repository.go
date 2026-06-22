package domain

import "context"

// UserRepository はユーザーの永続化操作を定義するインターフェース。
// 実装は infrastructure 層（repository パッケージ）に置く。
type UserRepository interface {
	// FindByEmail はメールアドレスでユーザーを検索する。見つからない場合は ErrUserNotFound を返す。
	FindByEmail(ctx context.Context, email Email) (*User, error)
	// ExistsByEmail はメールアドレスの重複チェック。
	ExistsByEmail(ctx context.Context, email Email) (bool, error)
	// Create はユーザーを作成し、デフォルトロールを紐付ける。
	Create(ctx context.Context, user *User) error
	// UpdatePassword はパスワードを更新する。
	UpdatePassword(ctx context.Context, userID uint, hashedPassword string) error
}
