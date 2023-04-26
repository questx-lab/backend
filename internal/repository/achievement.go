package repository

import (
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LeaderBoardFilter struct {
	ProjectID string
	Value     string

	Type string

	Offset int
	Limit  int
}

type UserAggregateRepository interface {
	BulkUpsertPoint(xcontext.Context, []*entity.UserAggregate) error
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

func (r *achievementRepository) BulkUpsertPoint(ctx xcontext.Context, es []*entity.UserAggregate) error {
	tx := ctx.DB().Model(&entity.UserAggregate{}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "project_id"},
				{Name: "user_id"},
				{Name: "value"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				// "total_task":  gorm.Expr("total_task + EXCLUDED.total_task"),
				// "total_point": gorm.Expr("total_point + EXCLUDED.total_point"),
				"total_task":  gorm.Expr("total_task + EXCLUDED.total_task"),
				"total_point": gorm.Expr("total_point + EXCLUDED.total_point"),
			}),
		}).
		Create(es)
	if err := tx.Error; err != nil {
		return err
	}
	if tx.RowsAffected != int64(len(es)) {
		return fmt.Errorf("update status not exec correctly")
	}
	return nil
}

func (r *achievementRepository) GetLeaderBoard(ctx xcontext.Context, filter *LeaderBoardFilter) ([]*entity.UserAggregate, error) {
	var result []*entity.UserAggregate
	tx := ctx.DB().Model(&entity.UserAggregate{}).
		Where(`project_id = ? 
	AND value = ?
	`, filter.ProjectID, filter.Value).
		Limit(filter.Limit).
		Offset(filter.Offset).
		Order(filter.Type).
		Find(&result)
	if err := tx.Error; err != nil {
		return nil, err
	}

	return result, nil
}
