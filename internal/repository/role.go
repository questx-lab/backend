package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type RoleRepository interface {
	Create(context.Context, *entity.Role) error
	UpdateByID(context.Context, string, *entity.Role) error
	DeleteByID(context.Context, string) error
	GetByID(context.Context, string) (*entity.Role, error)
	GetByIDs(context.Context, []string) ([]entity.Role, error)
	GetByName(context.Context, string) (*entity.Role, error)
	GetByNames(context.Context, []string) ([]entity.Role, error)
	GetByCommunityID(context.Context, string) ([]entity.Role, error)
	GetLatestPriorityByCommunityID(context.Context, string) (*entity.Role, error)
}

type roleRepository struct{}

func NewRoleRepository() RoleRepository {
	return &roleRepository{}
}

func (r *roleRepository) Create(ctx context.Context, e *entity.Role) error {
	if err := xcontext.DB(ctx).Create(e).Error; err != nil {
		return err
	}

	return nil
}

func (r *roleRepository) UpdateByID(ctx context.Context, id string, e *entity.Role) error {
	if err := xcontext.DB(ctx).Model(e).Where("id = ?", id).Updates(e).Error; err != nil {
		return err
	}

	return nil
}

func (r *roleRepository) GetByID(ctx context.Context, id string) (*entity.Role, error) {
	result := entity.Role{}
	if err := xcontext.DB(ctx).Take(&result, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *roleRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.Role, error) {
	result := []entity.Role{}
	err := xcontext.DB(ctx).
		Find(&result, "id IN (?)", ids).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	result := entity.Role{}
	if err := xcontext.DB(ctx).Take(&result, "name = ?", name).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *roleRepository) GetByNames(ctx context.Context, names []string) ([]entity.Role, error) {
	result := []entity.Role{}
	err := xcontext.DB(ctx).
		Find(&result, "name IN (?)", names).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *roleRepository) GetByCommunityID(ctx context.Context, communityID string) ([]entity.Role, error) {
	result := []entity.Role{}
	err := xcontext.DB(ctx).
		Find(&result, "community_id=?", communityID).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *roleRepository) DeleteByID(ctx context.Context, id string) error {
	if err := xcontext.DB(ctx).Delete(&entity.Role{}, "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}

func (r *roleRepository) GetLatestPriorityByCommunityID(ctx context.Context, communityID string) (*entity.Role, error) {
	var result entity.Role
	if err := xcontext.DB(ctx).Order("priority").Take(&result, "community_id = ?", communityID).Error; err != nil {
		return nil, err
	}

	return &result, nil
}
