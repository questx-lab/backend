package badge

import (
	"context"
	"database/sql"
	"errors"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type Manager struct {
	// This field is only written at initialization. After that, it is readonly.
	// So no need to use sync map here.
	badgeScanners map[string]BadgeScanner
	badgeRepo     repository.BadgeRepo
}

func NewManager(badgeRepo repository.BadgeRepo, badgeScanners ...BadgeScanner) *Manager {
	manager := &Manager{
		badgeRepo:     badgeRepo,
		badgeScanners: make(map[string]BadgeScanner),
	}

	for _, b := range badgeScanners {
		manager.badgeScanners[b.Name()] = b
	}

	return manager
}

func (m *Manager) WithBadges(badgeNames ...string) *contextManager {
	return &contextManager{
		manager:    m,
		badgeNames: badgeNames,
	}
}

type contextManager struct {
	manager    *Manager
	badgeNames []string
}

func (c *contextManager) ScanAndGive(ctx context.Context, userID, projectID string) error {
	for _, badgeName := range c.badgeNames {
		badgeScanner, ok := c.manager.badgeScanners[badgeName]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found badge name %s", badgeName)
			return errorx.Unknown
		}

		level, err := badgeScanner.Scan(ctx, userID, projectID)
		if err != nil {
			return err
		}

		// No need to update a badge with no level.
		if level == 0 {
			continue
		}

		actualProjectID := sql.NullString{Valid: true, String: projectID}
		if badgeScanner.IsGlobal() {
			actualProjectID = sql.NullString{Valid: false}
		}

		newBadge := &entity.Badge{
			UserID:      userID,
			ProjectID:   actualProjectID,
			Name:        badgeScanner.Name(),
			Level:       0,
			WasNotified: false,
		}

		currentBadge, err := c.manager.badgeRepo.Get(ctx, userID, projectID, badgeScanner.Name())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newBadge.Level = level
			} else {
				xcontext.Logger(ctx).Errorf("Cannot get the current badge: %v", err)
				return errorx.Unknown
			}
		} else if currentBadge.Level < level {
			newBadge.Level = level
		}

		if newBadge.Level > 0 {
			if err := c.manager.badgeRepo.Upsert(ctx, newBadge); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot update or create badge: %v", err)
				return errorx.Unknown
			}
		}
	}

	return nil
}
