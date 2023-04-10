package repository

import (
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type ClaimedQuestFilter struct {
	ProjectID string
	Status    entity.ClaimedQuestStatus
}

type ClaimedQuestRepository interface {
	Create(xcontext.Context, *entity.ClaimedQuest) error
	GetByID(xcontext.Context, string) (*entity.ClaimedQuest, error)
	GetLast(ctx xcontext.Context, userID, questID string) (*entity.ClaimedQuest, error)
	GetLastPendingOrAccepted(ctx xcontext.Context, userID, questID string) (*entity.ClaimedQuest, error)
	GetList(ctx xcontext.Context, filter *ClaimedQuestFilter, offset, limit int) ([]entity.ClaimedQuest, error)
	UpdateReviewByID(ctx xcontext.Context, id string, data *entity.ClaimedQuest) error
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

func (r *claimedQuestRepository) GetLastPendingOrAccepted(
	ctx xcontext.Context, userID, questID string,
) (*entity.ClaimedQuest, error) {
	result := entity.ClaimedQuest{}
	conditions := []entity.ClaimedQuestStatus{entity.Pending, entity.Accepted, entity.AutoAccepted}
	if err := ctx.DB().
		Where("user_id=? AND quest_id=? AND status IN (?)", userID, questID, conditions).
		Order("created_at").
		Last(&result).Error; err != nil {
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
		Order("created_at").
		Last(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *claimedQuestRepository) GetList(
	ctx xcontext.Context,
	filter *ClaimedQuestFilter,
	offset, limit int,
) ([]entity.ClaimedQuest, error) {
	result := []entity.ClaimedQuest{}
	tx := ctx.
		DB().
		Joins("join quests on quests.id = claimed_quests.quest_id").
		Where("quests.project_id = ?", filter.ProjectID)

	if filter.Status != "" {
		tx.Where("claimed_quests.status = ?", filter.Status)
	}

	err := tx.
		Offset(offset).
		Limit(limit).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *claimedQuestRepository) UpdateReviewByID(ctx xcontext.Context, id string, data *entity.ClaimedQuest) error {
	tx := ctx.DB().Model(&entity.ClaimedQuest{}).Where("id = ?", id).Updates(data)
	if err := tx.Error; err != nil {
		return err
	}
	if tx.RowsAffected != 1 {
		return fmt.Errorf("update status not exec correctly")
	}
	return nil
}
