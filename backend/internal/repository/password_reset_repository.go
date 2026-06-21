package repository

import (
	"context"
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

type passwordResetTokenRow struct {
	ID        uint       `gorm:"primaryKey"`
	UserID    uint       `gorm:"column:user_id"`
	TokenHash string     `gorm:"column:token_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (passwordResetTokenRow) TableName() string { return "password_reset_tokens" }

// PasswordResetRepository は domain.PasswordResetRepository の GORM 実装。
type PasswordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Issue(ctx context.Context, userID uint, ttl time.Duration) (string, error) {
	raw := make([]byte, passwordResetTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("issue reset token: %w", err)
	}
	plainToken := hex.EncodeToString(raw)
	hash := hashResetToken(plainToken)

	expiresAt := time.Now().UTC().Add(ttl)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND used_at IS NULL", userID).
			Delete(&passwordResetTokenRow{}).Error; err != nil {
			return err
		}
		return tx.Create(&passwordResetTokenRow{
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

func (r *PasswordResetRepository) Consume(ctx context.Context, plainToken string) (uint, error) {
	hash := hashResetToken(plainToken)
	var row passwordResetTokenRow
	q := r.db.WithContext(ctx).
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hash, time.Now().UTC()).
		First(&row)
	if q.Error != nil {
		if errors.Is(q.Error, gorm.ErrRecordNotFound) {
			return 0, ErrPasswordResetInvalid
		}
		return 0, q.Error
	}

	now := time.Now().UTC()
	if err := r.db.WithContext(ctx).Model(&row).Update("used_at", now).Error; err != nil {
		return 0, err
	}
	return row.UserID, nil
}

func hashResetToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
