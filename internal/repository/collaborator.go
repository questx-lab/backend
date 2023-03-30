package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"gorm.io/gorm"
)

type CollaboratorRepository interface {
	Create(ctx context.Context, e *entity.Collaborator) error
	GetList(ctx context.Context, offset, limit int) ([]*entity.Collaborator, error)
	Delete(ctx context.Context, projectID, userID string) error
	Get(ctx context.Context, projectID, userID string) (*entity.Collaborator, error)
	UpdateRole(ctx context.Context, userID, projectID string, role entity.Role) error
}

type collaboratorRepository struct {
	db *gorm.DB
}

func NewCollaboratorRepository(db *gorm.DB) CollaboratorRepository {
	return &collaboratorRepository{db: db}
}

func (r *collaboratorRepository) Create(ctx context.Context, e *entity.Collaborator) error {
	if err := r.db.Create(e).Error; err != nil {
		return err
	}
	return nil
}

func (r *collaboratorRepository) GetList(ctx context.Context, offset int, limit int) ([]*entity.Collaborator, error) {
	var result []*entity.Collaborator
	if err := r.db.Limit(limit).Offset(offset).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *collaboratorRepository) Delete(ctx context.Context, projectID, userID string) error {
	tx := r.db.
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Delete(&entity.Collaborator{})
	if err := tx.Error; err != nil {
		return err
	}

	return nil
}

func (r *collaboratorRepository) Get(ctx context.Context, projectID, userID string) (*entity.Collaborator, error) {
	var result entity.Collaborator
	err := r.db.
		Where("user_id=? AND project_id=?", userID, projectID).
		First(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *collaboratorRepository) UpdateRole(ctx context.Context, userID, projectID string, role entity.Role) error {
	if err := r.db.
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Update("role", role).Error; err != nil {
		return err
	}
	return nil
}
