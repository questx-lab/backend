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

const SharpScoutBadgeName = "sharp_scout"

// sharpScoutBadgeScanner scans badge level based on the number of successful
// invitation of a user.
type sharpScoutBadgeScanner struct {
	badgeRepo    repository.BadgeRepository
	followerRepo repository.FollowerRepository
}

func NewSharpScoutBadgeScanner(
	badgeRepo repository.BadgeRepository,
	followerRepo repository.FollowerRepository,
) *sharpScoutBadgeScanner {
	return &sharpScoutBadgeScanner{
		badgeRepo:    badgeRepo,
		followerRepo: followerRepo,
	}
}

func (sharpScoutBadgeScanner) Name() string {
	return SharpScoutBadgeName
}

func (sharpScoutBadgeScanner) IsGlobal() bool {
	return false
}

func (s *sharpScoutBadgeScanner) Scan(ctx context.Context, userID, communityID string) ([]entity.Badge, error) {
	follower, err := s.followerRepo.Get(ctx, userID, communityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.Unavailable, "User has not followed the community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get follower: %v", err)
		return nil, errorx.Unknown
	}

	suitableBadges, err := s.badgeRepo.GetLessThanValue(ctx, s.Name(), int(follower.InviteCount))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get the suitable badge of %s: %v", s.Name(), err)
		return nil, errorx.Unknown
	}

	return suitableBadges, nil
}
