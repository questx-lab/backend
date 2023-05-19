package repository

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type ClaimedQuestFilter struct {
	QuestIDs    []string
	UserIDs     []string
	Status      []entity.ClaimedQuestStatus
	Recurrences []entity.RecurrenceType
	Offset      int
	Limit       int
}

type GetLastClaimedQuestFilter struct {
	UserID    string
	QuestID   string
	ProjectID string
	Status    []entity.ClaimedQuestStatus
}

type ClaimedQuestRepository interface {
	Create(context.Context, *entity.ClaimedQuest) error
	GetByID(context.Context, string) (*entity.ClaimedQuest, error)
	GetByIDs(context.Context, []string) ([]entity.ClaimedQuest, error)
	GetLast(ctx context.Context, filter GetLastClaimedQuestFilter) (*entity.ClaimedQuest, error)
	GetList(ctx context.Context, projectID string, filter *ClaimedQuestFilter) ([]entity.ClaimedQuest, error)
	UpdateReviewByIDs(ctx context.Context, ids []string, data *entity.ClaimedQuest) error
}

type claimedQuestRepository struct{}

func NewClaimedQuestRepository() ClaimedQuestRepository {
	return &claimedQuestRepository{}
}

func (r *claimedQuestRepository) Create(ctx context.Context, data *entity.ClaimedQuest) error {
	if err := xcontext.DB(ctx).Create(data).Error; err != nil {
		return err
	}
	return nil
}

func (r *claimedQuestRepository) GetByID(ctx context.Context, id string) (*entity.ClaimedQuest, error) {
	result := &entity.ClaimedQuest{}
	if err := xcontext.DB(ctx).Take(result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *claimedQuestRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.ClaimedQuest, error) {
	result := []entity.ClaimedQuest{}
	if err := xcontext.DB(ctx).Find(&result, "id IN (?)", ids).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *claimedQuestRepository) GetLast(
	ctx context.Context, filter GetLastClaimedQuestFilter,
) (*entity.ClaimedQuest, error) {
	result := entity.ClaimedQuest{}
	tx := xcontext.DB(ctx).
		Order("claimed_quests.created_at DESC").
		Joins("join quests on quests.id = claimed_quests.quest_id")

	if filter.UserID != "" {
		tx = tx.Where("claimed_quests.user_id=?", filter.UserID)
	}

	if filter.QuestID != "" {
		tx = tx.Where("claimed_quests.quest_id=?", filter.QuestID)
	}

	if filter.ProjectID != "" {
		tx = tx.Where("quests.project_id=?", filter.ProjectID)
	}

	if len(filter.Status) > 0 {
		tx = tx.Where("claimed_quests.status in (?)", filter.Status)
	}

	if err := tx.Take(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *claimedQuestRepository) GetList(
	ctx context.Context,
	projectID string,
	filter *ClaimedQuestFilter,
) ([]entity.ClaimedQuest, error) {
	result := []entity.ClaimedQuest{}
	tx := xcontext.DB(ctx).
		Joins("join quests on quests.id = claimed_quests.quest_id").
		Where("quests.project_id=?", projectID).
		Offset(filter.Offset).
		Limit(filter.Limit).
		Order("claimed_quests.created_at ASC")

	if len(filter.Status) > 0 {
		tx.Where("claimed_quests.status IN (?)", filter.Status)
	}

	if len(filter.Recurrences) > 0 {
		tx.Where("quests.recurrence IN (?)", filter.Recurrences)
	}

	if len(filter.QuestIDs) > 0 {
		tx.Where("claimed_quests.quest_id IN (?)", filter.QuestIDs)
	}

	if len(filter.UserIDs) > 0 {
		tx.Where("claimed_quests.user_id IN (?)", filter.UserIDs)
	}

	err := tx.Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *claimedQuestRepository) UpdateReviewByIDs(ctx context.Context, ids []string, data *entity.ClaimedQuest) error {
	tx := xcontext.DB(ctx).Model(&entity.ClaimedQuest{}).Where("id IN (?)", ids).Updates(data)
	if err := tx.Error; err != nil {
		return err
	}

	if int(tx.RowsAffected) != len(ids) {
		return fmt.Errorf("update status not exec correctly")
	}

	return nil
}
