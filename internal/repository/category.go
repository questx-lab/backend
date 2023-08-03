package repository

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(ctx context.Context, e *entity.Category) error
	GetList(ctx context.Context, communityID string) ([]entity.Category, error)
	GetTemplates(ctx context.Context) ([]entity.Category, error)
	GetByID(ctx context.Context, id string) (*entity.Category, error)
	GetByName(ctx context.Context, communityID, name string) (*entity.Category, error)
	DeleteByID(ctx context.Context, id string) error
	UpdateByID(ctx context.Context, id string, data *entity.Category) error
	GetLastPosition(ctx context.Context, communityID string) (int, error)
	IncreasePosition(ctx context.Context, communityID string, from, to int) error
	DecreasePosition(ctx context.Context, communityID string, from, to int) error
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
	tx := xcontext.DB(ctx).Model(&entity.Category{})
	if communityID != "" {
		tx.Where("community_id=?", communityID)
	} else {
		tx.Where("community_id IS NOT NULL") // get all category except global ones.
	}

	if err := tx.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *categoryRepository) GetTemplates(ctx context.Context) ([]entity.Category, error) {
	var result []entity.Category
	if err := xcontext.DB(ctx).Find(&result, "community_id IS NULL").Error; err != nil {
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

func (r *categoryRepository) GetByName(ctx context.Context, communityID, name string) (*entity.Category, error) {
	var result entity.Category
	tx := xcontext.DB(ctx).Where("name=?", name)
	if communityID != "" {
		tx.Where("community_id=?", communityID)
	} else {
		tx.Where("community_id IS NULL")
	}
	if err := tx.Take(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *categoryRepository) GetLastPosition(ctx context.Context, communityID string) (int, error) {
	var result int
	err := xcontext.DB(ctx).Model(&entity.Category{}).Select("position").
		Where("community_id = ?", communityID).
		Order("position DESC").
		Take(&result).Error

	if err != nil {
		return 0, err
	}

	return result, err
}

func (r *categoryRepository) IncreasePosition(
	ctx context.Context, communityID string, from, to int,
) error {
	tx := xcontext.DB(ctx).Model(&entity.Quest{})

	if from != -1 {
		tx.Where("position >= ?", from)
	}

	if to != -1 {
		tx.Where("position <= ?", to)
	}

	if communityID == "" {
		tx.Where("community_id IS NULL")
	} else {
		tx.Where("community_id=?", communityID)
	}

	if err := tx.Update("position", gorm.Expr("position+?", 1)).Error; err != nil {
		return err
	}

	return nil
}

func (r *categoryRepository) DecreasePosition(
	ctx context.Context, communityID string, from, to int,
) error {
	tx := xcontext.DB(ctx).Model(&entity.Quest{}).
		Where("community_id=?", communityID)

	if from != -1 {
		tx.Where("position >= ?", from)
	}

	if to != -1 {
		tx.Where("position <= ?", to)
	}

	if err := tx.Update("position", gorm.Expr("position-?", 1)).Error; err != nil {
		return err
	}

	return nil
}
