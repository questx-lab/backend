package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LeaderBoardFilter struct {
	ProjectID  string
	RangeValue string

	Type string

	Offset int
	Limit  int
}

type UserAggregateRepository interface {
	Upsert(xcontext.Context, *entity.UserAggregate) error
	GetLeaderBoard(xcontext.Context, *LeaderBoardFilter) ([]*entity.UserAggregate, error)
}

type achievementRepository struct{}

func NewUserAggregateRepository() UserAggregateRepository {
	return &achievementRepository{}
}

func (r *achievementRepository) BulkInsert(ctx xcontext.Context, e []*entity.UserAggregate) error {
	tx := ctx.DB().Create(e)
	if err := tx.Error; err != nil {
		return err
	}
	return nil
}

func (r *achievementRepository) Upsert(ctx xcontext.Context, e *entity.UserAggregate) error {
	return ctx.DB().Model(&entity.UserAggregate{}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "project_id"},
				{Name: "user_id"},
				{Name: "range_value"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_task":  gorm.Expr("total_task + ?", e.TotalTask),
				"total_point": gorm.Expr("total_point + ?", e.TotalPoint),
			}),
		}).
		Create(e).Error
}

func (r *achievementRepository) GetLeaderBoard(ctx xcontext.Context, filter *LeaderBoardFilter) ([]*entity.UserAggregate, error) {
	var result []*entity.UserAggregate
	tx := ctx.DB().Model(&entity.UserAggregate{}).
		Where("project_id = ? AND range_value = ?", filter.ProjectID, filter.RangeValue).
		Limit(filter.Limit).
		Offset(filter.Offset).
		Order(filter.Type).
		Find(&result)
	if err := tx.Error; err != nil {
		return nil, err
	}

	return result, nil
}
