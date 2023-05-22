package badge

import "context"

type BadgeScanner interface {
	// Name returns the name of badge.
	Name() string

	// IsGlobal returns true if this badge isn't relevant to community, otherwise,
	// returns false if it is community-specific.
	IsGlobal() bool

	// Scan detects the badge should be given to user or not. It returns badge's
	// level which user will be awarded.
	Scan(ctx context.Context, userID, communityID string) (int, error)
}
