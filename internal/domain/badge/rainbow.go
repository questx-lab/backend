package badge

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

const RainBowBadgeName = "rainbow"

// rainbowBadgeScanner scans badge level based on the number of continuous days
// which user claimed quest.
type rainbowBadgeScanner struct {
	badgeRepo    repository.BadgeRepository
	followerRepo repository.FollowerRepository
}

func NewRainBowBadgeScanner(
	badgeRepo repository.BadgeRepository,
	followerRepo repository.FollowerRepository,
) *rainbowBadgeScanner {
	return &rainbowBadgeScanner{
		badgeRepo:    badgeRepo,
		followerRepo: followerRepo,
	}
}

func (rainbowBadgeScanner) Name() string {
	return RainBowBadgeName
}

func (rainbowBadgeScanner) IsGlobal() bool {
	return false
}

func (s *rainbowBadgeScanner) Scan(ctx context.Context, userID, communityID string) ([]entity.Badge, error) {
	follower, err := s.followerRepo.Get(ctx, userID, communityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.Unavailable, "User has not followed the community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get follower: %v", err)
		return nil, errorx.Unknown
	}

	suitableBadges, err := s.badgeRepo.GetLessThanValue(ctx, s.Name(), int(follower.Streaks))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get the suitable badge of %s: %v", s.Name(), err)
		return nil, errorx.Unknown
	}

	return suitableBadges, nil
}
