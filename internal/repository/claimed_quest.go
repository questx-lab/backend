package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"gorm.io/gorm"
)

type ClaimedQuestRepository interface {
	Create(context.Context, *entity.ClaimedQuest) error
	GetByID(context.Context, string) (*entity.ClaimedQuest, error)
	GetLastPendingOrAccepted(ctx context.Context, userID, questID string) (*entity.ClaimedQuest, error)
	GetList(ctx context.Context, projectID string, offset, limit int) ([]entity.ClaimedQuest, error)
}

type claimedQuestRepository struct {
	db *gorm.DB
}

func NewClaimedQuestRepository(db *gorm.DB) *claimedQuestRepository {
	return &claimedQuestRepository{db: db}
}

func (r *claimedQuestRepository) Create(ctx context.Context, data *entity.ClaimedQuest) error {
	if err := r.db.Create(data).Error; err != nil {
		return err
	}
	return nil
}

func (r *claimedQuestRepository) GetByID(ctx context.Context, id string) (*entity.ClaimedQuest, error) {
	result := &entity.ClaimedQuest{}
	if err := r.db.First(result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *claimedQuestRepository) GetLastPendingOrAccepted(
	ctx context.Context, userID, questID string,
) (*entity.ClaimedQuest, error) {
	result := entity.ClaimedQuest{}
	conditions := []entity.ClaimedQuestStatus{entity.Pending, entity.Accepted, entity.AutoAccepted}
	if err := r.db.
		Where("user_id=? AND quest_id=? AND status IN (?)", userID, questID, conditions).
		Last(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *claimedQuestRepository) GetList(
	ctx context.Context, projectID string, offset, limit int,
) ([]entity.ClaimedQuest, error) {
	result := []entity.ClaimedQuest{}

	err := r.db.Where("quests.project_id = ?", projectID).
		Joins("join quests on quests.id=claimed_quests.quest_id").
		Offset(offset).
		Limit(limit).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
