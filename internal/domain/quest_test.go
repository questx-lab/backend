package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/testutil"

	"github.com/stretchr/testify/assert"
)

func createProject(name string, projectRepo repository.ProjectRepository) (string, error) {
	if err := projectRepo.Create(context.Background(), &entity.Project{
		Base:      entity.Base{ID: name},
		Name:      name,
		CreatedBy: name,
		Twitter:   "https://twitter.com/hashtag/Breaking2",
		Discord:   "https://discord.com/hashtag/Breaking2",
		Telegram:  "https://telegram.com/",
	}); err != nil {
		return "", nil
	}
	return name, nil
}

func Test_questDomain_Create(t *testing.T) {
	db := testutil.GetDatabaseTest()
	ctx := testutil.NewMockContextWithUserID(t.Name())

	projectRepo := repository.NewProjectRepository(db)
	questRepo := repository.NewQuestRepository(db)

	questDomain := NewQuestDomain(questRepo, projectRepo)

	projectID, err := createProject(t.Name(), projectRepo)
	assert.NoError(t, err)

	createQuestReq := &model.CreateQuestRequest{
		ProjectID: projectID,
		Title:     "new-quest",
	}

	questResp, err := questDomain.Create(ctx, createQuestReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, questResp.ID)

	var result entity.Quest
	tx := db.Model(&entity.Quest{}).Take(&result, "id", questResp.ID)
	assert.NoError(t, tx.Error)
	assert.Equal(t, result.ProjectID, projectID)
	assert.Equal(t, result.Status, "draft")
	assert.Equal(t, result.Title, createQuestReq.Title)
}

func Test_questDomain_Create_NoPermission(t *testing.T) {
	db := testutil.GetDatabaseTest()
	ctx := testutil.NewMockContextWithUserID("user-with-no-permission")

	projectRepo := repository.NewProjectRepository(db)
	questRepo := repository.NewQuestRepository(db)

	questDomain := NewQuestDomain(questRepo, projectRepo)

	projectID, err := createProject(t.Name(), projectRepo)
	assert.NoError(t, err)

	createQuestReq := &model.CreateQuestRequest{
		ProjectID: projectID,
		Title:     "new-quest",
	}

	_, err = questDomain.Create(ctx, createQuestReq)
	assert.ErrorAs(t, err, &errorx.Error{})
}

func Test_questDomain_GetShortForm(t *testing.T) {
	db := testutil.GetDatabaseTest()
	ctx := testutil.NewMockContextWithUserID(t.Name())

	projectRepo := repository.NewProjectRepository(db)
	questRepo := repository.NewQuestRepository(db)

	questDomain := NewQuestDomain(questRepo, projectRepo)

	projectID, err := createProject(t.Name(), projectRepo)
	assert.NoError(t, err)

	createQuestReq := &model.CreateQuestRequest{
		ProjectID: projectID,
		Title:     "new-quest",
	}

	questResp, err := questDomain.Create(ctx, createQuestReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, questResp.ID)

	shortFormResp, err := questDomain.GetShortForm(ctx, &model.GetShortQuestRequest{ID: questResp.ID})
	assert.NoError(t, err)
	assert.Equal(t, shortFormResp.ProjectID, createQuestReq.ProjectID)
	assert.Equal(t, shortFormResp.Title, createQuestReq.Title)
}
