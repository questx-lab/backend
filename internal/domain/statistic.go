package domain

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
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
	now := time.Now()
	var (
		val string
		ty  string
	)
	switch entity.AchievementRange(req.Range) {
	case entity.AchievementRangeWeek:
		year, week := now.ISOWeek()
		val = fmt.Sprintf(`week/%d/%d`, week, year)
	case entity.AchievementRangeMonth:
		month := now.Month()
		year := now.Year()
		val = fmt.Sprintf(`month/%d/%d`, month, year)
	case entity.AchievementRangeTotal:
	default:
		return nil, errorx.New(errorx.BadRequest, "Leader board range must be week, month, total")
	}

	switch req.Type {
	case "task":
		ty = "total_task"
	case "exp":
		ty = "exp_task"
	default:
		return nil, errorx.New(errorx.BadRequest, "Leader board type must be task or exp")
	}

	achievements, err := d.achievementRepo.GetLeaderBoard(ctx, &repository.LeaderBoardFilter{
		ProjectID: req.ProjectID,
		Range:     req.Range,
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
			UserID:    a.UserID,
			TotalTask: int64(a.TotalTask),
			TotalExp:  a.TotalExp,
		})
	}

	return &model.GetLeaderBoardResponse{
		Data: as,
	}, nil
}
