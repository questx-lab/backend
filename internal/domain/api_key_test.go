package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
)

func Test_apiKeyDomain_FullScenario(t *testing.T) {
	ctx := testutil.MockContext()
	testutil.CreateFixtureDb(ctx)

	apiKeyDomain := &apiKeyDomain{
		apiKeyRepo:   repository.NewAPIKeyRepository(),
		roleVerifier: common.NewCommunityRoleVerifier(repository.NewCollaboratorRepository(), repository.NewUserRepository()),
	}

	// Generate successfully.
	ctxUser1 := xcontext.WithRequestUserID(ctx, testutil.Community1.CreatedBy)
	_, err := apiKeyDomain.Generate(
		ctxUser1, &model.GenerateAPIKeyRequest{CommunityID: testutil.Community1.ID})
	require.NoError(t, err)

	// Cannot generate more than one API Key for a community.
	_, err = apiKeyDomain.Generate(
		ctxUser1, &model.GenerateAPIKeyRequest{CommunityID: testutil.Community1.ID})
	require.Equal(t, "Request failed", err.Error())

	// However, regenerate successfully.
	_, err = apiKeyDomain.Regenerate(
		ctxUser1, &model.RegenerateAPIKeyRequest{CommunityID: testutil.Community1.ID})
	require.NoError(t, err)

	// Revoke successfully.
	_, err = apiKeyDomain.Revoke(
		ctxUser1, &model.RevokeAPIKeyRequest{CommunityID: testutil.Community1.ID})
	require.NoError(t, err)
}
