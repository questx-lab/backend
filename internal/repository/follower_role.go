package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type FollowerRoleRepository interface {
	Get(ctx context.Context, userID, communityID string) ([]entity.FollowerRole, error)
	GetByCommunityAndUserIDs(ctx context.Context, communityID string, userIDs []string) ([]entity.FollowerRole, error)
	GetOwners(ctx context.Context, userID string) ([]entity.FollowerRole, error)
	GetFirstByRole(ctx context.Context, communityID, roleID string) (*entity.FollowerRole, error)
	Create(ctx context.Context, data *entity.FollowerRole) error
	Delete(ctx context.Context, userID, communityID, roleID string) error
	DeleteByRoles(ctx context.Context, userID, communityID string, roleIDs []string) error
}

type followerRoleRepository struct{}

func NewFollowerRoleRepository() FollowerRoleRepository {
	return &followerRoleRepository{}
}

func (r *followerRoleRepository) Get(ctx context.Context, userID, communityID string) ([]entity.FollowerRole, error) {
	var result []entity.FollowerRole
	err := xcontext.DB(ctx).
		Where("user_id=? AND community_id=?", userID, communityID).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *followerRoleRepository) GetByCommunityAndUserIDs(
	ctx context.Context, communityID string, userIDs []string,
) ([]entity.FollowerRole, error) {
	var result []entity.FollowerRole
	err := xcontext.DB(ctx).
		Where("user_id IN (?) AND community_id=?", userIDs, communityID).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *followerRoleRepository) GetOwners(ctx context.Context, userID string) ([]entity.FollowerRole, error) {
	var result []entity.FollowerRole
	err := xcontext.DB(ctx).
		Where("user_id=? AND role_id=?", userID, entity.OwnerBaseRole).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *followerRoleRepository) GetFirstByRole(ctx context.Context, communityID, roleID string) (*entity.FollowerRole, error) {
	var result entity.FollowerRole
	err := xcontext.DB(ctx).
		Where("community_id=? AND role_id=?", communityID, roleID).
		Take(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *followerRoleRepository) Create(ctx context.Context, data *entity.FollowerRole) error {
	return xcontext.DB(ctx).Create(data).Error
}

func (r *followerRoleRepository) Delete(ctx context.Context, userID, communityID, roleID string) error {
	return xcontext.DB(ctx).
		Where("user_id=? AND community_id=? AND role_id=?", userID, communityID, roleID).
		Delete(&entity.FollowerRole{}).Error
}

func (r *followerRoleRepository) DeleteByRoles(ctx context.Context, userID, communityID string, roleIDs []string) error {
	return xcontext.DB(ctx).
		Where("user_id = ? AND community_id = ? AND role_id IN (?)", userID, communityID, roleIDs).
		Delete(&entity.FollowerRole{}).Error
}
