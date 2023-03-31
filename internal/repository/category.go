package repository

import (
	"errors"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type CategoryRepository interface {
	Create(ctx xcontext.Context, e *entity.Category) error
	GetList(ctx xcontext.Context) ([]*entity.Category, error)
	GetByID(ctx xcontext.Context, id string) (*entity.Category, error)
	DeleteByID(ctx xcontext.Context, id string) error
	UpdateByID(ctx xcontext.Context, id string, data *entity.Category) error
	IsExisted(ctx xcontext.Context, projectID string, ids ...string) error
}

type categoryRepository struct{}

func NewCategoryRepository() CategoryRepository {
	return &categoryRepository{}
}

func (r *categoryRepository) Create(ctx xcontext.Context, e *entity.Category) error {
	if err := ctx.DB().Model(&entity.Category{}).Create(e).Error; err != nil {
		return err
	}
	return nil
}

func (r *categoryRepository) GetList(ctx xcontext.Context) ([]*entity.Category, error) {
	var result []*entity.Category
	if err := ctx.DB().Model(&entity.Collaborator{}).Find(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *categoryRepository) DeleteByID(ctx xcontext.Context, id string) error {
	tx := ctx.DB().
		Delete(&entity.Category{}, "id = ?", id)
	if err := tx.Error; err != nil {
		return err
	}

	return nil
}

func (r *categoryRepository) UpdateByID(ctx xcontext.Context, id string, data *entity.Category) error {
	tx := ctx.DB().
		Model(&entity.Category{}).
		Where("id = ?", id).
		Update("name = ?", data.Name)
	if err := tx.Error; err != nil {
		return err
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	return nil
}

func (r *categoryRepository) GetByID(ctx xcontext.Context, id string) (*entity.Category, error) {
	var result entity.Category
	if err := ctx.DB().Model(&entity.Category{}).Where("id = ?", id).First(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *categoryRepository) IsExisted(ctx xcontext.Context, projectID string, ids ...string) error {
	var count int64
	err := ctx.DB().Model(&entity.Category{}).
		Where("project_id = ? AND id IN (?)", projectID, ids).
		Count(&count).Error
	if err != nil {
		return err
	}

	if int(count) != len(ids) {
		return errors.New("some categories not found")
	}

	return nil
}
