package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func Test_questDomain_Create(t *testing.T) {
	db := testutil.GetDatabaseTest()
	projectRepo := repository.NewProjectRepository(db)
	questRepo := repository.NewQuestRepository(db)

	questDomain := NewQuestDomain(questRepo, projectRepo)

	t.Run("create quest successfully", func(t *testing.T) {
		project, err := testutil.SampleProject(db, nil)
		require.NoError(t, err)

		createQuestReq := &model.CreateQuestRequest{
			ProjectID:   project.ID,
			Title:       "new-quest",
			Type:        "Visit Link",
			Recurrence:  "Once",
			ConditionOp: "OR",
		}

		ctx := testutil.NewMockContextWithUserID(project.CreatedBy)
		questResp, err := questDomain.Create(ctx, createQuestReq)
		require.NoError(t, err)
		require.NotEmpty(t, questResp.ID)

		var result entity.Quest
		tx := db.Model(&entity.Quest{}).Take(&result, "id", questResp.ID)
		require.NoError(t, tx.Error)
		require.Equal(t, result.ProjectID, project.ID)
		require.Equal(t, result.Status, enum.ToEnum[entity.QuestStatusType]("Draft"))
		require.Equal(t, result.Title, createQuestReq.Title)
		require.Equal(t, result.Type, enum.ToEnum[entity.QuestType]("Visit Link"))
		require.Equal(t, result.Recurrence, enum.ToEnum[entity.QuestRecurrenceType]("Once"))
		require.Equal(t, result.ConditionOp, enum.ToEnum[entity.QuestConditionOpType]("OR"))
	})

	t.Run("no perrmission to create quest", func(t *testing.T) {
		project, err := testutil.SampleProject(db, &entity.Project{CreatedBy: "user-0"})
		require.NoError(t, err)

		createQuestReq := &model.CreateQuestRequest{
			ProjectID: project.ID,
			Title:     "new-quest",
		}

		ctx := testutil.NewMockContextWithUserID("user-1")
		_, err = questDomain.Create(ctx, createQuestReq)
		require.ErrorAs(t, err, &errorx.Error{})
	})
}

func Test_questDomain_GetShortForm(t *testing.T) {
	db := testutil.GetDatabaseTest()
	projectRepo := repository.NewProjectRepository(db)
	questRepo := repository.NewQuestRepository(db)

	questDomain := NewQuestDomain(questRepo, projectRepo)

	project, err := testutil.SampleProject(db, nil)
	require.NoError(t, err)

	createQuestReq := &model.CreateQuestRequest{
		ProjectID:   project.ID,
		Title:       "new-quest",
		Type:        "Visit Link",
		Recurrence:  "Once",
		ConditionOp: "OR",
	}

	ctx := testutil.NewMockContextWithUserID(project.CreatedBy)
	questResp, err := questDomain.Create(ctx, createQuestReq)
	require.NoError(t, err)
	require.NotEmpty(t, questResp.ID)

	shortFormResp, err := questDomain.GetShortForm(ctx, &model.GetShortQuestRequest{ID: questResp.ID})
	require.NoError(t, err)
	require.Equal(t, shortFormResp.ProjectID, createQuestReq.ProjectID)
	require.Equal(t, shortFormResp.Title, createQuestReq.Title)
}
