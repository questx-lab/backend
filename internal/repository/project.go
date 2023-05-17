package repository

import (
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GetListProjectFilter struct {
	Q      string
	Offset int
	Limit  int
}

type ProjectRepository interface {
	Create(ctx xcontext.Context, e *entity.Project) error
	GetList(ctx xcontext.Context, filter GetListProjectFilter) ([]entity.Project, error)
	GetByID(ctx xcontext.Context, id string) (*entity.Project, error)
	GetByName(ctx xcontext.Context, name string) (*entity.Project, error)
	UpdateByID(ctx xcontext.Context, id string, e *entity.Project) error
	DeleteByID(ctx xcontext.Context, id string) error
	GetFollowingList(ctx xcontext.Context, userID string, offset, limit int) ([]entity.Project, error)
}

type projectRepository struct{}

func NewProjectRepository() ProjectRepository {
	return &projectRepository{}
}

func (r *projectRepository) Create(ctx xcontext.Context, e *entity.Project) error {
	if err := ctx.DB().Model(e).Create(e).Error; err != nil {
		return err
	}

	return nil
}

func (r *projectRepository) GetList(ctx xcontext.Context, filter GetListProjectFilter) ([]entity.Project, error) {
	var result []entity.Project
	tx := ctx.DB().
		Limit(filter.Limit).
		Offset(filter.Offset)

	if filter.Q != "" {
		tx = tx.Select("*, MATCH(name,introduction) AGAINST (?) as score", filter.Q).
			Where("MATCH(name,introduction) AGAINST (?) > 0", filter.Q).
			Order("score DESC")
	}

	if err := tx.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) GetByID(ctx xcontext.Context, id string) (*entity.Project, error) {
	result := &entity.Project{}
	if err := ctx.DB().Take(result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) GetByName(ctx xcontext.Context, name string) (*entity.Project, error) {
	result := &entity.Project{}
	if err := ctx.DB().Take(result, "name=?", name).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) UpdateByID(ctx xcontext.Context, id string, e *entity.Project) error {
	tx := ctx.DB().
		Model(&entity.Project{}).
		Where("id=?", id).
		Omit("created_by", "created_at", "id").
		Updates(*e)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	return nil
}

func (r *projectRepository) DeleteByID(ctx xcontext.Context, id string) error {
	tx := ctx.DB().
		Delete(&entity.Project{}, "id=?", id)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	return nil
}

func (r *projectRepository) GetFollowingList(ctx xcontext.Context, userID string, offset, limit int) ([]entity.Project, error) {
	var result []entity.Project
	if err := ctx.DB().
		Joins("join participants on projects.id = participants.project_id").
		Where("participants.user_id=?", userID).
		Limit(limit).Offset(offset).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}
