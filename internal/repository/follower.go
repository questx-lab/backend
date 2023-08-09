package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StatisticFollowerFilter struct {
	UserID string
}

type GetListFollowerFilter struct {
	CommunityID    string
	Q              string
	IgnoreUserRole bool
	Offset         int
	Limit          int
}

type FollowerRepository interface {
	Get(ctx context.Context, userID, communityID string) (*entity.Follower, error)
	GetListByCommunityID(ctx context.Context, filter GetListFollowerFilter) ([]entity.Follower, error)
	GetListByUserID(ctx context.Context, userID string) ([]entity.Follower, error)
	GetByReferralCode(ctx context.Context, code string) (*entity.Follower, error)
	Create(ctx context.Context, data *entity.Follower) error
	IncreaseInviteCount(ctx context.Context, userID, communityID string) error
	DecreaseInviteCount(ctx context.Context, userID, communityID string) error
	IncreasePoint(ctx context.Context, userID, communityID string, point uint64, isQuest bool) error
	DecreasePoint(ctx context.Context, userID, communityID string, point uint64, isQuest bool) error
	CreateStreak(ctx context.Context, userID, communityID string, startTime time.Time) error
	GetLastStreak(ctx context.Context, userID, communityID string) (*entity.FollowerStreak, error)
	GetStreaks(ctx context.Context, userID, communityID string, begin, end time.Time) ([]entity.FollowerStreak, error)
	Count(ctx context.Context, filter StatisticFollowerFilter) (int64, error)
	IncreaseChatXP(ctx context.Context, userID, communityID string, xp int) error
	UpdateChatLevel(ctx context.Context, userID, communityID string, level int, thresholdXP int) error
	Delete(ctx context.Context, userID, communityID string) error
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

func (r *followerRepository) GetListByCommunityID(
	ctx context.Context, filter GetListFollowerFilter,
) ([]entity.Follower, error) {
	var result []entity.Follower
	tx := xcontext.DB(ctx).Model(&entity.Follower{}).
		Joins("join users on users.id=followers.user_id").
		Offset(filter.Offset).Limit(filter.Limit)

	if filter.CommunityID != "" {
		tx.Where("followers.community_id=?", filter.CommunityID)
	}

	if filter.Q != "" {
		tx.Where("users.name LIKE ?", filter.Q+"%")
	}

	if filter.IgnoreUserRole {
		tx.Joins("join follower_roles on follower_roles.user_id=followers.user_id AND follower_roles.community_id=followers.community_id").
			Where("follower_roles.role_id != ?", entity.UserBaseRole).Group("followers.user_id")
	}

	if err := tx.Find(&result).Error; err != nil {
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
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "community_id"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"deleted_at": sql.NullTime{},
			}),
		}).Create(data).Error
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

func (r *followerRepository) DecreaseInviteCount(ctx context.Context, userID, communityID string) error {
	tx := xcontext.DB(ctx).
		Model(&entity.Follower{}).
		Where("user_id = ? AND community_id = ?", userID, communityID).
		Update("invite_count", gorm.Expr("invite_count-1"))

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
		Where("user_id=? AND community_id=? AND points >= ?", userID, communityID, points).
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

func (r *followerRepository) IncreaseChatXP(
	ctx context.Context, userID, communityID string, xp int,
) error {
	tx := xcontext.DB(ctx).
		Model(&entity.Follower{}).
		Where("user_id=? AND community_id=?", userID, communityID).
		Updates(map[string]any{
			"total_chat_xp":   gorm.Expr("total_chat_xp+?", xp),
			"current_chat_xp": gorm.Expr("current_chat_xp+?", xp),
		})

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

func (r *followerRepository) UpdateChatLevel(
	ctx context.Context, userID, communityID string, level int, thresholdXP int,
) error {
	tx := xcontext.DB(ctx).
		Model(&entity.Follower{}).
		Where("user_id=? AND community_id=?", userID, communityID).
		Where("current_chat_xp >= ? AND chat_level=?", thresholdXP, level-1).
		Updates(map[string]any{
			"chat_level":      level,
			"current_chat_xp": gorm.Expr("current_chat_xp-?", thresholdXP),
		})

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

func (r *followerRepository) CreateStreak(
	ctx context.Context, userID, communityID string, startTime time.Time,
) error {
	return xcontext.DB(ctx).Model(&entity.FollowerStreak{}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "community_id"},
				{Name: "start_time"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"streaks": gorm.Expr("streaks+1"),
			}),
		}).
		Create(&entity.FollowerStreak{
			UserID:      userID,
			CommunityID: communityID,
			StartTime:   startTime,
			Streaks:     1,
		}).Error
}

func (r *followerRepository) GetLastStreak(
	ctx context.Context, userID, communityID string,
) (*entity.FollowerStreak, error) {
	var streak entity.FollowerStreak
	err := xcontext.DB(ctx).
		Where("user_id=? AND community_id=?", userID, communityID).
		Order("start_time DESC").
		Take(&streak).Error
	if err != nil {
		return nil, err
	}

	return &streak, nil
}

func (r *followerRepository) GetStreaks(
	ctx context.Context, userID, communityID string, begin, end time.Time,
) ([]entity.FollowerStreak, error) {
	var streaks []entity.FollowerStreak
	err := xcontext.DB(ctx).
		Where("user_id=? AND community_id=?", userID, communityID).
		Where("start_time>=? AND start_time <=?", begin, end).
		Find(&streaks).Error
	if err != nil {
		return nil, err
	}

	return streaks, nil
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

func (r *followerRepository) Count(ctx context.Context, filter StatisticFollowerFilter) (int64, error) {
	tx := xcontext.DB(ctx).Model(&entity.Follower{})

	if filter.UserID != "" {
		tx = tx.Where("user_id = ?", filter.UserID)
	}

	var result int64
	if err := tx.Count(&result).Error; err != nil {
		return 0, err
	}

	return result, nil
}

func (r *followerRepository) Delete(ctx context.Context, userID, communityID string) error {
	if err := xcontext.DB(ctx).Delete(&entity.Follower{}, "community_id = ? AND user_id = ?", communityID, userID).Error; err != nil {
		return err
	}

	return nil
}
