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
	levelConfig     []uint64
	participantRepo repository.ParticipantRepository
}

func NewRainBowBadgeScanner(
	participantRepo repository.ParticipantRepository,
	levelConfig []uint64,
) *rainbowBadgeScanner {
	return &rainbowBadgeScanner{
		levelConfig:     levelConfig,
		participantRepo: participantRepo,
	}
}

func (rainbowBadgeScanner) Name() string {
	return RainBowBadgeName
}

func (rainbowBadgeScanner) IsGlobal() bool {
	return false
}

func (s *rainbowBadgeScanner) Scan(ctx context.Context, userID, projectID string) (int, error) {
	participant, err := s.participantRepo.Get(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errorx.New(errorx.Unavailable, "User has not followed the project")
		}

		xcontext.Logger(ctx).Errorf("Cannot get participant: %v", err)
		return 0, errorx.Unknown
	}

	finalLevel := 0
	for level, value := range s.levelConfig {
		if participant.Streak < value {
			break
		}
		finalLevel = level + 1
	}

	return finalLevel, nil
}
