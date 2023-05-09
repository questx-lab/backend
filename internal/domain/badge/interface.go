package badge

import (
	"github.com/questx-lab/backend/pkg/xcontext"
)

type BadgeScanner interface {
	// Name returns the name of badge.
	Name() string

	// IsGlobal returns true if this badge isn't relevant to project, otherwise,
	// returns false if it is project-specific.
	IsGlobal() bool

	// Scan detects the badge should be given to user or not. It returns badge's
	// level which user will be awarded.
	Scan(ctx xcontext.Context, userID, projectID string) (int, error)
}
