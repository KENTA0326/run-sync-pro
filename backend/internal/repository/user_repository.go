package repository

import (
	"context"
	"errors"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"gorm.io/gorm"
)

// userRow は GORM 永続化用の構造体。ドメインエンティティとは分離する。
type userRow struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"column:name"`
	Email     string         `gorm:"uniqueIndex;not null;column:email;type:text"`
	Password  string         `gorm:"column:password"`
	Role      int            `gorm:"default:1;column:role"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (userRow) TableName() string { return "users" }

// UserRepository は domain.UserRepository の GORM 実装。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository は UserRepository を生成する。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	var row userRow
	err := r.db.WithContext(ctx).Where("email = ?", email.String()).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return r.toDomain(&row), nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email domain.Email) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&userRow{}).Where("email = ?", email.String()).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := r.toRow(user)
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		user.ID = row.ID

		var roleID uint
		if err := tx.Table("roles").Select("id").Where("name = ?", "viewer").Scan(&roleID).Error; err != nil {
			return err
		}
		if roleID == 0 {
			return errors.New("viewer role not found")
		}
		return tx.Table("user_roles").Create(map[string]any{
			"user_id": user.ID,
			"role_id": roleID,
		}).Error
	})
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID uint, hashedPassword string) error {
	return r.db.WithContext(ctx).Model(&userRow{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error
}

func (r *UserRepository) toDomain(row *userRow) *domain.User {
	email, _ := domain.NewEmail(row.Email)
	return &domain.User{
		ID:        row.ID,
		Name:      row.Name,
		Email:     email,
		Password:  row.Password,
		Role:      row.Role,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func (r *UserRepository) toRow(u *domain.User) *userRow {
	return &userRow{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email.String(),
		Password: u.Password,
		Role:     u.Role,
	}
}
