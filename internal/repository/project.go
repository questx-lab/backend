package repository

import (
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type ProjectRepository interface {
	Create(ctx xcontext.Context, e *entity.Project) error
	GetList(ctx xcontext.Context, offset, limit int) ([]*entity.Project, error)
	GetByID(ctx xcontext.Context, id string) (*entity.Project, error)
	UpdateByID(ctx xcontext.Context, id string, e *entity.Project) error
	DeleteByID(ctx xcontext.Context, id string) error
	GetListByUserID(ctx xcontext.Context, userID string, offset, limit int) ([]*entity.Project, error)
}

type projectRepository struct{}

func NewProjectRepository() ProjectRepository {
	return &projectRepository{}
}

func (r *projectRepository) Create(ctx xcontext.Context, e *entity.Project) error {
	if err := ctx.DB().Model(e).Create(e).Error; err != nil {
		return err
	}

	return nil
}

func (r *projectRepository) GetList(ctx xcontext.Context, offset int, limit int) ([]*entity.Project, error) {
	var result []*entity.Project
	if err := ctx.DB().Model(&entity.Project{}).Limit(limit).Offset(offset).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) GetByID(ctx xcontext.Context, id string) (*entity.Project, error) {
	result := &entity.Project{}
	if err := ctx.DB().Model(&entity.Project{}).Take(result, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) UpdateByID(ctx xcontext.Context, id string, e *entity.Project) error {
	tx := ctx.DB().
		Model(&entity.Project{}).
		Where("id = ?", id).
		Omit("created_by", "created_at", "id").
		Updates(*e)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	return nil
}

func (r *projectRepository) DeleteByID(ctx xcontext.Context, id string) error {
	tx := ctx.DB().
		Delete(&entity.Project{}, "id = ?", id)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	return nil
}

func (r *projectRepository) GetListByUserID(ctx xcontext.Context, userID string, offset, limit int) ([]*entity.Project, error) {
	var result []*entity.Project
	if err := ctx.
		DB().
		Model(&entity.Project{}).
		Joins("join collaborators on projects.id = collaborators.project_id").
		Where("collaborators.user_id = ?", userID).
		Limit(limit).Offset(offset).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}
