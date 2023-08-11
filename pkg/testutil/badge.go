package testutil

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
)

type MockBadge struct {
	NameValue     string
	IsGlobalValue bool
	ScanFunc      func(ctx context.Context, userID, communityID string) ([]entity.Badge, error)
}

func (b *MockBadge) Name() string {
	return b.NameValue
}

func (b *MockBadge) IsGlobal() bool {
	return b.IsGlobalValue
}

func (b *MockBadge) Scan(ctx context.Context, userID, communityID string) ([]entity.Badge, error) {
	if b.ScanFunc != nil {
		return b.ScanFunc(ctx, userID, communityID)
	}

	return nil, nil
}
