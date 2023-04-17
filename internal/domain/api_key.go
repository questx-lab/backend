package domain

import (
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type APIKeyDomain interface {
	Generate(xcontext.Context, *model.GenerateAPIKeyRequest) (*model.GenerateAPIKeyResponse, error)
	Regenerate(xcontext.Context, *model.RegenerateAPIKeyRequest) (*model.RegenerateAPIKeyResponse, error)
	Revoke(xcontext.Context, *model.RevokeAPIKeyRequest) (*model.RevokeAPIKeyResponse, error)
}

type apiKeyDomain struct {
	apiKeyRepo   repository.APIKeyRepository
	roleVerifier *common.ProjectRoleVerifier
}

func NewAPIKeyDomain(
	apiKeyRepo repository.APIKeyRepository,
	collaboratorRepo repository.CollaboratorRepository,
) *apiKeyDomain {
	return &apiKeyDomain{
		apiKeyRepo:   apiKeyRepo,
		roleVerifier: common.NewProjectRoleVerifier(collaboratorRepo),
	}
}

func (d *apiKeyDomain) Generate(
	ctx xcontext.Context, req *model.GenerateAPIKeyRequest,
) (*model.GenerateAPIKeyResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.Owner); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	key, err := crypto.GenerateRandomString()
	if err != nil {
		ctx.Logger().Errorf("Cannot generate api key: %v", err)
		return nil, errorx.Unknown
	}

	err = d.apiKeyRepo.Create(ctx, &entity.APIKey{
		ProjectID: req.ProjectID,
		Key:       crypto.Hash([]byte(key)),
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot save api key: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GenerateAPIKeyResponse{Key: key}, nil
}

func (d *apiKeyDomain) Regenerate(
	ctx xcontext.Context, req *model.RegenerateAPIKeyRequest,
) (*model.RegenerateAPIKeyResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.Owner); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	key, err := crypto.GenerateRandomString()
	if err != nil {
		ctx.Logger().Errorf("Cannot generate api key: %v", err)
		return nil, errorx.Unknown
	}

	err = d.apiKeyRepo.Update(ctx, req.ProjectID, crypto.Hash([]byte(key)))
	if err != nil {
		ctx.Logger().Errorf("Cannot save api key: %v", err)
		return nil, errorx.Unknown
	}

	return &model.RegenerateAPIKeyResponse{Key: key}, nil
}

func (d *apiKeyDomain) Revoke(
	ctx xcontext.Context, req *model.RevokeAPIKeyRequest,
) (*model.RevokeAPIKeyResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.Owner); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	err := d.apiKeyRepo.DeleteByProjectID(ctx, req.ProjectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot delete api key: %v", err)
		return nil, errorx.Unknown
	}

	return &model.RevokeAPIKeyResponse{}, nil
}
