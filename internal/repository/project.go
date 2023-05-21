package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type GetListProjectFilter struct {
	Q               string
	ReferredBy      string
	ReferralStatus  entity.ReferralStatusType
	Offset          int
	Limit           int
	OrderByTrending bool
}

type ProjectRepository interface {
	Create(ctx context.Context, e *entity.Project) error
	GetList(ctx context.Context, filter GetListProjectFilter) ([]entity.Project, error)
	GetByID(ctx context.Context, id string) (*entity.Project, error)
	GetByName(ctx context.Context, name string) (*entity.Project, error)
	UpdateByID(ctx context.Context, id string, e entity.Project) error
	GetByIDs(ctx context.Context, ids []string) ([]entity.Project, error)
	UpdateReferralStatusByIDs(ctx context.Context, ids []string, status entity.ReferralStatusType) error
	DeleteByID(ctx context.Context, id string) error
	GetFollowingList(ctx context.Context, userID string, offset, limit int) ([]entity.Project, error)
	IncreaseFollowers(ctx context.Context, projectID string) error
	UpdateTrendingScore(ctx context.Context, projectID string, score int) error
}

type projectRepository struct{}

func NewProjectRepository() ProjectRepository {
	return &projectRepository{}
}

func (r *projectRepository) Create(ctx context.Context, e *entity.Project) error {
	if err := xcontext.DB(ctx).Model(e).Create(e).Error; err != nil {
		return err
	}

	return nil
}

func (r *projectRepository) GetList(ctx context.Context, filter GetListProjectFilter) ([]entity.Project, error) {
	var result []entity.Project
	tx := xcontext.DB(ctx).
		Limit(filter.Limit).
		Offset(filter.Offset)

	if filter.OrderByTrending {
		tx = tx.Order("trending_score DESC")
	}

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

func (r *projectRepository) GetByID(ctx context.Context, id string) (*entity.Project, error) {
	result := &entity.Project{}
	if err := xcontext.DB(ctx).Take(result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) GetByName(ctx context.Context, name string) (*entity.Project, error) {
	result := &entity.Project{}
	if err := xcontext.DB(ctx).Take(result, "name=?", name).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.Project, error) {
	result := []entity.Project{}
	tx := xcontext.DB(ctx).Take(&result, "id IN (?)", ids)
	if tx.Error != nil {
		return nil, tx.Error
	}

	if len(result) != len(ids) {
		return nil, fmt.Errorf("got %d records, but expected %d", len(result), len(ids))
	}

	return result, nil
}

func (r *projectRepository) UpdateByID(ctx context.Context, id string, e entity.Project) error {
	tx := xcontext.DB(ctx).
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
	ctx context.Context, ids []string, status entity.ReferralStatusType,
) error {
	tx := xcontext.DB(ctx).
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

func (r *projectRepository) DeleteByID(ctx context.Context, id string) error {
	tx := xcontext.DB(ctx).
		Delete(&entity.Project{}, "id=?", id)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	return nil
}

func (r *projectRepository) GetFollowingList(ctx context.Context, userID string, offset, limit int) ([]entity.Project, error) {
	var result []entity.Project
	if err := xcontext.DB(ctx).
		Joins("join participants on projects.id = participants.project_id").
		Where("participants.user_id=?", userID).
		Limit(limit).Offset(offset).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) IncreaseFollowers(ctx context.Context, projectID string) error {
	tx := xcontext.DB(ctx).
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

func (r *projectRepository) UpdateTrendingScore(ctx context.Context, projectID string, score int) error {
	return xcontext.DB(ctx).
		Model(&entity.Project{}).
		Where("id=?", projectID).
		Update("trending_score", score).Error
}
