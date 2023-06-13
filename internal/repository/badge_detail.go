package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type BadgeDetailRepository interface {
	Create(ctx context.Context, badge *entity.BadgeDetail) error
	GetLatest(ctx context.Context, userID, communityID, badgeName string) (*entity.BadgeDetail, error)
	GetAll(ctx context.Context, userID, communityID string) ([]entity.BadgeDetail, error)
	UpdateNotification(ctx context.Context, userID, communityID string) error
}

type badgeDetailRepository struct{}

func NewBadgeDetailRepository() *badgeDetailRepository {
	return &badgeDetailRepository{}
}

func (r *badgeDetailRepository) Create(ctx context.Context, badgeDetail *entity.BadgeDetail) error {
	return xcontext.DB(ctx).Create(badgeDetail).Error
}

func (r *badgeDetailRepository) GetLatest(ctx context.Context, userID, communityID, badgeName string) (*entity.BadgeDetail, error) {
	result := &entity.BadgeDetail{}
	err := xcontext.DB(ctx).Model(&entity.BadgeDetail{}).
		Joins("join badges on badges.id=badge_details.badge_id").
		Where("badge_details.user_id=? AND badge_details.community_id=? AND badges.name=?",
			userID, communityID, badgeName).
		Order("badges.level DESC").
		Take(result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *badgeDetailRepository) GetAll(ctx context.Context, userID, communityID string) ([]entity.BadgeDetail, error) {
	result := []entity.BadgeDetail{}
	tx := xcontext.DB(ctx).Where("user_id=?", userID)
	if communityID != "" {
		tx = tx.Where("community_id=?", communityID)
	}

	if err := tx.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *badgeDetailRepository) UpdateNotification(ctx context.Context, userID, communityID string) error {
	tx := xcontext.DB(ctx).Model(&entity.BadgeDetail{}).Where("user_id=?", userID)
	if communityID != "" {
		tx.Where("community_id=?", communityID)
	} else {
		tx.Where("community_id is NULL")
	}

	return tx.Update("was_notified", true).Error
}
