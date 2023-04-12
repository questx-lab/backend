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
	achievementRepo repository.AchievementRepository
}

func NewStatisticDomain(achievementRepo repository.AchievementRepository) StatisticDomain {
	return &statisticDomain{
		achievementRepo: achievementRepo,
	}
}

func (d *statisticDomain) GetLeaderBoard(ctx xcontext.Context, req *model.GetLeaderBoardRequest) (*model.GetLeaderBoardResponse, error) {
	var (
		ty string
	)
	val, err := dateutil.GetCurrentValueByRange(entity.AchievementRange(req.Range))
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
		ProjectID: req.ProjectID,
		Value:     val,
		Type:      ty,

		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to get leader board")
	}

	var as []model.Achievement

	for _, a := range achievements {
		as = append(as, model.Achievement{
			UserID:     a.UserID,
			TotalTask:  a.TotalTask,
			TotalPoint: a.TotalPoint,
		})
	}

	return &model.GetLeaderBoardResponse{
		Data: as,
	}, nil
}
