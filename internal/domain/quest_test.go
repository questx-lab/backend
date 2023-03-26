package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_questDomain_Create(t *testing.T) {
	db := testutil.DefaultTestDb(t)
	projectRepo := repository.NewProjectRepository(db)
	questRepo := repository.NewQuestRepository(db)

	questDomain := NewQuestDomain(questRepo, projectRepo)

	t.Run("create quest successfully", func(t *testing.T) {
		project, err := testutil.SampleProject(db, nil)
		assert.NoError(t, err)

		createQuestReq := &model.CreateQuestRequest{
			ProjectID: project.ID,
			Title:     "new-quest",
		}

		ctx := testutil.NewMockContextWithUserID(project.CreatedBy)
		questResp, err := questDomain.Create(ctx, createQuestReq)
		assert.NoError(t, err)
		assert.NotEmpty(t, questResp.ID)

		var result entity.Quest
		tx := db.Model(&entity.Quest{}).Take(&result, "id", questResp.ID)
		assert.NoError(t, tx.Error)
		assert.Equal(t, result.ProjectID, project.ID)
		assert.Equal(t, result.Status, "draft")
		assert.Equal(t, result.Title, createQuestReq.Title)
	})

	t.Run("no perrmission to create quest", func(t *testing.T) {
		project, err := testutil.SampleProject(db, &entity.Project{CreatedBy: "user-0"})
		assert.NoError(t, err)

		createQuestReq := &model.CreateQuestRequest{
			ProjectID: project.ID,
			Title:     "new-quest",
		}

		ctx := testutil.NewMockContextWithUserID("user-1")
		_, err = questDomain.Create(ctx, createQuestReq)
		assert.ErrorAs(t, err, &errorx.Error{})
	})
}

func Test_questDomain_GetShortForm(t *testing.T) {
	db := testutil.GetEmptyTestDb()
	projectRepo := repository.NewProjectRepository(db)
	questRepo := repository.NewQuestRepository(db)

	questDomain := NewQuestDomain(questRepo, projectRepo)

	project, err := testutil.SampleProject(db, nil)
	assert.NoError(t, err)

	createQuestReq := &model.CreateQuestRequest{
		ProjectID: project.ID,
		Title:     "new-quest",
	}

	ctx := testutil.NewMockContextWithUserID(project.CreatedBy)
	questResp, err := questDomain.Create(ctx, createQuestReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, questResp.ID)

	shortFormResp, err := questDomain.GetShortForm(ctx, &model.GetShortQuestRequest{ID: questResp.ID})
	assert.NoError(t, err)
	assert.Equal(t, shortFormResp.ProjectID, createQuestReq.ProjectID)
	assert.Equal(t, shortFormResp.Title, createQuestReq.Title)
}
