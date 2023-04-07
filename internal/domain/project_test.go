package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/questx-lab/backend/pkg/testutil"

	"github.com/stretchr/testify/require"
)

func Test_projectDomain_Create(t *testing.T) {
	ctx := testutil.NewMockContextWithUserID(nil, testutil.User1.ID)
	testutil.CreateFixtureDb(ctx)
	projectRepo := repository.NewProjectRepository()
	collaboratorRepo := repository.NewCollaboratorRepository()
	domain := NewProjectDomain(projectRepo, collaboratorRepo)

	req := &model.CreateProjectRequest{
		Name:     "test",
		Twitter:  "https://twitter.com/hashtag/Breaking2",
		Discord:  "https://discord.com/hashtag/Breaking2",
		Telegram: "https://telegram.com/",
	}
	resp, err := domain.Create(ctx, req)
	require.NoError(t, err)

	var result entity.Project
	tx := ctx.DB().Model(&entity.Project{}).Take(&result, "id", resp.ID)
	require.NoError(t, tx.Error)
	require.Equal(t, result.Name, req.Name)
	require.Equal(t, result.Discord, req.Discord)
	require.Equal(t, result.Twitter, req.Twitter)
	require.Equal(t, result.Telegram, req.Telegram)
	require.Equal(t, result.CreatedBy, testutil.User1.ID)
}

func Test_projectDomain_GetMyList(t *testing.T) {
	ctx := testutil.NewMockContextWithUserID(nil, testutil.Project1.CreatedBy)
	testutil.CreateFixtureDb(ctx)
	domain := NewProjectDomain(repository.NewProjectRepository(), repository.NewCollaboratorRepository())
	result, err := domain.GetMyList(ctx, &model.GetMyListProjectRequest{
		Offset: 0,
		Limit:  10,
	})

	require.NoError(t, err)
	require.Equal(t, 1, len(result.Projects))

	actual := result.Projects[0]

	expected := model.Project{
		ID:        testutil.Project1.ID,
		Name:      testutil.Project1.Name,
		CreatedBy: testutil.Project1.CreatedBy,
	}

	require.True(t, reflectutil.PartialEqual(&expected, &actual))
}
