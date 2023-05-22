package repository

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/dateutil"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/puzpuzpuz/xsync"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LeaderBoardFilter struct {
	CommunityID string
	RangeValue  string

	OrderField string

	Offset int
	Limit  int
}

type LeaderBoardKey struct {
	CommunityID string
	OrderField  string
	Range       entity.UserAggregateRange
}

func (k *LeaderBoardKey) GetKey() string {
	return fmt.Sprintf("%s|%s|%s", k.CommunityID, k.OrderField, k.Range)
}

type LeaderBoardValue struct {
	Data       []entity.UserAggregate
	OrderField string
	RangeValue string
}

type UserAggregateRepository interface {
	Upsert(context.Context, *entity.UserAggregate) error
	GetTotal(ctx context.Context, userID, communityID string) (*entity.UserAggregate, error)
	GetLeaderBoard(context.Context, *LeaderBoardFilter) ([]entity.UserAggregate, error)
	GetPrevLeaderBoard(ctx context.Context, filter LeaderBoardKey) ([]entity.UserAggregate, error)
}

type achievementRepository struct {
	prevLeaderBoard *xsync.MapOf[string, LeaderBoardValue]
}

func NewUserAggregateRepository() UserAggregateRepository {
	return &achievementRepository{
		prevLeaderBoard: xsync.NewMapOf[LeaderBoardValue](),
	}
}

func (r *achievementRepository) BulkInsert(ctx context.Context, e []*entity.UserAggregate) error {
	tx := xcontext.DB(ctx).Create(e)
	if err := tx.Error; err != nil {
		return err
	}
	return nil
}

func (r *achievementRepository) GetTotal(ctx context.Context, userID, communityID string) (*entity.UserAggregate, error) {
	var result entity.UserAggregate
	tx := xcontext.DB(ctx).Model(&entity.UserAggregate{}).
		Where("user_id=? AND community_id=? AND `range`=?", userID, communityID, entity.UserAggregateRangeTotal).
		Take(&result)
	if err := tx.Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *achievementRepository) Upsert(ctx context.Context, e *entity.UserAggregate) error {
	return xcontext.DB(ctx).Model(&entity.UserAggregate{}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "community_id"},
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

func (r *achievementRepository) GetLeaderBoard(ctx context.Context, filter *LeaderBoardFilter) ([]entity.UserAggregate, error) {
	var result []entity.UserAggregate
	tx := xcontext.DB(ctx).Model(&entity.UserAggregate{}).
		Where("community_id=? AND range_value=?", filter.CommunityID, filter.RangeValue).
		Limit(filter.Limit).
		Offset(filter.Offset).
		Order(filter.OrderField + " DESC").
		Find(&result)
	if err := tx.Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *achievementRepository) GetPrevLeaderBoard(ctx context.Context, filter LeaderBoardKey) ([]entity.UserAggregate, error) {
	prev, ok := r.prevLeaderBoard.Load(filter.GetKey())
	rangeValue, err := dateutil.GetPreviousValueByRange(filter.Range)
	if err != nil {
		return nil, err
	}

	if !ok || prev.RangeValue != rangeValue {
		var result []entity.UserAggregate
		tx := xcontext.DB(ctx).Model(&entity.UserAggregate{}).
			Where("community_id=? AND range_value=?", filter.CommunityID, rangeValue).
			Order(filter.OrderField + " DESC").
			Find(&result)
		if err := tx.Error; err != nil {
			return nil, err
		}

		r.prevLeaderBoard.Store(filter.GetKey(), LeaderBoardValue{
			Data:       result,
			OrderField: filter.OrderField,
			RangeValue: rangeValue,
		})

		return result, nil
	}

	return prev.Data, nil
}
