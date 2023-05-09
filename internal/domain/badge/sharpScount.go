package badge

import (
	"errors"

	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

const SharpScoutBadgeName = "sharp scout"

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

func (b *sharpScoutBadgeScanner) Name() string {
	return SharpScoutBadgeName
}

func (b *sharpScoutBadgeScanner) IsGlobal() bool {
	return false
}

func (b *sharpScoutBadgeScanner) Scan(ctx xcontext.Context, userID, projectID string) (int, error) {
	participant, err := b.participantRepo.Get(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errorx.New(errorx.Unavailable, "User has not followed the project")
		}

		ctx.Logger().Errorf("Cannot get participant: %v", err)
		return 0, errorx.Unknown
	}

	finalLevel := 0
	for level, value := range b.levelConfig {
		if participant.InviteCount < value {
			break
		}
		finalLevel = level + 1
	}

	return finalLevel, nil
}
