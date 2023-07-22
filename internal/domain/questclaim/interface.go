package questclaim

import (
	"context"
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
)

type ActionForClaim interface {
	Name() string
	Message() string
	Is(ActionForClaim) bool
	WithMessage(m string, a ...any) ActionForClaim
}

type actionForClaim struct {
	name    string
	message string
}

func (a actionForClaim) Name() string {
	return a.name
}

func (a actionForClaim) Message() string {
	return a.message
}

func (a actionForClaim) WithMessage(m string, args ...any) ActionForClaim {
	a.message = fmt.Sprintf(m, args...)
	return a
}

func (a actionForClaim) Is(another ActionForClaim) bool {
	return a.Name() == another.Name()
}

var (
	Accepted         = actionForClaim{name: "accepted"}
	Rejected         = actionForClaim{name: "rejected"}
	NeedManualReview = actionForClaim{name: "need_manual_review"}
)

// Processor automatically reviews the submission data or action of user with
// the validation data. It helps us to determine we should accept, reject, or
// manual review the claimed quest.
type Processor interface {
	// Always return errorx in this method.
	GetActionForClaim(ctx context.Context, submissionData string) (ActionForClaim, error)

	// RetryAfter returns the necessary time the user must wait for claiming
	// a quest after it was auto rejected.
	RetryAfter() time.Duration
}

// Condition is the prerequisite to claim the quest.
type Condition interface {
	// Always return errorx in this method.
	Check(ctx context.Context) (bool, error)

	// Statement returns the condition statement of this condition.
	Statement() string
}

// Reward gives rewards (point, etc.) to user.
type Reward interface {
	// Always return errorx in this method.
	Give(ctx context.Context) error

	WithClaimedQuest(claimedQuest *entity.ClaimedQuest)
	WithLuckybox(luckybox *entity.GameLuckybox)
	WithReferralCommunity(referralCommunity *entity.Community)
	WithLotteryWinner(winner *entity.LotteryWinner)
	WithWalletAddress(chain, address string)
}
