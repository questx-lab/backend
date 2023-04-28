package domain

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/dateutil"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type StatisticDomain interface {
	GetLeaderBoard(xcontext.Context, *model.GetLeaderBoardRequest) (*model.GetLeaderBoardResponse, error)
}

type statisticDomain struct {
	achievementRepo repository.UserAggregateRepository
}

func NewStatisticDomain(achievementRepo repository.UserAggregateRepository) StatisticDomain {
	return &statisticDomain{
		achievementRepo: achievementRepo,
	}
}

func (d *statisticDomain) GetLeaderBoard(ctx xcontext.Context, req *model.GetLeaderBoardRequest) (*model.GetLeaderBoardResponse, error) {
	var ty string
	val, err := dateutil.GetCurrentValueByRange(entity.UserAggregateRange(req.Range))
	if err != nil {
		return nil, errorx.New(errorx.BadRequest, err.Error())
	}

	switch req.Type {
	case "task":
		ty = "total_task"
	case "point":
		ty = "total_point"
	default:
		return nil, errorx.New(errorx.BadRequest, "Leader board type must be task or point")
	}

	achievements, err := d.achievementRepo.GetLeaderBoard(ctx, &repository.LeaderBoardFilter{
		ProjectID:  req.ProjectID,
		RangeValue: val,
		Type:       ty,

		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to get leader board")
	}

	prevAchievements, err := d.achievementRepo.GetPrevLeaderBoard(ctx, repository.LeaderBoardKey{
		ProjectID: req.ProjectID,
		Type:      ty,
	})

	if err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to get prev leader board")
	}

	m := make(map[string]uint64)

	for i, a := range prevAchievements {
		m[a.UserID] = uint64(i)
	}

	var as []model.UserAggregate

	for _, a := range achievements {
		rank, ok := m[a.UserID]
		if !ok {
			rank = 0
		}
		as = append(as, model.UserAggregate{
			UserID:      a.UserID,
			TotalTask:   a.TotalTask,
			TotalPoint:  a.TotalPoint,
			PrevRank:    rank,
			CurrentRank: uint64(req.Offset),
		})
	}

	return &model.GetLeaderBoardResponse{
		Data: as,
	}, nil
}
