package badge

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

const SharpScoutBadgeName = "sharp scout"

// sharpScoutBadgeScanner scans badge level based on the number of successful
// invitation of a user.
type sharpScoutBadgeScanner struct {
	levelConfig  []uint64
	followerRepo repository.FollowerRepository
}

func NewSharpScoutBadgeScanner(
	followerRepo repository.FollowerRepository,
	levelConfig []uint64,
) *sharpScoutBadgeScanner {
	return &sharpScoutBadgeScanner{levelConfig: levelConfig, followerRepo: followerRepo}
}

func (sharpScoutBadgeScanner) Name() string {
	return SharpScoutBadgeName
}

func (sharpScoutBadgeScanner) IsGlobal() bool {
	return false
}

func (s *sharpScoutBadgeScanner) Scan(ctx context.Context, userID, communityID string) (int, error) {
	follower, err := s.followerRepo.Get(ctx, userID, communityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errorx.New(errorx.Unavailable, "User has not followed the community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get follower: %v", err)
		return 0, errorx.Unknown
	}

	finalLevel := 0
	for level, value := range s.levelConfig {
		if follower.InviteCount < value {
			break
		}
		finalLevel = level + 1
	}

	return finalLevel, nil
}
