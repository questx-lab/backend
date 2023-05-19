package badge

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

const QuestWarriorBadgeName = "quest warrior"

// questWarriorBadgeScanner scans badge level based on the number of quests user
// claimed.
type questWarriorBadgeScanner struct {
	levelConfig       []uint64
	userAggregateRepo repository.UserAggregateRepository
}

func NewQuestWarriorBadgeScanner(
	userAggregateRepo repository.UserAggregateRepository,
	levelConfig []uint64,
) *questWarriorBadgeScanner {
	return &questWarriorBadgeScanner{levelConfig: levelConfig, userAggregateRepo: userAggregateRepo}
}

func (*questWarriorBadgeScanner) Name() string {
	return QuestWarriorBadgeName
}

func (*questWarriorBadgeScanner) IsGlobal() bool {
	return false
}

func (s *questWarriorBadgeScanner) Scan(ctx context.Context, userID, projectID string) (int, error) {
	userAggregate, err := s.userAggregateRepo.GetTotal(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}

		xcontext.Logger(ctx).Errorf("Cannot get user aggregate: %v", err)
		return 0, errorx.Unknown
	}

	finalLevel := 0
	for level, value := range s.levelConfig {
		if userAggregate.TotalTask < value {
			break
		}
		finalLevel = level + 1
	}

	return finalLevel, nil
}
