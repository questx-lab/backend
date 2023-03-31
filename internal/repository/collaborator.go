package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type CollaboratorRepository interface {
	Create(ctx xcontext.Context, e *entity.Collaborator) error
	GetList(ctx xcontext.Context, offset, limit int) ([]*entity.Collaborator, error)
	Delete(ctx xcontext.Context, projectID, userID string) error
	Get(ctx xcontext.Context, projectID, userID string) (*entity.Collaborator, error)
	UpdateRole(ctx xcontext.Context, userID, projectID string, role entity.Role) error
}

type collaboratorRepository struct{}

func NewCollaboratorRepository() CollaboratorRepository {
	return &collaboratorRepository{}
}

func (r *collaboratorRepository) Create(ctx xcontext.Context, e *entity.Collaborator) error {
	if err := ctx.DB().Create(e).Error; err != nil {
		return err
	}
	return nil
}

func (r *collaboratorRepository) GetList(ctx xcontext.Context, offset int, limit int) ([]*entity.Collaborator, error) {
	var result []*entity.Collaborator
	if err := ctx.DB().Limit(limit).Offset(offset).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *collaboratorRepository) Delete(ctx xcontext.Context, projectID, userID string) error {
	tx := ctx.DB().
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Delete(&entity.Collaborator{})
	if err := tx.Error; err != nil {
		return err
	}

	return nil
}

func (r *collaboratorRepository) Get(ctx xcontext.Context, projectID, userID string) (*entity.Collaborator, error) {
	var result entity.Collaborator
	err := ctx.DB().
		Where("user_id=? AND project_id=?", userID, projectID).
		First(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *collaboratorRepository) UpdateRole(ctx xcontext.Context, userID, projectID string, role entity.Role) error {
	if err := ctx.DB().
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Update("role", role).Error; err != nil {
		return err
	}
	return nil
}
