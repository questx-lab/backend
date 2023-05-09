package badge

import (
	"errors"

	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

const SharpScoutBadgeName = "sharp scout"

type sharpScoutBadge struct {
	levelConfig     []uint64
	participantRepo repository.ParticipantRepository
}

func NewSharpScoutBadge(
	participantRepo repository.ParticipantRepository,
	levelConfig []uint64,
) *sharpScoutBadge {
	return &sharpScoutBadge{levelConfig: levelConfig, participantRepo: participantRepo}
}

func (b *sharpScoutBadge) Name() string {
	return SharpScoutBadgeName
}

func (b *sharpScoutBadge) IsGlobal() bool {
	return false
}

func (b *sharpScoutBadge) Scan(ctx xcontext.Context, userID, projectID string) (int, error) {
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
		finalLevel = level
	}

	return finalLevel, nil
}
