package testutil

import (
	"context"
	"fmt"
	"sync"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type TestDatabaseDomain interface {
	TestDatabaseMaximumHit(ctx context.Context, req *TestDatabaseMaximumHitRequest) (*TestDatabaseMaximumHitResponse, error)
}

func NewTestDatabaseDomain(
	claimedQuestRepo repository.ClaimedQuestRepository,
	communityRepo repository.CommunityRepository,
	userRepo repository.UserRepository,
) TestDatabaseDomain {
	return &testDatabaseDomain{
		claimedQuestRepo:   claimedQuestRepo,
		communityRepo:      communityRepo,
		globalRoleVerifier: common.NewGlobalRoleVerifier(userRepo),
		userRepo:           userRepo,
	}
}

type testDatabaseDomain struct {
	claimedQuestRepo   repository.ClaimedQuestRepository
	communityRepo      repository.CommunityRepository
	userRepo           repository.UserRepository
	globalRoleVerifier *common.GlobalRoleVerifier
}

type TestDatabaseMaximumHitRequest struct {
	BunchHit int
}
type TestDatabaseMaximumHitResponse struct{}

func (d *testDatabaseDomain) TestDatabaseMaximumHit(ctx context.Context, req *TestDatabaseMaximumHitRequest) (*TestDatabaseMaximumHitResponse, error) {
	if err := d.globalRoleVerifier.Verify(ctx, entity.GlobalAdminRoles...); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	var wg sync.WaitGroup
	period, err := statistic.ToPeriod("month")
	if err != nil {
		return nil, err
	}
	communities, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{})
	if err != nil {
		return nil, err
	}
	if len(communities) == 0 {
		if err != nil {
			return nil, fmt.Errorf("no community")
		}
	}
	for i := 1; i <= req.BunchHit; i++ {
		wg.Add(1)
		_, err := d.claimedQuestRepo.Statistic(
			ctx,
			repository.StatisticClaimedQuestFilter{
				CommunityID:   communities[0].ID,
				ReviewedStart: period.Start(),
				ReviewedEnd:   period.End(),
				Status:        []entity.ClaimedQuestStatus{entity.Accepted, entity.AutoAccepted},
			},
		)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot load statistic from database: %v", err)
			return nil, errorx.Unknown
		}
	}

	return &TestDatabaseMaximumHitResponse{}, nil
}
