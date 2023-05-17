package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type SearchQuestFilter struct {
	Q         string
	ProjectID string
	Offset    int
	Limit     int
}

type QuestRepository interface {
	Create(ctx xcontext.Context, quest *entity.Quest) error
	GetByID(ctx xcontext.Context, id string) (*entity.Quest, error)
	GetByIDs(ctx xcontext.Context, ids []string) ([]entity.Quest, error)
	GetList(ctx xcontext.Context, filter SearchQuestFilter) ([]entity.Quest, error)
	GetTemplates(ctx xcontext.Context, filter SearchQuestFilter) ([]entity.Quest, error)
	Update(ctx xcontext.Context, data *entity.Quest) error
	Delete(ctx xcontext.Context, data *entity.Quest) error
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
	ctx xcontext.Context, filter SearchQuestFilter,
) ([]entity.Quest, error) {
	var result []entity.Quest
	tx := ctx.DB().Model(&entity.Quest{}).
		Offset(filter.Offset).
		Limit(filter.Limit).
		Order("created_at DESC")

	if filter.ProjectID != "" {
		tx = tx.Where("project_id=?", filter.ProjectID)
	} else {
		// Do not include templates in this API.
		tx = tx.Where("project_id IS NOT NULL")
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
	ctx xcontext.Context, filter SearchQuestFilter,
) ([]entity.Quest, error) {
	var result []entity.Quest
	tx := ctx.DB().Model(&entity.Quest{}).
		Offset(filter.Offset).
		Limit(filter.Limit).
		Order("created_at DESC").
		Where("project_id IS NULL")

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
	return ctx.DB().Where("id = ?", data.ID).Updates(data).Error
}

func (r *questRepository) Delete(ctx xcontext.Context, data *entity.Quest) error {
	return ctx.DB().Where("id = ?", data.ID).Delete(data).Error
}
