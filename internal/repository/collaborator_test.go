package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/testutil"
	"gorm.io/gorm"
)

func setup_collaboratorRepository(db *gorm.DB) {
	db.Create(&entity.User{
		Base: entity.Base{
			ID:        "valid-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Address: "address",
		Name:    "name",
	})

	db.Create(&entity.Project{
		Base: entity.Base{
			ID:        "valid-project-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CreatedBy: "valid-created-by-id",
		Name:      "name",
		Twitter:   "twitter",
		Discord:   "discord",
		Telegram:  "telegram",
	})
}
func Test_collaboratorRepository_Create(t *testing.T) {
	ctx := context.Background()
	db := testutil.GetDatabaseTest()

	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx context.Context
		e   *entity.Collaborator
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
		setup   func(db *gorm.DB)
	}{
		{
			name: "happy case",
			fields: fields{
				db: db,
			},
			args: args{
				ctx: ctx,
				e: &entity.Collaborator{
					Base: entity.Base{
						ID:        uuid.NewString(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					UserID:    "valid-user-id",
					ProjectID: "valid-project-id",
					Role:      entity.CollaboratorRoleOwner,
					CreatedBy: "valid-created-by",
				},
			},
			setup: setup_collaboratorRepository,
		},
		{
			name: "invalid userID",
			fields: fields{
				db: db,
			},
			args: args{
				ctx: ctx,
				e: &entity.Collaborator{
					Base: entity.Base{
						ID:        uuid.NewString(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					UserID:    "invalid-user-id",
					ProjectID: "valid-project-id",
					Role:      entity.CollaboratorRoleOwner,
					CreatedBy: "valid-created-by",
				},
			},
			setup:   setup_collaboratorRepository,
			wantErr: gorm.ErrInvalidData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &collaboratorRepository{
				db: tt.fields.db,
			}
			if err := r.Create(tt.args.ctx, tt.args.e); !errors.Is(tt.wantErr, err) {
				t.Errorf("collaboratorRepository.Create() expect error = %v, but got error = %v", err, tt.wantErr)
			}
		})
	}
}

func Test_collaboratorRepository_GetList(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx    context.Context
		offset int
		limit  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*entity.Collaborator
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &collaboratorRepository{
				db: tt.fields.db,
			}
			got, err := r.GetList(tt.args.ctx, tt.args.offset, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("collaboratorRepository.GetList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collaboratorRepository.GetList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_collaboratorRepository_DeleteByID(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &collaboratorRepository{
				db: tt.fields.db,
			}
			if err := r.DeleteByID(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("collaboratorRepository.DeleteByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
