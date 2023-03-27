package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func Test_questDomain_Create(t *testing.T) {
	db := testutil.DefaultTestDb(t)
	projectRepo := repository.NewProjectRepository(db)
	questRepo := repository.NewQuestRepository(db)

	questDomain := NewQuestDomain(questRepo, projectRepo)

	t.Run("create quest successfully", func(t *testing.T) {
		p, err := projectRepo.GetByID(context.Background(), testutil.Project1.ID)
		require.NoError(t, err)

		createQuestReq := &model.CreateQuestRequest{
			ProjectID:   p.ID,
			Title:       "new-quest",
			Type:        "Visit Link",
			Recurrence:  "Once",
			ConditionOp: "OR",
		}

		ctx := testutil.NewMockContextWithUserID(p.CreatedBy)
		questResp, err := questDomain.Create(ctx, createQuestReq)
		require.NoError(t, err)
		require.NotEmpty(t, questResp.ID)

		var result entity.Quest
		tx := db.Model(&entity.Quest{}).Take(&result, "id", questResp.ID)
		require.NoError(t, tx.Error)
		require.Equal(t, result.ProjectID, p.ID)
		require.Equal(t, result.Status, entity.QuestStatusDraft)
		require.Equal(t, result.Title, createQuestReq.Title)
		require.Equal(t, result.Type, entity.QuestVisitLink)
		require.Equal(t, result.Recurrence, entity.QuestRecurrenceOnce)
		require.Equal(t, result.ConditionOp, entity.QuestConditionOpOr)
	})

	t.Run("no perrmission to create quest", func(t *testing.T) {
		// 1. New project created by user 2
		otherProject := &entity.Project{
			Base: entity.Base{
				ID: "user2_project1",
			},
			Name:      "User2 Project1",
			CreatedBy: testutil.User2.ID,
		}
		err := projectRepo.Create(context.Background(), otherProject)
		require.NoError(t, err)

		// 2. Verify that user1 cannot create this project.
		createQuestReq := &model.CreateQuestRequest{
			ProjectID: otherProject.ID,
			Title:     "new-quest",
		}

		ctx := testutil.NewMockContextWithUserID(testutil.User1.ID)
		_, err = questDomain.Create(ctx, createQuestReq)
		require.ErrorAs(t, err, &errorx.Error{})
	})
}

func Test_questDomain_GetShortForm(t *testing.T) {
	db := testutil.DefaultTestDb(t)
	projectRepo := repository.NewProjectRepository(db)
	questRepo := repository.NewQuestRepository(db)

	questDomain := NewQuestDomain(questRepo, projectRepo)

	project, err := projectRepo.GetByID(context.Background(), "user1_project1")
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
