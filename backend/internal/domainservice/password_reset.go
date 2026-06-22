package domainservice

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const passwordResetTokenBytes = 32

// ErrPasswordResetInvalid はトークン無効・期限切れ・使用済み。
var ErrPasswordResetInvalid = errors.New("password reset token invalid")

// PasswordResetTokenRow は DB 上のリセットトークン行。
type PasswordResetTokenRow struct {
	ID        uint       `gorm:"primaryKey"`
	UserID    uint       `gorm:"column:user_id"`
	TokenHash string     `gorm:"column:token_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (PasswordResetTokenRow) TableName() string { return "password_reset_tokens" }

// IssuePasswordResetToken は平文トークンを発行し DB に保存する。平文トークンを返す（メール・URL 用）。
func IssuePasswordResetToken(db *gorm.DB, userID uint, ttl time.Duration) (plainToken string, err error) {
	raw := make([]byte, passwordResetTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("issue reset token: %w", err)
	}
	plainToken = hex.EncodeToString(raw)
	hash := hashResetToken(plainToken)

	expiresAt := time.Now().UTC().Add(ttl)
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND used_at IS NULL", userID).
			Delete(&PasswordResetTokenRow{}).Error; err != nil {
			return err
		}
		return tx.Create(&PasswordResetTokenRow{
			UserID:    userID,
			TokenHash: hash,
			ExpiresAt: expiresAt,
		}).Error
	})
	if err != nil {
		return "", fmt.Errorf("issue reset token db: %w", err)
	}
	return plainToken, nil
}

// ConsumePasswordResetToken はトークンを検証し、使用済みにする（パスワード更新は呼び出し側）。
func ConsumePasswordResetToken(db *gorm.DB, plainToken string) (userID uint, err error) {
	hash := hashResetToken(plainToken)
	var row PasswordResetTokenRow
	q := db.Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hash, time.Now().UTC()).
		First(&row)
	if q.Error != nil {
		if errors.Is(q.Error, gorm.ErrRecordNotFound) {
			return 0, ErrPasswordResetInvalid
		}
		return 0, q.Error
	}

	now := time.Now().UTC()
	res := db.Model(&row).Update("used_at", now)
	if res.Error != nil {
		return 0, res.Error
	}
	return row.UserID, nil
}

func hashResetToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// BuildPasswordResetURL はフロントのリセット画面 URL を組み立てる。
func BuildPasswordResetURL(frontendBase, plainToken string) string {
	return fmt.Sprintf("%s/reset-password?token=%s", trimTrailingSlash(frontendBase), plainToken)
}

func trimTrailingSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
