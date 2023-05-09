package testutil

import (
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type MockBadge struct {
	NameValue     string
	IsGlobalValue bool
	ScanFunc      func(ctx xcontext.Context, userID, projectID string) (int, error)
}

func (b *MockBadge) Name() string {
	return b.NameValue
}

func (b *MockBadge) IsGlobal() bool {
	return b.IsGlobalValue
}

func (b *MockBadge) Scan(ctx xcontext.Context, userID, projectID string) (int, error) {
	if b.ScanFunc != nil {
		return b.ScanFunc(ctx, userID, projectID)
	}

	return 0, errorx.New(errorx.NotImplemented, "Not implemented")
}
