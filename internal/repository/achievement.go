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
	Range     string
	Value     string

	Type string

	Offset int
	Limit  int
}

type AchievementRepository interface {
	UpsertPoint(xcontext.Context, *entity.Achievement) error
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

func (r *achievementRepository) UpsertPoint(ctx xcontext.Context, e *entity.Achievement) error {
	tx := ctx.DB().Model(&entity.Achievement{}).
		Where(
			`project_id = ? 
			AND user_id = ? 
			AND range = ?
			AND value = ?
			`,
			e.ProjectID, e.UserID, e.Range, e.Value).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "project_id"},
				{Name: "user_id"},
				{Name: "range"},
				{Name: "value"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_task": gorm.Expr("total_task + ?", e.TotalTask),
				"total_exp":  gorm.Expr("total_exp + ?", e.TotalExp),
			}),
		}).
		Create(e)
	if err := tx.Error; err != nil {
		return err
	}
	if tx.RowsAffected != 1 {
		return fmt.Errorf("update status not exec correctly")
	}
	return nil
}

func (r *achievementRepository) GetLeaderBoard(ctx xcontext.Context, filter *LeaderBoardFilter) ([]*entity.Achievement, error) {
	var result []*entity.Achievement
	tx := ctx.DB().Model(&entity.Achievement{}).
		Where(`project_id = ? 
	AND range = ? 
	AND value = ?
	`, filter.ProjectID, filter.Range, filter.Value).
		Limit(filter.Limit).
		Offset(filter.Offset).
		Order(filter.Type).
		Find(&result)
	if err := tx.Error; err != nil {
		return nil, err
	}

	return result, nil
}
