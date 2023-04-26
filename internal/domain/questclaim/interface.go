package questclaim

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type ActionForClaim string

const (
	Accepted         = ActionForClaim("accepted")
	Rejected         = ActionForClaim("rejected")
	NeedManualReview = ActionForClaim("need_manual_review")
)

// Processor automatically reviews the input or action of user with the validation data. It helps
// us to determine we should accept, reject, or manual review the claimed quest.
type Processor interface {
	// Always return errorx in this method.
	GetActionForClaim(ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string) (ActionForClaim, error)
}

// Condition is the prerequisite to claim the quest.
type Condition interface {
	// Always return errorx in this method.
	Check(ctx xcontext.Context) (bool, error)
}

// Reward gives rewards (point, badge, etc.) to user after the claimed quest is accepted.
type Reward interface {
	// Always return errorx in this method.
	Give(ctx xcontext.Context) error
}
