package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func Test_userDomain_GetReferralInfo(t *testing.T) {
	ctx := testutil.NewMockContext()
	testutil.CreateFixtureDb(ctx)

	domain := NewUserDomain(
		repository.NewUserRepository(),
		repository.NewParticipantRepository(),
	)

	referralResp, err := domain.GetReferralInfo(ctx, &model.GetReferralInfoRequest{
		ReferralCode: testutil.Participant1.ReferralCode,
	})
	require.NoError(t, err)
	require.Equal(t, referralResp.Project.ID, testutil.Project1.ID)
	require.Equal(t, referralResp.Project.Name, testutil.Project1.Name)
}
