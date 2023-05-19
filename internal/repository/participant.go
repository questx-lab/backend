package repository

import (
	"errors"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type ParticipantRepository interface {
	Get(ctx xcontext.Context, userID, projectID string) (*entity.Participant, error)
	GetList(ctx xcontext.Context, projectID string) ([]entity.Participant, error)
	GetByReferralCode(ctx xcontext.Context, code string) (*entity.Participant, error)
	Create(ctx xcontext.Context, data *entity.Participant) error
	IncreaseInviteCount(ctx xcontext.Context, userID, projectID string) error
	IncreaseStat(ctx xcontext.Context, userID, projectID string, point, streak int) error
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

func (r *participantRepository) GetList(ctx xcontext.Context, projectID string) ([]entity.Participant, error) {
	var result []entity.Participant
	err := ctx.DB().Where("project_id=?", projectID).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *participantRepository) Create(ctx xcontext.Context, data *entity.Participant) error {
	return ctx.DB().Create(data).Error
}

func (r *participantRepository) IncreaseInviteCount(ctx xcontext.Context, userID, projectID string) error {
	tx := ctx.DB().
		Model(&entity.Participant{}).
		Where("user_id=? AND project_id=?", userID, projectID).
		Update("invite_count", gorm.Expr("invite_count+1"))

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected > 1 {
		return errors.New("the number of affected rows is invalid")
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *participantRepository) IncreaseStat(
	ctx xcontext.Context, userID, projectID string, points, streak int,
) error {
	updateMap := map[string]any{
		"points": gorm.Expr("points+?", points),
		"streak": gorm.Expr("streak+?", streak),
	}

	// Reset the streak if parameter is -1.
	if streak == -1 {
		updateMap["streak"] = 1
	}

	tx := ctx.DB().
		Model(&entity.Participant{}).
		Where("user_id=? AND project_id=?", userID, projectID).
		Updates(updateMap)

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

func (r *participantRepository) GetByReferralCode(
	ctx xcontext.Context, code string,
) (*entity.Participant, error) {
	var result entity.Participant
	if err := ctx.DB().Take(&result, "invite_code=?", code).Error; err != nil {
		return nil, err
	}

	if err := ctx.DB().Take(&result.Project, "id=?", result.ProjectID).Error; err != nil {
		return nil, err
	}

	if err := ctx.DB().Take(&result.User, "id=?", result.UserID).Error; err != nil {
		return nil, err
	}

	return &result, nil
}
