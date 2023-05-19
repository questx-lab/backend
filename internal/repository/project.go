package repository

import (
	"errors"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type GetListProjectFilter struct {
	Q              string
	ReferredBy     string
	ReferralStatus entity.ReferralStatusType
	Offset         int
	Limit          int
}

type ProjectRepository interface {
	Create(ctx xcontext.Context, e *entity.Project) error
	GetList(ctx xcontext.Context, filter GetListProjectFilter) ([]entity.Project, error)
	GetByID(ctx xcontext.Context, id string) (*entity.Project, error)
	GetByName(ctx xcontext.Context, name string) (*entity.Project, error)
	UpdateByID(ctx xcontext.Context, id string, e entity.Project) error
	GetByIDs(ctx xcontext.Context, ids []string) ([]entity.Project, error)
	UpdateReferralStatusByIDs(ctx xcontext.Context, ids []string, status entity.ReferralStatusType) error
	DeleteByID(ctx xcontext.Context, id string) error
	GetFollowingList(ctx xcontext.Context, userID string, offset, limit int) ([]entity.Project, error)
	IncreaseFollowers(ctx xcontext.Context, projectID string) error
}

type projectRepository struct{}

func NewProjectRepository() ProjectRepository {
	return &projectRepository{}
}

func (r *projectRepository) Create(ctx xcontext.Context, e *entity.Project) error {
	if err := ctx.DB().Model(e).Create(e).Error; err != nil {
		return err
	}

	return nil
}

func (r *projectRepository) GetList(ctx xcontext.Context, filter GetListProjectFilter) ([]entity.Project, error) {
	var result []entity.Project
	tx := ctx.DB().
		Limit(filter.Limit).
		Offset(filter.Offset)

	if filter.Q != "" {
		tx = tx.Select("*, MATCH(name,introduction) AGAINST (?) as score", filter.Q).
			Where("MATCH(name,introduction) AGAINST (?) > 0", filter.Q).
			Order("score DESC")
	}

	if filter.ReferredBy != "" {
		tx = tx.Where("referred_by=?", filter.ReferredBy)
	}

	if filter.ReferralStatus != "" {
		tx = tx.Where("referral_status=?", filter.ReferralStatus)
	}

	if err := tx.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) GetByID(ctx xcontext.Context, id string) (*entity.Project, error) {
	result := &entity.Project{}
	if err := ctx.DB().Take(result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) GetByName(ctx xcontext.Context, name string) (*entity.Project, error) {
	result := &entity.Project{}
	if err := ctx.DB().Take(result, "name=?", name).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) GetByIDs(ctx xcontext.Context, ids []string) ([]entity.Project, error) {
	result := []entity.Project{}
	tx := ctx.DB().Take(&result, "id IN (?)", ids)
	if tx.Error != nil {
		return nil, tx.Error
	}

	if len(result) != len(ids) {
		return nil, fmt.Errorf("got %d records, but expected %d", len(result), len(ids))
	}

	return result, nil
}

func (r *projectRepository) UpdateByID(ctx xcontext.Context, id string, e entity.Project) error {
	tx := ctx.DB().
		Model(&entity.Project{}).
		Where("id=?", id).
		Omit("created_by", "created_at", "id").
		Updates(e)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	return nil
}

func (r *projectRepository) UpdateReferralStatusByIDs(
	ctx xcontext.Context, ids []string, status entity.ReferralStatusType,
) error {
	tx := ctx.DB().
		Model(&entity.Project{}).
		Where("id IN (?)", ids).
		Update("referral_status", status)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return errors.New("row affected is empty")
	}

	if int(tx.RowsAffected) != len(ids) {
		return fmt.Errorf("got %d row affected, but expected %d", tx.RowsAffected, len(ids))
	}

	return nil
}

func (r *projectRepository) DeleteByID(ctx xcontext.Context, id string) error {
	tx := ctx.DB().
		Delete(&entity.Project{}, "id=?", id)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	return nil
}

func (r *projectRepository) GetFollowingList(ctx xcontext.Context, userID string, offset, limit int) ([]entity.Project, error) {
	var result []entity.Project
	if err := ctx.DB().
		Joins("join participants on projects.id = participants.project_id").
		Where("participants.user_id=?", userID).
		Limit(limit).Offset(offset).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) IncreaseFollowers(ctx xcontext.Context, projectID string) error {
	tx := ctx.DB().
		Model(&entity.Project{}).
		Where("id=?", projectID).
		Update("followers", gorm.Expr("followers+1"))

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
