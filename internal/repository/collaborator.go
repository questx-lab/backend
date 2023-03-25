package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"gorm.io/gorm"
)

type CollaboratorRepository interface {
	Create(ctx context.Context, e *entity.Collaborator) error
	GetList(ctx context.Context, offset, limit int) ([]*entity.Collaborator, error)
	DeleteByID(ctx context.Context, id string) error
	GetCollaborator(ctx context.Context, projectID, userID string) (*entity.Collaborator, error)
}

type collaboratorRepository struct {
	db *gorm.DB
}

func NewCollaboratorRepository(db *gorm.DB) CollaboratorRepository {
	return &collaboratorRepository{
		db: db,
	}
}

func (r *collaboratorRepository) Create(ctx context.Context, e *entity.Collaborator) error {
	if err := r.db.Model(&entity.Collaborator{}).Create(e).Error; err != nil {
		return err
	}
	return nil
}

func (r *collaboratorRepository) GetList(ctx context.Context, offset int, limit int) ([]*entity.Collaborator, error) {
	var result []*entity.Collaborator
	if err := r.db.Model(&entity.Collaborator{}).Limit(limit).Offset(offset).Find(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *collaboratorRepository) DeleteByID(ctx context.Context, id string) error {
	tx := r.db.
		Delete(&entity.Collaborator{}, "id = ?", id)
	if err := tx.Error; err != nil {
		return err
	}

	return nil
}

func (r *collaboratorRepository) GetCollaborator(ctx context.Context, projectID, userID string) (*entity.Collaborator, error) {
	var result entity.Collaborator
	if err := r.db.
		Model(&entity.Collaborator{}).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		First(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}
