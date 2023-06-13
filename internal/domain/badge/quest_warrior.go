package badge

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

const QuestWarriorBadgeName = "quest_warrior"

// questWarriorBadgeScanner scans badge level based on the number of quests user
// claimed.
type questWarriorBadgeScanner struct {
	badgeRepo    repository.BadgeRepository
	followerRepo repository.FollowerRepository
}

func NewQuestWarriorBadgeScanner(
	badgeRepo repository.BadgeRepository,
	followerRepo repository.FollowerRepository,
) *questWarriorBadgeScanner {
	return &questWarriorBadgeScanner{
		badgeRepo:    badgeRepo,
		followerRepo: followerRepo,
	}
}

func (*questWarriorBadgeScanner) Name() string {
	return QuestWarriorBadgeName
}

func (*questWarriorBadgeScanner) IsGlobal() bool {
	return false
}

func (s *questWarriorBadgeScanner) Scan(ctx context.Context, userID, communityID string) ([]entity.Badge, error) {
	follower, err := s.followerRepo.Get(ctx, userID, communityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		xcontext.Logger(ctx).Errorf("Cannot get user aggregate: %v", err)
		return nil, errorx.Unknown
	}

	suitableBadges, err := s.badgeRepo.GetLessThanValue(ctx, s.Name(), int(follower.Quests))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get the suitable badge of %s: %v", s.Name(), err)
		return nil, errorx.Unknown
	}

	return suitableBadges, nil
}
