package testutil

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"

	"golang.org/x/sync/errgroup"
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
	BunchHit      int  `json:"bunch_hit"`
	IsStrongQuery bool `json:"is_strong_query"`
}
type TestDatabaseMaximumHitResponse struct{}

func (d *testDatabaseDomain) TestDatabaseMaximumHit(ctx context.Context, req *TestDatabaseMaximumHitRequest) (*TestDatabaseMaximumHitResponse, error) {
	if err := d.globalRoleVerifier.Verify(ctx, entity.GlobalAdminRoles...); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	eg, _ := errgroup.WithContext(ctx)
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
	startTime := time.Now()
	count := int64(0)
	xcontext.Logger(ctx).Errorf("Start test database with bunch_hit: %v", req.BunchHit)
	for i := 1; i <= req.BunchHit; i++ {

		eg.Go(func() error {
			var err error
			if req.IsStrongQuery {
				_, err = d.claimedQuestRepo.Statistic(
					ctx,
					repository.StatisticClaimedQuestFilter{
						CommunityID:   communities[0].ID,
						ReviewedStart: period.Start(),
						ReviewedEnd:   period.End(),
						Status:        []entity.ClaimedQuestStatus{entity.Accepted, entity.AutoAccepted},
					},
				)
			} else {
				_, err = d.communityRepo.GetByID(
					ctx,
					communities[0].ID,
				)
			}
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot load statistic from database: %v", err)
				return err
			} else {
				atomic.AddInt64(&count, 1)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		xcontext.Logger(ctx).Errorf("Success transaction amount got %d txs", count)
		return nil, err
	}

	xcontext.Logger(ctx).Errorf("Test database took: %v seconds", time.Since(startTime).Seconds())

	return &TestDatabaseMaximumHitResponse{}, nil
}
