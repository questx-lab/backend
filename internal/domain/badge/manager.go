package badge

import (
	"context"
	"database/sql"
	"errors"

	"github.com/questx-lab/backend/internal/common"
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

	badgeRepo       repository.BadgeRepository
	badgeDetailRepo repository.BadgeDetailRepository
}

func NewManager(
	badgeRepo repository.BadgeRepository,
	badgeDetailRepo repository.BadgeDetailRepository,
	badgeScanners ...BadgeScanner,
) *Manager {
	manager := &Manager{
		badgeRepo:       badgeRepo,
		badgeDetailRepo: badgeDetailRepo,
		badgeScanners:   make(map[string]BadgeScanner),
	}

	for _, b := range badgeScanners {
		manager.badgeScanners[b.Name()] = b
	}

	return manager
}

func (m *Manager) GetAllBadgeNames() []string {
	return common.MapKeys(m.badgeScanners)
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

func (c *contextManager) ScanAndGive(ctx context.Context, userID, communityID string) error {
	for _, badgeName := range c.badgeNames {
		badgeScanner, ok := c.manager.badgeScanners[badgeName]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found badge name %s", badgeName)
			return errorx.Unknown
		}

		suitableBadges, err := badgeScanner.Scan(ctx, userID, communityID)
		if err != nil {
			return err
		}

		// No need to update if cannot scan any suitable badge.
		if len(suitableBadges) == 0 {
			continue
		}

		actualCommunityID := sql.NullString{Valid: true, String: communityID}
		if badgeScanner.IsGlobal() {
			actualCommunityID = sql.NullString{Valid: false}
		}

		// Get the current level badge which user received. We need only give
		// user badges which is higher level.
		latestLevel := 0
		latestBadgeDetail, err := c.manager.badgeDetailRepo.GetLatest(
			ctx, userID, communityID, badgeScanner.Name())
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				xcontext.Logger(ctx).Errorf("Cannot get the latest badge detail: %v", err)
				return errorx.Unknown
			}
		} else {
			latestBadge, err := c.manager.badgeRepo.GetByID(ctx, latestBadgeDetail.BadgeID)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get the latest badge: %v", err)
				return errorx.Unknown
			}

			latestLevel = latestBadge.Level
		}

		for _, badge := range suitableBadges {
			if badge.Level <= latestLevel {
				continue
			}

			newBadgeDetail := &entity.BadgeDetail{
				UserID:      userID,
				CommunityID: actualCommunityID,
				BadgeID:     badge.ID,
				WasNotified: false,
			}

			if err := c.manager.badgeDetailRepo.Create(ctx, newBadgeDetail); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot create new badge to user: %v", err)
				return errorx.Unknown
			}
		}
	}

	return nil
}
