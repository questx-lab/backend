package domain

import (
	"context"
	"database/sql"
	"errors"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryDomain interface {
	Create(context.Context, *model.CreateCategoryRequest) (*model.CreateCategoryResponse, error)
	GetList(context.Context, *model.GetListCategoryRequest) (*model.GetListCategoryResponse, error)
	GetTemplate(context.Context, *model.GetTemplateCategoryRequest) (*model.GetTemplateCategoryResponse, error)
	UpdateByID(context.Context, *model.UpdateCategoryByIDRequest) (*model.UpdateCategoryByIDResponse, error)
	DeleteByID(context.Context, *model.DeleteCategoryByIDRequest) (*model.DeleteCategoryByIDResponse, error)
}

type categoryDomain struct {
	categoryRepo  repository.CategoryRepository
	questRepo     repository.QuestRepository
	communityRepo repository.CommunityRepository
	roleVerifier  *common.CommunityRoleVerifier
}

func NewCategoryDomain(
	categoryRepo repository.CategoryRepository,
	questRepo repository.QuestRepository,
	communityRepo repository.CommunityRepository,
	roleVerifier *common.CommunityRoleVerifier,
) CategoryDomain {
	return &categoryDomain{
		categoryRepo:  categoryRepo,
		questRepo:     questRepo,
		communityRepo: communityRepo,
		roleVerifier:  roleVerifier,
	}
}

func (d *categoryDomain) Create(ctx context.Context, req *model.CreateCategoryRequest) (*model.CreateCategoryResponse, error) {
	if req.Name == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow an empty category name")
	}

	communityID := ""
	if req.CommunityHandle != "" {
		community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found community")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}

		communityID = community.ID
	}

	if err := d.roleVerifier.Verify(ctx, communityID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if _, err := d.categoryRepo.GetByName(ctx, communityID, req.Name); !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			return nil, errorx.New(errorx.AlreadyExists, "Duplicated category name")
		}

		xcontext.Logger(ctx).Errorf("Cannot get category by name: %v", err)
		return nil, errorx.Unknown
	}

	lastPosition, err := d.categoryRepo.GetLastPosition(ctx, communityID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		xcontext.Logger(ctx).Errorf("Cannot get last position: %v", err)
		return nil, errorx.Unknown
	}

	if err == nil {
		lastPosition += 1
	}

	category := &entity.Category{
		Base:        entity.Base{ID: uuid.NewString()},
		CommunityID: sql.NullString{Valid: true, String: communityID},
		Name:        req.Name,
		CreatedBy:   xcontext.RequestUserID(ctx),
		Position:    lastPosition,
	}

	if communityID == "" {
		category.CommunityID = sql.NullString{Valid: false}
	}

	if err := d.categoryRepo.Create(ctx, category); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create category: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateCategoryResponse{Category: model.ConvertCategory(category)}, nil
}

func (d *categoryDomain) GetList(
	ctx context.Context, req *model.GetListCategoryRequest,
) (*model.GetListCategoryResponse, error) {
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

	categoryEntities, err := d.categoryRepo.GetList(ctx, community.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get the category list: %v", err)
		return nil, errorx.Unknown
	}

	data := []model.Category{}
	for _, e := range categoryEntities {
		data = append(data, model.ConvertCategory(&e))
	}

	return &model.GetListCategoryResponse{Categories: data}, nil
}

func (d *categoryDomain) GetTemplate(
	ctx context.Context, req *model.GetTemplateCategoryRequest,
) (*model.GetTemplateCategoryResponse, error) {
	categoryEntities, err := d.categoryRepo.GetTemplates(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get the category list: %v", err)
		return nil, errorx.Unknown
	}

	data := []model.Category{}
	for _, e := range categoryEntities {
		data = append(data, model.ConvertCategory(&e))
	}

	return &model.GetTemplateCategoryResponse{Categories: data}, nil
}

func (d *categoryDomain) UpdateByID(ctx context.Context, req *model.UpdateCategoryByIDRequest) (*model.UpdateCategoryByIDResponse, error) {
	category, err := d.categoryRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found category")
		}

		xcontext.Logger(ctx).Errorf("Cannot get the category: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, category.CommunityID.String); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	if req.Position > category.Position {
		err = d.categoryRepo.DecreasePosition(ctx, category.CommunityID.String, category.Position, req.Position)
	} else if req.Position < category.Position {
		err = d.categoryRepo.IncreasePosition(ctx, category.CommunityID.String, req.Position, category.Position)
	}
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot adjust category position: %v", err)
		return nil, errorx.Unknown
	}

	err = d.categoryRepo.UpdateByID(ctx, req.ID, &entity.Category{
		Name: req.Name, Position: req.Position,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update category: %v", err)
		return nil, errorx.Unknown
	}

	ctx = xcontext.WithCommitDBTransaction(ctx)

	newCategory, err := d.categoryRepo.GetByID(ctx, req.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get new category: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateCategoryByIDResponse{Category: model.ConvertCategory(newCategory)}, nil
}

func (d *categoryDomain) DeleteByID(ctx context.Context, req *model.DeleteCategoryByIDRequest) (*model.DeleteCategoryByIDResponse, error) {
	category, err := d.categoryRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found category")
		}

		xcontext.Logger(ctx).Errorf("Cannot get the category: %v", err)
		return nil, errorx.Unknown
	}

	if err = d.roleVerifier.Verify(ctx, category.CommunityID.String); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	err = d.questRepo.RemoveQuestCategory(ctx, category.CommunityID.String, category.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot remove category from quests: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.categoryRepo.DeleteByID(ctx, req.ID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete category: %v", err)
		return nil, errorx.Unknown
	}

	err = d.categoryRepo.DecreasePosition(ctx, category.CommunityID.String, category.Position, -1)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot adjust category position: %v", err)
		return nil, errorx.Unknown
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.DeleteCategoryByIDResponse{}, nil
}
