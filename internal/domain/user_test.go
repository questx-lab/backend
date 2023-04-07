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

	inviteResp, err := domain.GetInvite(ctx, &model.GetInviteRequest{
		InviteCode: testutil.Participant1.InviteCode,
	})
	require.NoError(t, err)
	require.Equal(t, inviteResp.Project.ID, testutil.Project1.ID)
	require.Equal(t, inviteResp.Project.Name, testutil.Project1.Name)
}
