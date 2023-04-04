package repository

import (
	"errors"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type ParticipantRepository interface {
	Get(ctx xcontext.Context, userID, projectID string) (*entity.Participant, error)
	Create(ctx xcontext.Context, userID, projectID string) error
	Increase(ctx xcontext.Context, userID, projectID string, point uint64) error
}

type participantRepository struct{}

func NewParticipantRepository() *participantRepository {
	return &participantRepository{}
}

func (r *participantRepository) Get(ctx xcontext.Context, userID, projectID string) (*entity.Participant, error) {
	var result entity.Participant
	err := ctx.DB().Where("user_id=? AND project_id=?", userID, projectID).Take(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *participantRepository) Create(ctx xcontext.Context, userID, projectID string) error {
	return ctx.DB().Create(&entity.Participant{
		UserID:    userID,
		ProjectID: projectID,
		Points:    0,
	}).Error
}

func (r *participantRepository) Increase(ctx xcontext.Context, userID, projectID string, points uint64) error {
	tx := ctx.DB().
		Model(&entity.Participant{}).
		Where("user_id=? AND project_id=?", userID, projectID).
		Update("points", gorm.Expr("points+?", points))

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected > 1 {
		return errors.New("the number of rows effected is invalid")
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
