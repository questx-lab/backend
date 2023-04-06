package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func Test_apiKeyDomain_FullScenario(t *testing.T) {
	ctx := testutil.NewMockContext()
	testutil.CreateFixtureDb(ctx)

	apiKeyDomain := &apiKeyDomain{
		apiKeyRepo:   repository.NewAPIKeyRepository(),
		roleVerifier: common.NewProjectRoleVerifier(repository.NewCollaboratorRepository()),
	}

	// Generate successfully.
	ctxUser1 := testutil.NewMockContextWithUserID(ctx, testutil.Project1.CreatedBy)
	_, err := apiKeyDomain.Generate(
		ctxUser1, &model.GenerateAPIKeyRequest{ProjectID: testutil.Project1.ID})
	require.NoError(t, err)

	// Cannot generate more than one API Key for a project.
	_, err = apiKeyDomain.Generate(
		ctxUser1, &model.GenerateAPIKeyRequest{ProjectID: testutil.Project1.ID})
	require.Equal(t, "Request failed", err.Error())

	// However, regenerate successfully.
	_, err = apiKeyDomain.Regenerate(
		ctxUser1, &model.RegenerateAPIKeyRequest{ProjectID: testutil.Project1.ID})
	require.NoError(t, err)

	// Revoke successfully.
	_, err = apiKeyDomain.Revoke(
		ctxUser1, &model.RevokeAPIKeyRequest{ProjectID: testutil.Project1.ID})
	require.NoError(t, err)
}
