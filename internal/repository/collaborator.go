package repository

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm/clause"
)

type CollaboratorRepository interface {
	Upsert(ctx context.Context, e *entity.Collaborator) error
	GetListByUserID(ctx context.Context, userID string, offset, limit int) ([]entity.Collaborator, error)
	GetListByProjectID(ctx context.Context, projectID string, offset, limit int) ([]entity.Collaborator, error)
	Delete(ctx context.Context, projectID, userID string) error
	Get(ctx context.Context, projectID, userID string) (*entity.Collaborator, error)
}

type collaboratorRepository struct{}

func NewCollaboratorRepository() CollaboratorRepository {
	return &collaboratorRepository{}
}

func (r *collaboratorRepository) Upsert(ctx context.Context, collab *entity.Collaborator) error {
	if err := xcontext.DB(ctx).Model(&entity.Collaborator{}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "project_id"},
				{Name: "user_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"role": collab.Role,
			}),
		}).Create(collab).Error; err != nil {
		return err
	}
	return nil
}

func (r *collaboratorRepository) Delete(ctx context.Context, projectID, userID string) error {
	tx := xcontext.DB(ctx).
		Where("user_id=? AND project_id=?", userID, projectID).
		Delete(&entity.Collaborator{})
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	return nil
}

func (r *collaboratorRepository) Get(ctx context.Context, projectID, userID string) (*entity.Collaborator, error) {
	var result entity.Collaborator
	err := xcontext.DB(ctx).
		Where("user_id=? AND project_id=?", userID, projectID).
		First(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *collaboratorRepository) GetListByProjectID(ctx context.Context, projectID string, offset, limit int) ([]entity.Collaborator, error) {
	var result []entity.Collaborator
	err := xcontext.DB(ctx).
		Where("project_id=?", projectID).
		Limit(limit).
		Offset(offset).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	for i := range result {
		if err := xcontext.DB(ctx).Take(&result[i].User, "id=?", result[i].UserID).Error; err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (r *collaboratorRepository) GetListByUserID(ctx context.Context, userID string, offset, limit int) ([]entity.Collaborator, error) {
	var result []entity.Collaborator
	err := xcontext.DB(ctx).
		Limit(limit).
		Offset(offset).
		Find(&result, "user_id=?", userID).Error
	if err != nil {
		return nil, err
	}

	for i := range result {
		if err := xcontext.DB(ctx).Take(&result[i].Project, "id=?", result[i].ProjectID).Error; err != nil {
			return nil, err
		}
	}

	return result, nil
}
