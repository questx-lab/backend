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

type AchievementRepository interface {
	BulkUpsertPoint(xcontext.Context, []*entity.Achievement) error
	GetLeaderBoard(xcontext.Context, *LeaderBoardFilter) ([]*entity.Achievement, error)
}

type achievementRepository struct{}

func NewAchievementRepository() AchievementRepository {
	return &achievementRepository{}
}

func (r *achievementRepository) BulkInsert(ctx xcontext.Context, e []*entity.Achievement) error {
	tx := ctx.DB().Create(e)
	if err := tx.Error; err != nil {
		return err
	}
	return nil
}

func (r *achievementRepository) BulkUpsertPoint(ctx xcontext.Context, es []*entity.Achievement) error {
	tx := ctx.DB().Model(&entity.Achievement{}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "project_id"},
				{Name: "user_id"},
				{Name: "value"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_task": gorm.Expr("total_task + EXCLUDED.total_task"),
				"total_exp":  gorm.Expr("total_exp + EXCLUDED.total_exp"),
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

func (r *achievementRepository) GetLeaderBoard(ctx xcontext.Context, filter *LeaderBoardFilter) ([]*entity.Achievement, error) {
	var result []*entity.Achievement
	tx := ctx.DB().Model(&entity.Achievement{}).
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
