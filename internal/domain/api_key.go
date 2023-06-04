package domain

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type APIKeyDomain interface {
	Generate(context.Context, *model.GenerateAPIKeyRequest) (*model.GenerateAPIKeyResponse, error)
	Regenerate(context.Context, *model.RegenerateAPIKeyRequest) (*model.RegenerateAPIKeyResponse, error)
	Revoke(context.Context, *model.RevokeAPIKeyRequest) (*model.RevokeAPIKeyResponse, error)
}

type apiKeyDomain struct {
	apiKeyRepo    repository.APIKeyRepository
	communityRepo repository.CommunityRepository
	roleVerifier  *common.CommunityRoleVerifier
}

func NewAPIKeyDomain(
	apiKeyRepo repository.APIKeyRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	communityRepo repository.CommunityRepository,
) *apiKeyDomain {
	return &apiKeyDomain{
		apiKeyRepo:    apiKeyRepo,
		communityRepo: communityRepo,
		roleVerifier:  common.NewCommunityRoleVerifier(collaboratorRepo, userRepo),
	}
}

func (d *apiKeyDomain) Generate(
	ctx context.Context, req *model.GenerateAPIKeyRequest,
) (*model.GenerateAPIKeyResponse, error) {
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community handle")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, community.ID, entity.Owner); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	key, err := crypto.GenerateRandomString()
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate api key: %v", err)
		return nil, errorx.Unknown
	}

	err = d.apiKeyRepo.Create(ctx, &entity.APIKey{
		CommunityID: community.ID,
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
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community handle")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, community.ID, entity.Owner); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	key, err := crypto.GenerateRandomString()
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate api key: %v", err)
		return nil, errorx.Unknown
	}

	err = d.apiKeyRepo.Update(ctx, community.ID, crypto.SHA256([]byte(key)))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot save api key: %v", err)
		return nil, errorx.Unknown
	}

	return &model.RegenerateAPIKeyResponse{Key: key}, nil
}

func (d *apiKeyDomain) Revoke(
	ctx context.Context, req *model.RevokeAPIKeyRequest,
) (*model.RevokeAPIKeyResponse, error) {
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community handle")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, community.ID, entity.Owner); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if err := d.apiKeyRepo.DeleteByCommunityID(ctx, community.ID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete api key: %v", err)
		return nil, errorx.Unknown
	}

	return &model.RevokeAPIKeyResponse{}, nil
}
