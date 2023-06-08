package repository

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FollowerRepository interface {
	Get(ctx context.Context, userID, communityID string) (*entity.Follower, error)
	GetListByCommunityID(ctx context.Context, communityID string) ([]entity.Follower, error)
	GetListByUserID(ctx context.Context, userID string) ([]entity.Follower, error)
	Create(ctx context.Context, data *entity.Follower) error
	IncreaseInviteCount(ctx context.Context, userID, communityID string) error
	IncreasePoint(ctx context.Context, userID, communityID string, point uint64, isQuest bool) error
	DecreasePoint(ctx context.Context, userID, communityID string, point uint64, isQuest bool) error
	UpdateStreak(ctx context.Context, userID, communityID string, isStreak bool) error
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

func (r *followerRepository) GetListByCommunityID(ctx context.Context, communityID string) ([]entity.Follower, error) {
	var result []entity.Follower
	err := xcontext.DB(ctx).Where("community_id=?", communityID).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *followerRepository) GetListByUserID(ctx context.Context, userID string) ([]entity.Follower, error) {
	var result []entity.Follower
	err := xcontext.DB(ctx).Where("user_id=?", userID).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *followerRepository) Create(ctx context.Context, data *entity.Follower) error {
	return xcontext.DB(ctx).
		Unscoped(). // Also find in soft deleted records.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "community_id"},
				{Name: "user_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"deleted_at": gorm.DeletedAt{Valid: false},
			}),
		}).Create(data).Error
}

func (r *followerRepository) IncreaseInviteCount(ctx context.Context, userID, communityID string) error {
	tx := xcontext.DB(ctx).
		Unscoped(). // Also update deleted record.
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

func (r *followerRepository) IncreasePoint(
	ctx context.Context, userID, communityID string, points uint64, isQuest bool,
) error {
	updateMap := map[string]any{
		"points": gorm.Expr("points+?", points),
	}

	if isQuest {
		updateMap["quests"] = gorm.Expr("quests+1")
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

func (r *followerRepository) DecreasePoint(
	ctx context.Context, userID, communityID string, points uint64, isQuest bool,
) error {
	updateMap := map[string]any{
		"points": gorm.Expr("points-?", points),
	}

	if isQuest {
		updateMap["quests"] = gorm.Expr("quests-1")
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

func (r *followerRepository) UpdateStreak(
	ctx context.Context, userID, communityID string, isStreak bool,
) error {
	updateMap := map[string]any{"streaks": gorm.Expr("streaks+1")}
	if !isStreak {
		updateMap["streaks"] = 1
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
