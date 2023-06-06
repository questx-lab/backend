package domain

import (
	"context"
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/reflectutil"
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
	domain := NewCommunityDomain(communityRepo, collaboratorRepo, userRepo, questRepo, nil, nil, nil)

	req := &model.CreateCommunityRequest{
		Handle:      "test",
		DisplayName: "for testing",
		Twitter:     "https://twitter.com/hashtag/Breaking2",
	}
	resp, err := domain.Create(ctx, req)
	require.NoError(t, err)

	var result entity.Community
	tx := xcontext.DB(ctx).Model(&entity.Community{}).Take(&result, "handle", resp.Handle)
	require.NoError(t, tx.Error)
	require.Equal(t, result.Handle, req.Handle)
	require.Equal(t, result.Twitter, req.Twitter)
	require.Equal(t, result.CreatedBy, testutil.User1.ID)
}

func Test_communityDomain_TransferCommunity(t *testing.T) {
	ctx := testutil.MockContextWithUserID(testutil.User1.ID)
	testutil.CreateFixtureDb(ctx)
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{})
	collaboratorRepo := repository.NewCollaboratorRepository()
	userRepo := repository.NewUserRepository()
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	domain := NewCommunityDomain(communityRepo, collaboratorRepo, userRepo, questRepo, nil, nil, nil)
	type args struct {
		ctx context.Context
		req *model.TransferCommunityRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *model.TransferCommunityResponse
		wantErr error
		setup   func()
	}{
		{
			name: "happy case",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.TransferCommunityRequest{
					CommunityHandle: testutil.Community2.Handle,
					ToID:            testutil.User3.ID,
				},
			},
			want: &model.TransferCommunityResponse{},
		},
		{
			name: "err permission denied",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User2.ID),
				req: &model.TransferCommunityRequest{
					CommunityHandle: testutil.Community2.Handle,
					ToID:            testutil.User3.ID,
				},
			},
			wantErr: errorx.New(errorx.PermissionDenied, "Permission denied"),
		},
		{
			name: "err user not found",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.TransferCommunityRequest{
					CommunityHandle: testutil.Community2.Handle,
					ToID:            "wrong_to_id",
				},
			},
			wantErr: errorx.New(errorx.NotFound, "Not found user"),
		},
		{
			name: "err community not found",
			args: args{
				ctx: testutil.MockContextWithUserID(testutil.User1.ID),
				req: &model.TransferCommunityRequest{
					CommunityHandle: "community not found",
					ToID:            testutil.User2.ID,
				},
			},
			wantErr: errorx.New(errorx.NotFound, "Not found community"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)

			got, err := domain.TransferCommunity(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tt.wantErr.Error(), err.Error())
			}

			if tt.want == nil {
				require.Nil(t, got)
			} else {
				require.True(t, reflectutil.PartialEqual(tt.want, got), "%v != %v", tt.want, got)
			}
		})
	}

}

func Test_communityDomain_TransferCommunity_multi_transfer(t *testing.T) {
	ctx := testutil.MockContextWithUserID(testutil.User1.ID)
	testutil.CreateFixtureDb(ctx)
	communityRepo := repository.NewCommunityRepository(&testutil.MockSearchCaller{})
	collaboratorRepo := repository.NewCollaboratorRepository()
	userRepo := repository.NewUserRepository()
	questRepo := repository.NewQuestRepository(&testutil.MockSearchCaller{})
	domain := NewCommunityDomain(communityRepo, collaboratorRepo, userRepo, questRepo, nil, nil, nil)

	req := &model.TransferCommunityRequest{
		CommunityHandle: testutil.Community2.Handle,
		ToID:            testutil.User3.ID,
	}

	_, err := domain.TransferCommunity(ctx, req)
	require.NoError(t, err)

	req = &model.TransferCommunityRequest{
		CommunityHandle: testutil.Community2.Handle,
		ToID:            testutil.User2.ID,
	}

	_, err = domain.TransferCommunity(ctx, req)
	require.NoError(t, err)
}
