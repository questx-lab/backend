package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm/clause"
)

type BadgeRepository interface {
	Create(ctx context.Context, badge *entity.Badge) error
	Get(ctx context.Context, name string, level int) (*entity.Badge, error)
	GetByID(ctx context.Context, id string) (*entity.Badge, error)
	GetLessThanValue(ctx context.Context, name string, value int) ([]entity.Badge, error)
	GetAll(ctx context.Context) ([]entity.Badge, error)
}

type badgeRepository struct{}

func NewBadgeRepository() *badgeRepository {
	return &badgeRepository{}
}

func (r *badgeRepository) Create(ctx context.Context, badge *entity.Badge) error {
	return xcontext.DB(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "name"},
				{Name: "level"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"value":       badge.Value,
				"description": badge.Description,
				"icon_url":    badge.IconURL,
			}),
		}).Create(badge).Error
}

func (r *badgeRepository) Get(ctx context.Context, name string, level int) (*entity.Badge, error) {
	result := &entity.Badge{}
	if err := xcontext.DB(ctx).Where("name=? AND level=?", name, level).Take(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *badgeRepository) GetByID(ctx context.Context, id string) (*entity.Badge, error) {
	result := &entity.Badge{}
	if err := xcontext.DB(ctx).Where("id?", id).Take(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *badgeRepository) GetAll(ctx context.Context) ([]entity.Badge, error) {
	result := []entity.Badge{}
	if err := xcontext.DB(ctx).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *badgeRepository) GetLessThanValue(ctx context.Context, name string, value int) ([]entity.Badge, error) {
	result := []entity.Badge{}
	err := xcontext.DB(ctx).
		Where("name=? AND value<=?", name, value).
		Order("level ASC").
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
