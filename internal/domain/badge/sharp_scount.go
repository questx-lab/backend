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
	levelConfig     []uint64
	participantRepo repository.ParticipantRepository
}

func NewSharpScoutBadgeScanner(
	participantRepo repository.ParticipantRepository,
	levelConfig []uint64,
) *sharpScoutBadgeScanner {
	return &sharpScoutBadgeScanner{levelConfig: levelConfig, participantRepo: participantRepo}
}

func (sharpScoutBadgeScanner) Name() string {
	return SharpScoutBadgeName
}

func (sharpScoutBadgeScanner) IsGlobal() bool {
	return false
}

func (s *sharpScoutBadgeScanner) Scan(ctx context.Context, userID, projectID string) (int, error) {
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
		if participant.InviteCount < value {
			break
		}
		finalLevel = level + 1
	}

	return finalLevel, nil
}
