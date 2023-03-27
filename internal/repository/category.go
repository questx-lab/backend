package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(ctx context.Context, e *entity.Category) error
	GetList(ctx context.Context) ([]*entity.Category, error)
	GetByID(ctx context.Context, id string) (*entity.Category, error)
	DeleteByID(ctx context.Context, id string) error
	UpdateByID(ctx context.Context, id string, data *entity.Category) error
	IsExisted(ctx context.Context, projectID string, ids ...string) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, e *entity.Category) error {
	if err := r.db.Model(&entity.Category{}).Create(e).Error; err != nil {
		return err
	}
	return nil
}

func (r *categoryRepository) GetList(ctx context.Context) ([]*entity.Category, error) {
	var result []*entity.Category
	if err := r.db.Model(&entity.Collaborator{}).Find(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *categoryRepository) DeleteByID(ctx context.Context, id string) error {
	tx := r.db.
		Delete(&entity.Category{}, "id = ?", id)
	if err := tx.Error; err != nil {
		return err
	}

	return nil
}

func (r *categoryRepository) UpdateByID(ctx context.Context, id string, data *entity.Category) error {
	tx := r.db.
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

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*entity.Category, error) {
	var result entity.Category
	if err := r.db.Model(&entity.Category{}).Where("id = ?", id).First(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *categoryRepository) IsExisted(ctx context.Context, projectID string, ids ...string) error {
	var count int64
	err := r.db.Model(&entity.Category{}).
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
