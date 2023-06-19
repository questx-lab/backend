package badge

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
)

type BadgeScanner interface {
	// Name returns the name of badge.
	Name() string

	// IsGlobal returns true if this badge isn't relevant to community, otherwise,
	// returns false if it is community-specific.
	IsGlobal() bool

	// Scan detects the badge should be given to user or not. It returns a list
	// of badges ordered by descending of level, which the last badge in list
	// is the largest level that user receives.
	Scan(ctx context.Context, userID, communityID string) ([]entity.Badge, error)
}
