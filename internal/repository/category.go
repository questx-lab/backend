package repository

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type CategoryRepository interface {
	Create(ctx context.Context, e *entity.Category) error
	GetList(ctx context.Context, communityID string) ([]entity.Category, error)
	GetByID(ctx context.Context, id string) (*entity.Category, error)
	DeleteByID(ctx context.Context, id string) error
	UpdateByID(ctx context.Context, id string, data *entity.Category) error
}

type categoryRepository struct{}

func NewCategoryRepository() CategoryRepository {
	return &categoryRepository{}
}

func (r *categoryRepository) Create(ctx context.Context, e *entity.Category) error {
	if err := xcontext.DB(ctx).Create(e).Error; err != nil {
		return err
	}
	return nil
}

func (r *categoryRepository) GetList(ctx context.Context, communityID string) ([]entity.Category, error) {
	var result []entity.Category
	if err := xcontext.DB(ctx).Find(&result, "community_id=?", communityID).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *categoryRepository) DeleteByID(ctx context.Context, id string) error {
	tx := xcontext.DB(ctx).
		Delete(&entity.Category{}, "id=?", id)
	if err := tx.Error; err != nil {
		return err
	}

	return nil
}

func (r *categoryRepository) UpdateByID(ctx context.Context, id string, data *entity.Category) error {
	tx := xcontext.DB(ctx).
		Model(&entity.Category{}).
		Where("id=?", id).
		Updates(data)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	return nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*entity.Category, error) {
	var result entity.Category
	if err := xcontext.DB(ctx).Where("id=?", id).Take(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}
