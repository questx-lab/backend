package questclaim

import (
	"github.com/questx-lab/backend/pkg/router"
)

type Validator interface {
	// Return true if it's automatically accepted.
	// Return false if it's automatically rejected.
	// Return an errorx of NeedManualReview code for manual review later.
	// Always return errorx in this method.
	Validate(ctx router.Context, input string) (bool, error)
}

type Condition interface {
	// Return true if it passes the condition, otherwise, return false.W
	// Always return errorx in this method.
	Check(ctx router.Context) (bool, error)
}

type Award interface {
	Give(ctx router.Context) error
}
