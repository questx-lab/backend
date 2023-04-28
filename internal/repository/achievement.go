package repository

import (
	"fmt"
	"time"

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
	GetPrevLeaderBoard(ctx xcontext.Context, filter LeaderBoardKey) ([]*entity.UserAggregate, error)
}

type LeaderBoardKey struct {
	ProjectID string
	Type      string
	Range     string
}
type LeaderBoardValue struct {
	Data       []*entity.UserAggregate
	Type       string
	RangeValue string
}

type achievementRepository struct {
	prevLeaderBoard map[LeaderBoardKey]LeaderBoardValue
}

func NewUserAggregateRepository() UserAggregateRepository {
	return &achievementRepository{
		prevLeaderBoard: make(map[LeaderBoardKey]LeaderBoardValue),
	}
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

func (r *achievementRepository) GetPrevLeaderBoard(ctx xcontext.Context, filter LeaderBoardKey) ([]*entity.UserAggregate, error) {
	prev, ok := r.prevLeaderBoard[filter]
	rangeValue, err := getVal(filter.Range)
	if err != nil {
		return nil, err
	}
	if !ok || prev.RangeValue != rangeValue {
		var result []*entity.UserAggregate
		tx := ctx.DB().Model(&entity.UserAggregate{}).
			Where("project_id = ? AND range_value = ?", filter.ProjectID, rangeValue).
			Order(filter.Type).
			Find(&result)
		if err := tx.Error; err != nil {
			return nil, err
		}
		r.prevLeaderBoard[filter] = LeaderBoardValue{
			Data:       result,
			Type:       filter.Type,
			RangeValue: rangeValue,
		}
		return result, nil
	}
	return prev.Data, nil

}

func getVal(typeV string) (string, error) {
	var val string
	now := time.Now()
	switch entity.UserAggregateRange(typeV) {
	case entity.UserAggregateRangeWeek:
		year, week := now.AddDate(0, 0, -7).ISOWeek()
		val = fmt.Sprintf(`week/%d/%d`, week, year)
	case entity.UserAggregateRangeMonth:
		month := now.AddDate(0, -1, 0).Month()
		year := now.Year()
		val = fmt.Sprintf(`month/%d/%d`, month, year)
	case entity.UserAggregateRangeTotal:
		val = "total"
	default:
		return "", fmt.Errorf("leader board range must be week, month, total. but got %s", typeV)
	}
	return val, nil
}
