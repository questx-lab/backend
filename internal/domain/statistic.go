package domain

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/dateutil"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type StatisticDomain interface {
	GetLeaderBoard(context.Context, *model.GetLeaderBoardRequest) (*model.GetLeaderBoardResponse, error)
}

type statisticDomain struct {
	achievementRepo repository.UserAggregateRepository
	userRepo        repository.UserRepository
}

func NewStatisticDomain(
	achievementRepo repository.UserAggregateRepository,
	userRepo repository.UserRepository,
) StatisticDomain {
	return &statisticDomain{
		achievementRepo: achievementRepo,
		userRepo:        userRepo,
	}
}

func (d *statisticDomain) GetLeaderBoard(ctx context.Context, req *model.GetLeaderBoardRequest) (*model.GetLeaderBoardResponse, error) {
	val, err := dateutil.GetCurrentValueByRange(entity.UserAggregateRange(req.Range))
	if err != nil {
		return nil, errorx.New(errorx.BadRequest, err.Error())
	}

	var orderField string
	switch req.Type {
	case "task":
		orderField = "total_task"
	case "point":
		orderField = "total_point"
	default:
		return nil, errorx.New(errorx.BadRequest, "Leader board type must be task or point")
	}

	enumRange, err := enum.ToEnum[entity.UserAggregateRange](req.Range)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid range: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid range: %v", req.Range)
	}

	achievements, err := d.achievementRepo.GetLeaderBoard(ctx, &repository.LeaderBoardFilter{
		CommunityID: req.CommunityID,
		RangeValue:  val,
		OrderField:  orderField,
		Offset:      req.Offset,
		Limit:       req.Limit,
	})
	if err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to get leader board")
	}

	var userIDs []string
	for _, a := range achievements {
		userIDs = append(userIDs, a.UserID)
	}

	users, err := d.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get user list in leaderboard: %v", err)
		return nil, errorx.Unknown
	}

	userMap := map[string]entity.User{}
	for _, u := range users {
		userMap[u.ID] = u
	}

	prevAchievements, err := d.achievementRepo.GetPrevLeaderBoard(ctx, repository.LeaderBoardKey{
		CommunityID: req.CommunityID,
		OrderField:  orderField,
		Range:       enumRange,
	})
	if err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to get previous leader board")
	}

	prevRankMap := make(map[string]uint64)
	for i, a := range prevAchievements {
		prevRankMap[a.UserID] = uint64(i) + 1
	}

	data := []model.UserAggregate{}
	for i, a := range achievements {
		prevRank, ok := prevRankMap[a.UserID]
		if !ok {
			prevRank = 0
		}

		user, ok := userMap[a.UserID]
		if !ok {
			return nil, errorx.Unknown
		}

		data = append(data, model.UserAggregate{
			UserID: a.UserID,
			User: model.User{
				ID:      user.ID,
				Address: user.Address.String,
				Name:    user.Name,
				Role:    string(user.Role),
			},
			TotalTask:   a.TotalTask,
			TotalPoint:  a.TotalPoint,
			PrevRank:    prevRank,
			CurrentRank: uint64(req.Offset + i + 1),
		})
	}

	return &model.GetLeaderBoardResponse{LeaderBoard: data}, nil
}
