package badge

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

const RainBowBadgeName = "rainbow"

// rainbowBadgeScanner scans badge level based on the number of continuous days
// which user claimed quest.
type rainbowBadgeScanner struct {
	levelConfig  []uint64
	followerRepo repository.FollowerRepository
}

func NewRainBowBadgeScanner(
	followerRepo repository.FollowerRepository,
	levelConfig []uint64,
) *rainbowBadgeScanner {
	return &rainbowBadgeScanner{
		levelConfig:  levelConfig,
		followerRepo: followerRepo,
	}
}

func (rainbowBadgeScanner) Name() string {
	return RainBowBadgeName
}

func (rainbowBadgeScanner) IsGlobal() bool {
	return false
}

func (s *rainbowBadgeScanner) Scan(ctx context.Context, userID, communityID string) (int, error) {
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
		if follower.Streak < value {
			break
		}
		finalLevel = level + 1
	}

	return finalLevel, nil
}
