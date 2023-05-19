package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm/clause"
)

type BadgeRepo interface {
	Upsert(ctx context.Context, badge *entity.Badge) error
	Get(ctx context.Context, userID, projectID, badgeName string) (*entity.Badge, error)
	GetAll(ctx context.Context, userID, projectID string) ([]entity.Badge, error)
	UpdateNotification(ctx context.Context, userID, projectID string) error
}

type badgeRepo struct{}

func NewBadgeRepository() *badgeRepo {
	return &badgeRepo{}
}

func (r *badgeRepo) Upsert(ctx context.Context, badge *entity.Badge) error {
	return xcontext.DB(ctx).Model(&entity.Badge{}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "project_id"},
				{Name: "user_id"},
				{Name: "name"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"level":        badge.Level,
				"was_notified": badge.WasNotified,
			}),
		}).
		Create(badge).Error
}

func (r *badgeRepo) Get(ctx context.Context, userID, projectID, badgeName string) (*entity.Badge, error) {
	result := &entity.Badge{}
	err := xcontext.DB(ctx).
		Where("user_id=? AND project_id=? AND name=?", userID, projectID, badgeName).
		Take(result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *badgeRepo) GetAll(ctx context.Context, userID, projectID string) ([]entity.Badge, error) {
	result := []entity.Badge{}
	tx := xcontext.DB(ctx).Where("user_id=?", userID)
	if projectID != "" {
		tx = tx.Where("project_id=?", projectID)
	}

	if err := tx.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *badgeRepo) UpdateNotification(ctx context.Context, userID, projectID string) error {
	tx := xcontext.DB(ctx).Model(&entity.Badge{}).Where("user_id=?", userID)
	if projectID != "" {
		tx = tx.Where("project_id=?", projectID)
	}

	return tx.Update("was_notified", true).Error
}
