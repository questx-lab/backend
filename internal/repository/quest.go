package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"gorm.io/gorm"
)

type QuestRepository interface {
	Create(ctx context.Context, quest *entity.Quest) error
	GetByID(ctx context.Context, id string) (*entity.Quest, error)
	GetListShortForm(ctx context.Context, projectID string, offset int, limit int) ([]entity.Quest, error)
}

type questRepository struct {
	db *gorm.DB
}

func NewQuestRepository(db *gorm.DB) *questRepository {
	return &questRepository{db: db}
}

func (r *questRepository) Create(ctx context.Context, quest *entity.Quest) error {
	if err := r.db.Create(quest).Error; err != nil {
		return err
	}

	return nil
}

func (r *questRepository) GetListShortForm(
	ctx context.Context, projectID string, offset int, limit int,
) ([]entity.Quest, error) {
	var result []entity.Quest
	err := r.db.Model(&entity.Quest{}).
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

func (r *questRepository) GetByID(ctx context.Context, id string) (*entity.Quest, error) {
	result := &entity.Quest{}
	if err := r.db.First(result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}
