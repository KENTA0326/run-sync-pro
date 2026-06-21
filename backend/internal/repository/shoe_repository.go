package repository

import (
	"context"
	"errors"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"gorm.io/gorm"
)

type shoeRow struct {
	ID            uint           `gorm:"primaryKey"`
	UserID        uint           `gorm:"not null;index"`
	Brand         string         `gorm:"column:brand"`
	Model         string         `gorm:"column:model"`
	PurchaseDate  time.Time      `gorm:"column:purchase_date"`
	TotalDistance float64        `gorm:"default:0;column:total_distance"`
	IsActive      bool           `gorm:"default:true;column:is_active"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (shoeRow) TableName() string { return "shoes" }

// ShoeRepository は domain.ShoeRepository の GORM 実装。
type ShoeRepository struct {
	db *gorm.DB
}

func NewShoeRepository(db *gorm.DB) *ShoeRepository {
	return &ShoeRepository{db: db}
}

func (r *ShoeRepository) Create(ctx context.Context, shoe *domain.Shoe) error {
	row := r.toRow(shoe)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	shoe.ID = row.ID
	shoe.CreatedAt = row.CreatedAt
	shoe.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *ShoeRepository) FindByID(ctx context.Context, id, userID uint) (*domain.Shoe, error) {
	var row shoeRow
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND is_active = ?", id, userID, true).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrShoeNotFound
		}
		return nil, err
	}
	return r.toDomain(&row), nil
}

func (r *ShoeRepository) ListActive(ctx context.Context, userID uint, limit, offset int) ([]*domain.Shoe, int64, error) {
	base := r.db.WithContext(ctx).Model(&shoeRow{}).
		Where("user_id = ? AND is_active = ?", userID, true)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []shoeRow
	if err := base.Order("purchase_date DESC, id DESC").
		Limit(limit).Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	shoes := make([]*domain.Shoe, len(rows))
	for i := range rows {
		shoes[i] = r.toDomain(&rows[i])
	}
	return shoes, total, nil
}

func (r *ShoeRepository) ListAllActive(ctx context.Context, userID uint) ([]*domain.Shoe, error) {
	var rows []shoeRow
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	shoes := make([]*domain.Shoe, len(rows))
	for i := range rows {
		shoes[i] = r.toDomain(&rows[i])
	}
	return shoes, nil
}

func (r *ShoeRepository) Deactivate(ctx context.Context, id, userID uint) error {
	res := r.db.WithContext(ctx).Model(&shoeRow{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_active", false)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrShoeNotFound
	}
	return nil
}

func (r *ShoeRepository) AddDistance(ctx context.Context, id uint, distance float64) error {
	return r.db.WithContext(ctx).Model(&shoeRow{}).
		Where("id = ?", id).
		Update("total_distance", gorm.Expr("total_distance + ?", distance)).Error
}

func (r *ShoeRepository) toDomain(row *shoeRow) *domain.Shoe {
	return &domain.Shoe{
		ID:            row.ID,
		UserID:        row.UserID,
		Brand:         row.Brand,
		Model:         row.Model,
		PurchaseDate:  row.PurchaseDate,
		TotalDistance: row.TotalDistance,
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func (r *ShoeRepository) toRow(s *domain.Shoe) *shoeRow {
	return &shoeRow{
		ID:            s.ID,
		UserID:        s.UserID,
		Brand:         s.Brand,
		Model:         s.Model,
		PurchaseDate:  s.PurchaseDate,
		TotalDistance: s.TotalDistance,
		IsActive:      s.IsActive,
	}
}
