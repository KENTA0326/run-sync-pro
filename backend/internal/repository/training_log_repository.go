package repository

import (
	"context"
	"errors"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"gorm.io/gorm"
)

type trainingLogRow struct {
	ID           uint           `gorm:"primaryKey"`
	UserID       uint           `gorm:"not null;index"`
	TrainingDate time.Time      `gorm:"column:training_date"`
	Distance     float64        `gorm:"column:distance"`
	Duration     int            `gorm:"column:duration"`
	Pace         string         `gorm:"column:pace"`
	Memo         string         `gorm:"column:memo"`
	Kind         int            `gorm:"column:kind"`
	ShoeID       uint           `gorm:"not null;index;column:shoe_id"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (trainingLogRow) TableName() string { return "training_logs" }

type trainingLogShoeRow struct {
	ID           uint      `gorm:"primaryKey"`
	Brand        string    `gorm:"column:brand"`
	Model        string    `gorm:"column:model"`
	PurchaseDate time.Time `gorm:"column:purchase_date"`
}

func (trainingLogShoeRow) TableName() string { return "shoes" }

// TrainingLogRepository は domain.TrainingLogRepository の GORM 実装。
type TrainingLogRepository struct {
	db *gorm.DB
}

func NewTrainingLogRepository(db *gorm.DB) *TrainingLogRepository {
	return &TrainingLogRepository{db: db}
}

func (r *TrainingLogRepository) Create(ctx context.Context, log *domain.TrainingLog) error {
	row := r.toRow(log)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	log.ID = row.ID
	log.CreatedAt = row.CreatedAt
	log.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *TrainingLogRepository) FindByID(ctx context.Context, id, userID uint) (*domain.TrainingLog, error) {
	var row trainingLogRow
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTrainingLogNotFound
		}
		return nil, err
	}
	log := r.toDomain(&row)
	var shoe trainingLogShoeRow
	if err := r.db.WithContext(ctx).Where("id = ?", row.ShoeID).First(&shoe).Error; err == nil {
		log.Shoe = &domain.Shoe{
			ID:           shoe.ID,
			Brand:        shoe.Brand,
			Model:        shoe.Model,
			PurchaseDate: shoe.PurchaseDate,
		}
	}
	return log, nil
}

func (r *TrainingLogRepository) List(ctx context.Context, userID uint, limit, offset int) ([]*domain.TrainingLog, int64, error) {
	base := r.db.WithContext(ctx).Model(&trainingLogRow{}).Where("user_id = ?", userID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []trainingLogRow
	if err := base.Order("training_date DESC, id DESC").
		Limit(limit).Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	shoeIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		shoeIDs = append(shoeIDs, row.ShoeID)
	}
	shoeMap := r.loadShoes(ctx, shoeIDs)

	logs := make([]*domain.TrainingLog, len(rows))
	for i := range rows {
		logs[i] = r.toDomain(&rows[i])
		if s, ok := shoeMap[rows[i].ShoeID]; ok {
			logs[i].Shoe = s
		}
	}
	return logs, total, nil
}

func (r *TrainingLogRepository) ListAll(ctx context.Context, userID uint) ([]*domain.TrainingLog, error) {
	var rows []trainingLogRow
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("training_date ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	logs := make([]*domain.TrainingLog, len(rows))
	for i := range rows {
		logs[i] = r.toDomain(&rows[i])
	}
	return logs, nil
}

func (r *TrainingLogRepository) loadShoes(ctx context.Context, ids []uint) map[uint]*domain.Shoe {
	if len(ids) == 0 {
		return nil
	}
	var shoes []trainingLogShoeRow
	r.db.WithContext(ctx).Where("id IN ?", ids).Find(&shoes)
	m := make(map[uint]*domain.Shoe, len(shoes))
	for _, s := range shoes {
		m[s.ID] = &domain.Shoe{
			ID:           s.ID,
			Brand:        s.Brand,
			Model:        s.Model,
			PurchaseDate: s.PurchaseDate,
		}
	}
	return m
}

func (r *TrainingLogRepository) toDomain(row *trainingLogRow) *domain.TrainingLog {
	return &domain.TrainingLog{
		ID:           row.ID,
		UserID:       row.UserID,
		TrainingDate: row.TrainingDate,
		Distance:     row.Distance,
		Duration:     row.Duration,
		Pace:         row.Pace,
		Memo:         row.Memo,
		Kind:         row.Kind,
		ShoeID:       row.ShoeID,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func (r *TrainingLogRepository) toRow(log *domain.TrainingLog) *trainingLogRow {
	return &trainingLogRow{
		ID:           log.ID,
		UserID:       log.UserID,
		TrainingDate: log.TrainingDate,
		Distance:     log.Distance,
		Duration:     log.Duration,
		Pace:         log.Pace,
		Memo:         log.Memo,
		Kind:         log.Kind,
		ShoeID:       log.ShoeID,
	}
}
