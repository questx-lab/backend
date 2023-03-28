package domain

import (
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type CategoryDomain interface {
	Create(ctx router.Context, req *model.CreateCategoryRequest) (*model.CreateCategoryResponse, error)
	GetList(ctx router.Context, req *model.GetListCategoryRequest) (*model.GetListCategoryResponse, error)
	UpdateByID(ctx router.Context, req *model.UpdateCategoryByIDRequest) (*model.UpdateCategoryByIDResponse, error)
	DeleteByID(ctx router.Context, req *model.DeleteCategoryByIDRequest) (*model.DeleteCategoryByIDResponse, error)
}

type categoryDomain struct {
	categoryRepo     repository.CategoryRepository
	projectRepo      repository.ProjectRepository
	collaboratorRepo repository.CollaboratorRepository
}

func NewCategoryDomain(
	categoryRepo repository.CategoryRepository,
	projectRepo repository.ProjectRepository,
	collaboratorRepo repository.CollaboratorRepository,
) CategoryDomain {
	return &categoryDomain{
		categoryRepo:     categoryRepo,
		projectRepo:      projectRepo,
		collaboratorRepo: collaboratorRepo,
	}
}

func (d *categoryDomain) Create(ctx router.Context, req *model.CreateCategoryRequest) (*model.CreateCategoryResponse, error) {
	userID := ctx.GetUserID()

	if _, err := d.projectRepo.GetByID(ctx, req.ProjectID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found project")
		}

		ctx.Logger().Errorf("Cannot get the project: %v", err)
		return nil, errorx.Unknown
	}

	if reason := verifyProjectPermission(ctx, d.collaboratorRepo, req.ProjectID); reason != "" {
		return nil, errorx.New(errorx.PermissionDenied, reason)
	}

	e := &entity.Category{
		Base: entity.Base{
			ID: uuid.NewString(),
		},
		ProjectID: req.ProjectID,
		Name:      req.Name,
		CreatedBy: userID,
	}

	if err := d.categoryRepo.Create(ctx, e); err != nil {
		ctx.Logger().Errorf("Cannot create category: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateCategoryResponse{ID: e.ID}, nil
}

func (d *categoryDomain) GetList(ctx router.Context, req *model.GetListCategoryRequest) (*model.GetListCategoryResponse, error) {
	categoryEntities, err := d.categoryRepo.GetList(ctx)
	if err != nil {
		ctx.Logger().Errorf("Cannot get the category list: %v", err)
		return nil, errorx.Unknown
	}

	var data []*model.Category
	for _, e := range categoryEntities {
		data = append(data, &model.Category{
			ID:          e.ID,
			Name:        e.Name,
			Description: e.Description,
			ProjectID:   e.Project.ID,
			ProjectName: e.Project.Name,
			CreatedBy:   e.CreatedBy,
			CreatedAt:   e.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   e.UpdatedAt.Format(time.RFC3339Nano),
		})
	}

	return &model.GetListCategoryResponse{Categories: data}, nil
}

func (d *categoryDomain) UpdateByID(ctx router.Context, req *model.UpdateCategoryByIDRequest) (*model.UpdateCategoryByIDResponse, error) {
	category, err := d.categoryRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found category")
		}

		ctx.Logger().Errorf("Cannot get the category: %v", err)
		return nil, errorx.Unknown
	}

	if reason := verifyProjectPermission(ctx, d.collaboratorRepo, category.ProjectID); reason != "" {
		return nil, errorx.New(errorx.PermissionDenied, reason)
	}

	if err := d.categoryRepo.UpdateByID(ctx, req.ID, &entity.Category{}); err != nil {
		ctx.Logger().Errorf("Cannot update category: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateCategoryByIDResponse{}, nil
}

func (d *categoryDomain) DeleteByID(ctx router.Context, req *model.DeleteCategoryByIDRequest) (*model.DeleteCategoryByIDResponse, error) {
	category, err := d.categoryRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found category")
		}

		ctx.Logger().Errorf("Cannot get the category: %v", err)
		return nil, errorx.Unknown
	}

	if reason := verifyProjectPermission(ctx, d.collaboratorRepo, category.ProjectID); reason != "" {
		return nil, errorx.New(errorx.PermissionDenied, reason)
	}

	if err := d.categoryRepo.DeleteByID(ctx, req.ID); err != nil {
		ctx.Logger().Errorf("Cannot delete category: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteCategoryByIDResponse{}, nil
}

func verifyProjectPermission(
	ctx router.Context,
	collaboratorRepo repository.CollaboratorRepository,
	projectID string,
) string {
	userID := ctx.GetUserID()

	collaborator, err := collaboratorRepo.GetCollaborator(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "User does not have permission"
		}

		ctx.Logger().Errorf("Cannot get the collaborator: %v", err)
		return errorx.Unknown.Message
	}

	if !slices.Contains([]entity.Role{
		entity.Owner,
		entity.Editor,
	}, collaborator.Role) {
		return "User role does not have permission"
	}

	return ""
}
