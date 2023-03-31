package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type QuestRepository interface {
	Create(ctx xcontext.Context, quest *entity.Quest) error
	GetByID(ctx xcontext.Context, id string) (*entity.Quest, error)
	GetList(ctx xcontext.Context, projectID string, offset int, limit int) ([]entity.Quest, error)
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
		Select("id", "type", "title", "status", "category_ids", "recurrence").
		Where("project_id=?", projectID).
		Offset(offset).
		Limit(limit).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *questRepository) GetByID(ctx xcontext.Context, id string) (*entity.Quest, error) {
	result := entity.Quest{}
	if err := ctx.DB().First(&result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return &result, nil
}
