package domain

import (
	"context"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type APIKeyDomain interface {
	Generate(context.Context, *model.GenerateAPIKeyRequest) (*model.GenerateAPIKeyResponse, error)
	Regenerate(context.Context, *model.RegenerateAPIKeyRequest) (*model.RegenerateAPIKeyResponse, error)
	Revoke(context.Context, *model.RevokeAPIKeyRequest) (*model.RevokeAPIKeyResponse, error)
}

type apiKeyDomain struct {
	apiKeyRepo   repository.APIKeyRepository
	roleVerifier *common.CommunityRoleVerifier
}

func NewAPIKeyDomain(
	apiKeyRepo repository.APIKeyRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
) *apiKeyDomain {
	return &apiKeyDomain{
		apiKeyRepo:   apiKeyRepo,
		roleVerifier: common.NewCommunityRoleVerifier(collaboratorRepo, userRepo),
	}
}

func (d *apiKeyDomain) Generate(
	ctx context.Context, req *model.GenerateAPIKeyRequest,
) (*model.GenerateAPIKeyResponse, error) {
	if req.CommunityID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community id")
	}

	if err := d.roleVerifier.Verify(ctx, req.CommunityID, entity.Owner); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	key, err := crypto.GenerateRandomString()
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate api key: %v", err)
		return nil, errorx.Unknown
	}

	err = d.apiKeyRepo.Create(ctx, &entity.APIKey{
		CommunityID: req.CommunityID,
		Key:         crypto.SHA256([]byte(key)),
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot save api key: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GenerateAPIKeyResponse{Key: key}, nil
}

func (d *apiKeyDomain) Regenerate(
	ctx context.Context, req *model.RegenerateAPIKeyRequest,
) (*model.RegenerateAPIKeyResponse, error) {
	if req.CommunityID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community id")
	}

	if err := d.roleVerifier.Verify(ctx, req.CommunityID, entity.Owner); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	key, err := crypto.GenerateRandomString()
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate api key: %v", err)
		return nil, errorx.Unknown
	}

	err = d.apiKeyRepo.Update(ctx, req.CommunityID, crypto.SHA256([]byte(key)))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot save api key: %v", err)
		return nil, errorx.Unknown
	}

	return &model.RegenerateAPIKeyResponse{Key: key}, nil
}

func (d *apiKeyDomain) Revoke(
	ctx context.Context, req *model.RevokeAPIKeyRequest,
) (*model.RevokeAPIKeyResponse, error) {
	if req.CommunityID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community id")
	}

	if err := d.roleVerifier.Verify(ctx, req.CommunityID, entity.Owner); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	err := d.apiKeyRepo.DeleteByCommunityID(ctx, req.CommunityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete api key: %v", err)
		return nil, errorx.Unknown
	}

	return &model.RevokeAPIKeyResponse{}, nil
}
