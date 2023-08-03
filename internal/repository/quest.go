package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/domain/search"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type SearchQuestFilter struct {
	Q           string
	CategoryIDs []string
	CommunityID string
	Statuses    []entity.QuestStatusType
	Offset      int
	Limit       int
}

type StatisticQuestFilter struct {
	CommunityID string
}

type QuestRepository interface {
	Create(ctx context.Context, quest *entity.Quest) error
	GetByID(ctx context.Context, id string) (*entity.Quest, error)
	GetByIDIncludeSoftDeleted(ctx context.Context, id string) (*entity.Quest, error)
	GetByIDs(ctx context.Context, ids []string) ([]entity.Quest, error)
	GetByIDsIncludeSoftDeleted(ctx context.Context, ids []string) ([]entity.Quest, error)
	GetList(ctx context.Context, filter SearchQuestFilter) ([]entity.Quest, error)
	GetTemplates(ctx context.Context, filter SearchQuestFilter) ([]entity.Quest, error)
	Save(ctx context.Context, data *entity.Quest) error
	Delete(ctx context.Context, data *entity.Quest) error
	Count(ctx context.Context, filter StatisticQuestFilter) (int64, error)
	UpdateCategory(ctx context.Context, questID, categoryID string) error
	UpdatePosition(ctx context.Context, questID string, newPosition int) error
	IncreasePosition(ctx context.Context, communityID, categoryID string, from, to int) error
	DecreasePosition(ctx context.Context, communityID, categoryID string, from, to int) error
	RemoveQuestCategory(ctx context.Context, communityID, categoryID string) error
}

type questRepository struct {
	searchCaller client.SearchCaller
}

func NewQuestRepository(searchCaller client.SearchCaller) *questRepository {
	return &questRepository{searchCaller: searchCaller}
}

func (r *questRepository) Create(ctx context.Context, quest *entity.Quest) error {
	if err := xcontext.DB(ctx).Create(quest).Error; err != nil {
		return err
	}

	if !quest.IsTemplate {
		err := r.searchCaller.IndexQuest(ctx, quest.ID, search.QuestData{
			Title:       quest.Title,
			Description: string(quest.Description),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *questRepository) GetList(
	ctx context.Context, filter SearchQuestFilter,
) ([]entity.Quest, error) {
	if filter.Q == "" {
		var result []entity.Quest
		tx := xcontext.DB(ctx).Model(&entity.Quest{}).
			Offset(filter.Offset).
			Limit(filter.Limit).
			Order("created_at DESC").
			Order("is_highlight DESC").
			Where("is_template=false")

		if filter.CommunityID != "" {
			tx.Where("community_id=?", filter.CommunityID)
		}

		if len(filter.CategoryIDs) != 0 {
			tx.Where("category_id IN (?)", filter.CategoryIDs)
		}

		if len(filter.Statuses) == 0 {
			tx.Where("status in (?)", filter.Statuses)
		}

		if err := tx.Find(&result).Error; err != nil {
			return nil, err
		}

		return result, nil
	} else {
		ids, err := r.searchCaller.SearchQuest(ctx, filter.Q, filter.Offset, filter.Limit)
		if err != nil {
			return nil, err
		}

		quests, err := r.GetByIDs(ctx, ids)
		if err != nil {
			return nil, err
		}

		questSet := map[string]entity.Quest{}
		for _, q := range quests {
			questSet[q.ID] = q
		}

		orderedQuests := []entity.Quest{}
		for _, id := range ids {
			orderedQuests = append(orderedQuests, questSet[id])
		}

		return orderedQuests, nil
	}
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

func (r *questRepository) GetByIDIncludeSoftDeleted(ctx context.Context, id string) (*entity.Quest, error) {
	result := entity.Quest{}
	if err := xcontext.DB(ctx).Unscoped().Take(&result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *questRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.Quest, error) {
	result := []entity.Quest{}
	err := xcontext.DB(ctx).
		Order("is_highlight DESC").
		Order("created_at DESC").
		Find(&result, "id IN (?)", ids).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *questRepository) GetByIDsIncludeSoftDeleted(ctx context.Context, ids []string) ([]entity.Quest, error) {
	result := []entity.Quest{}
	err := xcontext.DB(ctx).
		Unscoped().
		Order("is_highlight DESC").
		Order("created_at DESC").
		Find(&result, "id IN (?)", ids).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *questRepository) Save(ctx context.Context, data *entity.Quest) error {
	if err := xcontext.DB(ctx).Save(data).Error; err != nil {
		return err
	}

	if data.Status == entity.QuestActive {
		err := r.searchCaller.IndexQuest(ctx, data.ID, search.QuestData{
			Title:       data.Title,
			Description: string(data.Description),
		})
		if err != nil {
			return err
		}
	} else {
		err := r.searchCaller.DeleteQuest(ctx, data.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *questRepository) Delete(ctx context.Context, data *entity.Quest) error {
	if err := xcontext.DB(ctx).Where("id=?", data.ID).Delete(data).Error; err != nil {
		return err
	}

	if err := r.searchCaller.DeleteQuest(ctx, data.ID); err != nil {
		return err
	}

	return nil
}

func (r *questRepository) Count(ctx context.Context, filter StatisticQuestFilter) (int64, error) {
	tx := xcontext.DB(ctx).Model(&entity.Quest{})

	if filter.CommunityID != "" {
		tx = tx.Where("community_id=?", filter.CommunityID)
	}

	var result int64
	if err := tx.Count(&result).Error; err != nil {
		return 0, err
	}

	return result, nil
}

func (r *questRepository) UpdateCategory(ctx context.Context, questID, categoryID string) error {
	tx := xcontext.DB(ctx).Model(&entity.Quest{}).
		Where("id=?", questID)

	if categoryID == "" {
		return tx.Update("category_id", nil).Error
	}

	return tx.Update("category_id", categoryID).Error
}

func (r *questRepository) UpdatePosition(ctx context.Context, questID string, newPosition int) error {
	return xcontext.DB(ctx).Model(&entity.Quest{}).
		Where("id=?", questID).
		Update("position", newPosition).Error
}

func (r *questRepository) IncreasePosition(
	ctx context.Context, communityID, categoryID string, from, to int,
) error {
	tx := xcontext.DB(ctx).Model(&entity.Quest{})

	if from != -1 {
		tx.Where("position >= ?", from)
	}

	if to != -1 {
		tx.Where("position <= ?", to)
	}

	if communityID == "" {
		tx.Where("community_id IS NULL")
	} else {
		tx.Where("community_id=?", communityID)
	}

	if categoryID == "" {
		tx.Where("category_id IS NULL")
	} else {
		tx.Where("category_id=?", categoryID)
	}

	if err := tx.Update("position", gorm.Expr("position+?", 1)).Error; err != nil {
		return err
	}

	return nil
}

func (r *questRepository) DecreasePosition(
	ctx context.Context, communityID, categoryID string, from, to int,
) error {
	tx := xcontext.DB(ctx).Model(&entity.Quest{}).
		Where("community_id=?", communityID)

	if from != -1 {
		tx.Where("position >= ?", from)
	}

	if to != -1 {
		tx.Where("position <= ?", to)
	}

	if categoryID == "" {
		tx.Where("category_id IS NULL")
	} else {
		tx.Where("category_id=?", categoryID)
	}

	if err := tx.Update("position", gorm.Expr("position-?", 1)).Error; err != nil {
		return err
	}

	return nil
}

func (r *questRepository) RemoveQuestCategory(ctx context.Context, communityID, categoryID string) error {
	return xcontext.DB(ctx).Model(&entity.Quest{}).
		Where("community_id=? AND category_id=?", communityID, categoryID).
		Update("category_id=?", nil).Error
}
