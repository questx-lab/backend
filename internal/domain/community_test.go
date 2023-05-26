package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/stretchr/testify/require"
)

func Test_communityDomain_Create(t *testing.T) {
	ctx := testutil.MockContextWithUserID(testutil.User1.ID)
	testutil.CreateFixtureDb(ctx)
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{})
	collaboratorRepo := repository.NewCollaboratorRepository()
	userRepo := repository.NewUserRepository()
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	domain := NewCommunityDomain(communityRepo, collaboratorRepo, userRepo, questRepo, nil, nil)

	req := &model.CreateCommunityRequest{
		Handle:      "test",
		DisplayName: "for testing",
		Twitter:     "https://twitter.com/hashtag/Breaking2",
	}
	resp, err := domain.Create(ctx, req)
	require.NoError(t, err)

	var result entity.Community
	tx := xcontext.DB(ctx).Model(&entity.Community{}).Take(&result, "id", resp.ID)
	require.NoError(t, tx.Error)
	require.Equal(t, result.Handle, req.Handle)
	require.Equal(t, result.Twitter, req.Twitter)
	require.Equal(t, result.CreatedBy, testutil.User1.ID)
}
