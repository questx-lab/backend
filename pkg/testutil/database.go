package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
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
	categoryRepository repository.CategoryRepository,
) TestDatabaseDomain {
	return &testDatabaseDomain{
		claimedQuestRepo:   claimedQuestRepo,
		communityRepo:      communityRepo,
		globalRoleVerifier: common.NewGlobalRoleVerifier(userRepo),
		userRepo:           userRepo,
		categoryRepository: categoryRepository,
	}
}

type testDatabaseDomain struct {
	claimedQuestRepo   repository.ClaimedQuestRepository
	communityRepo      repository.CommunityRepository
	userRepo           repository.UserRepository
	globalRoleVerifier *common.GlobalRoleVerifier
	categoryRepository repository.CategoryRepository
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
	communityID := communities[0].ID
	startTime := time.Now()

	insertBunchHit := req.BunchHit * 5 / 100 // insert is 5%

	readHitSuccess := int64(0)
	writeHitSuccess := int64(0)
	xcontext.Logger(ctx).Errorf("Start test database with read_hit = %v, write_hit = %v", req.BunchHit-insertBunchHit, insertBunchHit)

	for i := 1; i <= insertBunchHit; i++ {
		eg.Go(func() error {
			id := uuid.NewString()
			if err := d.categoryRepository.Create(ctx, &entity.Category{
				Base: entity.Base{
					ID: id,
				},
				Name: fmt.Sprintf("test-%s", id),
				CommunityID: sql.NullString{
					String: communityID,
					Valid:  true,
				},
				CreatedBy: xcontext.RequestUserID(ctx),
			}); err != nil {
				return err
			} else {
				atomic.AddInt64(&writeHitSuccess, 1)
			}

			return nil
		})
	}

	for i := 1; i <= req.BunchHit-insertBunchHit; i++ {

		eg.Go(func() error {
			var err error
			if req.IsStrongQuery {
				_, err = d.claimedQuestRepo.Statistic(
					ctx,
					repository.StatisticClaimedQuestFilter{
						CommunityID:   communityID,
						ReviewedStart: period.Start(),
						ReviewedEnd:   period.End(),
						Status:        []entity.ClaimedQuestStatus{entity.Accepted, entity.AutoAccepted},
					},
				)
			} else {
				_, err = d.communityRepo.GetByID(
					ctx,
					communityID,
				)
			}
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot load statistic from database: %v", err)
				return err
			} else {
				atomic.AddInt64(&readHitSuccess, 1)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		xcontext.Logger(ctx).Errorf("Read hit success = %v, Write hit success = %v", readHitSuccess, writeHitSuccess)
		return nil, err
	}

	xcontext.Logger(ctx).Errorf("Test database took: %v seconds", time.Since(startTime).Seconds())

	return &TestDatabaseMaximumHitResponse{}, nil
}
