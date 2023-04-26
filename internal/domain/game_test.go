package domain

import (
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/stretchr/testify/require"
)

func Test_verifyUserRole(t *testing.T) {
	type args struct {
		ctx         xcontext.Context
		acceptRoles []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		// TODO: Add test cases.
		{
			name: "happy case",
			args: args{
				ctx:         testutil.NewMockContextWithUserID(nil, testutil.User1.ID),
				acceptRoles: []string{entity.SuperAdminRole},
			},
		},
		{
			name: "user not found",
			args: args{
				ctx:         testutil.NewMockContextWithUserID(nil, "invalid_id"),
				acceptRoles: []string{entity.SuperAdminRole},
			},
			wantErr: errorx.New(errorx.NotFound, "Not found user"),
		},
		{
			name: "user not have permission",
			args: args{
				ctx:         testutil.NewMockContextWithUserID(nil, testutil.User2.ID),
				acceptRoles: []string{entity.SuperAdminRole},
			},
			wantErr: errorx.New(errorx.Unauthenticated, "User doesn't have permission"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CreateFixtureDb(tt.args.ctx)
			userRepo := repository.NewUserRepository()
			if err := verifyUserRole(tt.args.ctx, userRepo, tt.args.acceptRoles); err != nil || tt.wantErr != nil {
				require.Error(t, err)
				require.Error(t, tt.wantErr)
				require.Equal(t, err.Error(), tt.wantErr.Error())
			}
		})
	}
}
