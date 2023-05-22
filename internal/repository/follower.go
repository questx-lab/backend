package repository

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type FollowerRepository interface {
	Get(ctx context.Context, userID, communityID string) (*entity.Follower, error)
	GetList(ctx context.Context, communityID string) ([]entity.Follower, error)
	GetByReferralCode(ctx context.Context, code string) (*entity.Follower, error)
	Create(ctx context.Context, data *entity.Follower) error
	IncreaseInviteCount(ctx context.Context, userID, communityID string) error
	IncreaseStat(ctx context.Context, userID, communityID string, point, streak int) error
}

type followerRepository struct{}

func NewFollowerRepository() *followerRepository {
	return &followerRepository{}
}

func (r *followerRepository) Get(ctx context.Context, userID, communityID string) (*entity.Follower, error) {
	var result entity.Follower
	err := xcontext.DB(ctx).Where("user_id=? AND community_id=?", userID, communityID).Take(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *followerRepository) GetList(ctx context.Context, communityID string) ([]entity.Follower, error) {
	var result []entity.Follower
	err := xcontext.DB(ctx).Where("community_id=?", communityID).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *followerRepository) Create(ctx context.Context, data *entity.Follower) error {
	return xcontext.DB(ctx).Create(data).Error
}

func (r *followerRepository) IncreaseInviteCount(ctx context.Context, userID, communityID string) error {
	tx := xcontext.DB(ctx).
		Model(&entity.Follower{}).
		Where("user_id=? AND community_id=?", userID, communityID).
		Update("invite_count", gorm.Expr("invite_count+1"))

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected > 1 {
		return errors.New("the number of affected rows is invalid")
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *followerRepository) IncreaseStat(
	ctx context.Context, userID, communityID string, points, streak int,
) error {
	updateMap := map[string]any{
		"points": gorm.Expr("points+?", points),
		"streak": gorm.Expr("streak+?", streak),
	}

	// Reset the streak if parameter is -1.
	if streak == -1 {
		updateMap["streak"] = 1
	}

	tx := xcontext.DB(ctx).
		Model(&entity.Follower{}).
		Where("user_id=? AND community_id=?", userID, communityID).
		Updates(updateMap)

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected > 1 {
		return errors.New("the number of rows effected is invalid")
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *followerRepository) GetByReferralCode(
	ctx context.Context, code string,
) (*entity.Follower, error) {
	var result entity.Follower
	if err := xcontext.DB(ctx).Take(&result, "invite_code=?", code).Error; err != nil {
		return nil, err
	}

	if err := xcontext.DB(ctx).Take(&result.Community, "id=?", result.CommunityID).Error; err != nil {
		return nil, err
	}

	if err := xcontext.DB(ctx).Take(&result.User, "id=?", result.UserID).Error; err != nil {
		return nil, err
	}

	return &result, nil
}
