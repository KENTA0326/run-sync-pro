package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTの秘密鍵（本来は環境変数から読み込むべき項目）
var jwtKey = []byte("your_secret_key")

type authStd struct{}

// NewAuth は本番用の認証具体実装を返す。
func NewAuth() *authStd {
	return &authStd{}
}

func (authStd) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (authStd) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (authStd) GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		// Unix は瞬間なのでタイムゾーンとは無関係（Now はホストローカル時計ソースのみ）。
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
