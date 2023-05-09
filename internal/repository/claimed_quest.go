package repository

import (
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

type ClaimedQuestRepository interface {
	Create(xcontext.Context, *entity.ClaimedQuest) error
	GetByID(xcontext.Context, string) (*entity.ClaimedQuest, error)
	GetByIDs(xcontext.Context, []string) ([]entity.ClaimedQuest, error)
	GetLast(ctx xcontext.Context, userID, questID string) (*entity.ClaimedQuest, error)
	GetLastPendingOrAccepted(ctx xcontext.Context, userID, questID string) (*entity.ClaimedQuest, error)
	GetList(ctx xcontext.Context, projectID string, filter *ClaimedQuestFilter) ([]entity.ClaimedQuest, error)
	UpdateReviewByIDs(ctx xcontext.Context, ids []string, data *entity.ClaimedQuest) error
}

type claimedQuestRepository struct{}

func NewClaimedQuestRepository() ClaimedQuestRepository {
	return &claimedQuestRepository{}
}

func (r *claimedQuestRepository) Create(ctx xcontext.Context, data *entity.ClaimedQuest) error {
	if err := ctx.DB().Create(data).Error; err != nil {
		return err
	}
	return nil
}

func (r *claimedQuestRepository) GetByID(ctx xcontext.Context, id string) (*entity.ClaimedQuest, error) {
	result := &entity.ClaimedQuest{}
	if err := ctx.DB().Take(result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *claimedQuestRepository) GetByIDs(ctx xcontext.Context, ids []string) ([]entity.ClaimedQuest, error) {
	result := []entity.ClaimedQuest{}
	if err := ctx.DB().Find(&result, "id IN (?)", ids).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *claimedQuestRepository) GetLastPendingOrAccepted(
	ctx xcontext.Context, userID, questID string,
) (*entity.ClaimedQuest, error) {
	result := entity.ClaimedQuest{}
	status := []entity.ClaimedQuestStatus{entity.Pending, entity.Accepted, entity.AutoAccepted}
	if err := ctx.DB().
		Where("user_id=? AND quest_id=? AND status IN (?)", userID, questID, status).
		Order("created_at desc").
		Take(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *claimedQuestRepository) GetLast(
	ctx xcontext.Context, userID, questID string,
) (*entity.ClaimedQuest, error) {
	result := entity.ClaimedQuest{}
	if err := ctx.DB().
		Where("user_id=? AND quest_id=?", userID, questID).
		Order("created_at desc").
		Take(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *claimedQuestRepository) GetList(
	ctx xcontext.Context,
	projectID string,
	filter *ClaimedQuestFilter,
) ([]entity.ClaimedQuest, error) {
	result := []entity.ClaimedQuest{}
	tx := ctx.DB().
		Joins("join quests on quests.id = claimed_quests.quest_id").
		Where("quests.project_id = ?", projectID).
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

func (r *claimedQuestRepository) UpdateReviewByIDs(ctx xcontext.Context, ids []string, data *entity.ClaimedQuest) error {
	tx := ctx.DB().Model(&entity.ClaimedQuest{}).Where("id IN (?)", ids).Updates(data)
	if err := tx.Error; err != nil {
		return err
	}

	if int(tx.RowsAffected) != len(ids) {
		return fmt.Errorf("update status not exec correctly")
	}

	return nil
}
