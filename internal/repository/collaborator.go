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
	GetListByCommunityID(ctx context.Context, communityID string, offset, limit int) ([]entity.Collaborator, error)
	Delete(ctx context.Context, communityID, userID string) error
	Get(ctx context.Context, communityID, userID string) (*entity.Collaborator, error)
	DeleteOldOwnerByCommunityID(ctx context.Context, communityID string) error
}

type collaboratorRepository struct{}

func NewCollaboratorRepository() CollaboratorRepository {
	return &collaboratorRepository{}
}

func (r *collaboratorRepository) Upsert(ctx context.Context, collab *entity.Collaborator) error {
	if err := xcontext.DB(ctx).Model(&entity.Collaborator{}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "community_id"},
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

func (r *collaboratorRepository) Delete(ctx context.Context, communityID, userID string) error {
	tx := xcontext.DB(ctx).
		Where("user_id=? AND community_id=?", userID, communityID).
		Delete(&entity.Collaborator{})
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	return nil
}

func (r *collaboratorRepository) Get(ctx context.Context, communityID, userID string) (*entity.Collaborator, error) {
	var result entity.Collaborator
	err := xcontext.DB(ctx).
		Where("user_id=? AND community_id=?", userID, communityID).
		First(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *collaboratorRepository) GetListByCommunityID(
	ctx context.Context, communityID string, offset, limit int,
) ([]entity.Collaborator, error) {
	var result []entity.Collaborator
	err := xcontext.DB(ctx).
		Where("community_id=?", communityID).
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
		if err := xcontext.DB(ctx).Take(&result[i].Community, "id=?", result[i].CommunityID).Error; err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (r *collaboratorRepository) DeleteOldOwnerByCommunityID(ctx context.Context, communityID string) error {
	tx := xcontext.DB(ctx).
		Where("community_id = ? AND role = ?", communityID, entity.Owner).
		Delete(&entity.Collaborator{})
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row effected is wrong")
	}

	return nil
}
