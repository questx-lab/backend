package questutil

import (
	"github.com/questx-lab/backend/pkg/router"
)

type Validator interface {
	Validate(ctx router.Context, input string) (bool, error)
}

type Condition interface {
	Check(ctx router.Context) (bool, error)
}

type Award interface {
	Give(ctx router.Context) error
}
