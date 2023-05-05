package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type QuestRepository interface {
	Create(ctx xcontext.Context, quest *entity.Quest) error
	GetByID(ctx xcontext.Context, id string) (*entity.Quest, error)
	GetByIDs(ctx xcontext.Context, ids []string) ([]entity.Quest, error)
	GetList(ctx xcontext.Context, projectID string, offset int, limit int) ([]entity.Quest, error)
	Update(ctx xcontext.Context, data *entity.Quest) error
}

type questRepository struct{}

func NewQuestRepository() *questRepository {
	return &questRepository{}
}

func (r *questRepository) Create(ctx xcontext.Context, quest *entity.Quest) error {
	if err := ctx.DB().Create(quest).Error; err != nil {
		return err
	}

	return nil
}

func (r *questRepository) GetList(
	ctx xcontext.Context, projectID string, offset int, limit int,
) ([]entity.Quest, error) {
	var result []entity.Quest
	err := ctx.DB().Model(&entity.Quest{}).
		Where("project_id=?", projectID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *questRepository) GetByID(ctx xcontext.Context, id string) (*entity.Quest, error) {
	result := entity.Quest{}
	if err := ctx.DB().Take(&result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *questRepository) GetByIDs(ctx xcontext.Context, ids []string) ([]entity.Quest, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	result := []entity.Quest{}
	if err := ctx.DB().Find(&result, "id IN (?)", ids).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *questRepository) Update(ctx xcontext.Context, data *entity.Quest) error {
	return ctx.DB().Where("id=?", data.ID).Updates(data).Error
}
