package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"gorm.io/gorm"
)

type QuestRepository interface {
	Create(ctx context.Context, quest *entity.Quest) error
	GetShortForm(ctx context.Context, id string) (*entity.Quest, error)
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

func (r *questRepository) GetShortForm(ctx context.Context, id string) (*entity.Quest, error) {
	result := &entity.Quest{}
	if err := r.db.Model(&entity.Quest{}).
		Select("project_id", "type", "title", "category_ids", "recurrence").
		First(result, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}
