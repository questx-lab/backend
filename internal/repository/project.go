package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/router"
)

type ProjectRepository interface {
	Create(ctx router.Context, e *entity.Project) error
	GetList(ctx router.Context, offset, limit int) ([]*entity.Project, error)
	GetByID(ctx router.Context, id string) (*entity.Project, error)
	UpdateByID(ctx router.Context, id string, e *entity.Project) error
	DeleteByID(ctx router.Context, id string) error
}

type projectRepository struct{}

func NewProjectRepository() ProjectRepository {
	return &projectRepository{}
}

func (r *projectRepository) Create(ctx router.Context, e *entity.Project) error {
	if err := ctx.DB().Model(e).Create(e).Error; err != nil {
		return err
	}

	return nil
}

func (r *projectRepository) GetList(ctx router.Context, offset int, limit int) ([]*entity.Project, error) {
	var result []*entity.Project
	if err := ctx.DB().Model(&entity.Project{}).Limit(limit).Offset(offset).Find(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) GetByID(ctx router.Context, id string) (*entity.Project, error) {
	result := &entity.Project{}
	if err := ctx.DB().Model(&entity.Project{}).First(result, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) UpdateByID(ctx router.Context, id string, e *entity.Project) error {
	if err := ctx.DB().
		Model(&entity.Project{}).
		Where("id = ?", id).
		Omit("created_by", "created_at", "id").
		Updates(*e).Error; err != nil {
		return err
	}

	return nil
}

func (r *projectRepository) DeleteByID(ctx router.Context, id string) error {
	tx := ctx.DB().
		Delete(&entity.Project{}, "id = ?", id)
	if err := tx.Error; err != nil {
		return err
	}

	return nil
}
