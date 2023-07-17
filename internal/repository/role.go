package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type RoleRepository interface {
	CreateRole(context.Context, *entity.Role) error
	UpdateRoleByID(context.Context, string, *entity.Role) error
	GetRoleByID(context.Context, string) (*entity.Role, error)
	GetRoleByName(context.Context, string) (*entity.Role, error)
	GetRoleByNames(context.Context, []string) ([]*entity.Role, error)
}

type roleRepository struct{}

func NewRoleRepository() RoleRepository {
	return &roleRepository{}
}

func (r *roleRepository) CreateRole(ctx context.Context, e *entity.Role) error {
	if err := xcontext.DB(ctx).Create(e).Error; err != nil {
		return err
	}

	return nil
}

func (r *roleRepository) UpdateRoleByID(ctx context.Context, id string, e *entity.Role) error {
	if err := xcontext.DB(ctx).Model(e).Where("id = ?", id).Update("id", e).Error; err != nil {
		return err
	}

	return nil
}

func (r *roleRepository) GetRoleByID(ctx context.Context, id string) (*entity.Role, error) {
	result := entity.Role{}
	if err := xcontext.DB(ctx).Take(&result, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *roleRepository) GetRoleByName(ctx context.Context, name string) (*entity.Role, error) {
	result := entity.Role{}
	if err := xcontext.DB(ctx).Take(&result, "name = ?", name).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *roleRepository) GetRoleByNames(ctx context.Context, ids []string) ([]*entity.Role, error) {
	result := []*entity.Role{}
	err := xcontext.DB(ctx).
		Find(&result, "id IN (?)", ids).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
