package domain

import (
	"crypto/sha256"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type APIKeyDomain interface {
	Generate(xcontext.Context, *model.GenerateAPIKeyRequest) (*model.GenerateAPIKeyResponse, error)
	Regenerate(xcontext.Context, *model.RegenerateAPIKeyRequest) (*model.RegenerateAPIKeyResponse, error)
	Revoke(xcontext.Context, *model.RevokeAPIKeyRequest) (*model.RevokeAPIKeyResponse, error)
}

type apiKeyDomain struct {
	apiKeyRepo repository.APIKeyRepository
}

func NewAPIKeyDomain(apiKeyRepo repository.APIKeyRepository) *apiKeyDomain {
	return &apiKeyDomain{apiKeyRepo: apiKeyRepo}
}

func (d *apiKeyDomain) Generate(
	ctx xcontext.Context, req *model.GenerateAPIKeyRequest,
) (*model.GenerateAPIKeyResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	// TODO: Only project owner can create api key.

	key, err := generateRandomString()
	if err != nil {
		ctx.Logger().Errorf("Cannot generate api key: %v", err)
		return nil, errorx.Unknown
	}

	hashedKey := sha256.Sum256([]byte(key))
	err = d.apiKeyRepo.Create(ctx, &entity.APIKey{
		ProjectID: req.ProjectID,
		Key:       string(hashedKey[:]),
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

	// TODO: Only project owner can regenerate api key.

	key, err := generateRandomString()
	if err != nil {
		ctx.Logger().Errorf("Cannot generate api key: %v", err)
		return nil, errorx.Unknown
	}

	hashedKey := sha256.Sum256([]byte(key))
	err = d.apiKeyRepo.Update(ctx, req.ProjectID, string(hashedKey[:]))
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

	// TODO: Only project owner can delete api key.

	err := d.apiKeyRepo.DeleteByProjectID(ctx, req.ProjectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot delete api key: %v", err)
		return nil, errorx.Unknown
	}

	return &model.RevokeAPIKeyResponse{}, nil
}
