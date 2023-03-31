package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/router"
)

type ClaimedQuestRepository interface {
	Create(router.Context, *entity.ClaimedQuest) error
	GetByID(router.Context, string) (*entity.ClaimedQuest, error)
	GetLastPendingOrAccepted(ctx router.Context, userID, questID string) (*entity.ClaimedQuest, error)
	GetList(ctx router.Context, projectID string, offset, limit int) ([]entity.ClaimedQuest, error)
}

type claimedQuestRepository struct{}

func NewClaimedQuestRepository() *claimedQuestRepository {
	return &claimedQuestRepository{}
}

func (r *claimedQuestRepository) Create(ctx router.Context, data *entity.ClaimedQuest) error {
	if err := ctx.DB().Create(data).Error; err != nil {
		return err
	}
	return nil
}

func (r *claimedQuestRepository) GetByID(ctx router.Context, id string) (*entity.ClaimedQuest, error) {
	result := &entity.ClaimedQuest{}
	if err := ctx.DB().First(result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *claimedQuestRepository) GetLastPendingOrAccepted(
	ctx router.Context, userID, questID string,
) (*entity.ClaimedQuest, error) {
	result := entity.ClaimedQuest{}
	conditions := []entity.ClaimedQuestStatus{entity.Pending, entity.Accepted, entity.AutoAccepted}
	if err := ctx.DB().
		Where("user_id=? AND quest_id=? AND status IN (?)", userID, questID, conditions).
		Last(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *claimedQuestRepository) GetList(
	ctx router.Context, projectID string, offset, limit int,
) ([]entity.ClaimedQuest, error) {
	result := []entity.ClaimedQuest{}

	err := ctx.DB().Where("quests.project_id = ?", projectID).
		Joins("join quests on quests.id=claimed_quests.quest_id").
		Offset(offset).
		Limit(limit).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
