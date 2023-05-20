package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type SearchQuestFilter struct {
	Q          string
	ProjectID  string
	CategoryID string
	Offset     int
	Limit      int
}

type QuestRepository interface {
	Create(ctx context.Context, quest *entity.Quest) error
	GetByID(ctx context.Context, id string) (*entity.Quest, error)
	GetByIDs(ctx context.Context, ids []string) ([]entity.Quest, error)
	GetList(ctx context.Context, filter SearchQuestFilter) ([]entity.Quest, error)
	GetTemplates(ctx context.Context, filter SearchQuestFilter) ([]entity.Quest, error)
	Update(ctx context.Context, data *entity.Quest) error
	Delete(ctx context.Context, data *entity.Quest) error
}

type questRepository struct{}

func NewQuestRepository() *questRepository {
	return &questRepository{}
}

func (r *questRepository) Create(ctx context.Context, quest *entity.Quest) error {
	if err := xcontext.DB(ctx).Create(quest).Error; err != nil {
		return err
	}

	return nil
}

func (r *questRepository) GetList(
	ctx context.Context, filter SearchQuestFilter,
) ([]entity.Quest, error) {
	var result []entity.Quest
	tx := xcontext.DB(ctx).Model(&entity.Quest{}).
		Offset(filter.Offset).
		Limit(filter.Limit).
		Order("created_at DESC").
		Order("is_highlight DESC").
		Where("is_template=false")

	if filter.ProjectID != "" {
		tx = tx.Where("project_id=?", filter.ProjectID)
	}

	if filter.CategoryID != "" {
		tx = tx.Where("category_id=?", filter.CategoryID)
	}

	if filter.Q != "" {
		tx = tx.Select("*, MATCH(title,description) AGAINST (?) as score", filter.Q).
			Where("MATCH(title,description) AGAINST (?) > 0", filter.Q).
			Order("score DESC")
	}

	if err := tx.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *questRepository) GetTemplates(
	ctx context.Context, filter SearchQuestFilter,
) ([]entity.Quest, error) {
	var result []entity.Quest
	tx := xcontext.DB(ctx).Model(&entity.Quest{}).
		Offset(filter.Offset).
		Limit(filter.Limit).
		Order("created_at DESC").
		Where("is_template=true")

	if filter.Q != "" {
		tx = tx.Select("*, MATCH(title,description) AGAINST (?) as score", filter.Q).
			Where("MATCH(title,description) AGAINST (?) > 0", filter.Q).
			Order("score DESC")
	}

	if err := tx.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *questRepository) GetByID(ctx context.Context, id string) (*entity.Quest, error) {
	result := entity.Quest{}
	if err := xcontext.DB(ctx).Take(&result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *questRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.Quest, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	result := []entity.Quest{}
	if err := xcontext.DB(ctx).Find(&result, "id IN (?)", ids).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *questRepository) Update(ctx context.Context, data *entity.Quest) error {
	return xcontext.DB(ctx).
		Omit("is_template", "created_at", "updated_at", "deleted_at", "id").
		Where("id = ?", data.ID).
		Updates(data).Error
}

func (r *questRepository) Delete(ctx context.Context, data *entity.Quest) error {
	return xcontext.DB(ctx).Where("id = ?", data.ID).Delete(data).Error
}
