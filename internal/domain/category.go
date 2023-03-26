package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/router"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type CategoryDomain interface {
	Create(ctx router.Context, req *model.CreateCategoryRequest) (*model.CreateCategoryResponse, error)
	GetList(ctx router.Context, req *model.GetListCategoryRequest) (*model.GetListCategoryResponse, error)
	GeyByID(ctx router.Context, req *model.GetCategoryByIDRequest) (*model.GetCategoryByIDResponse, error)
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
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("unable to retrieve project: %w", err)
	}

	collaborator, err := d.collaboratorRepo.GetCollaborator(ctx, req.ProjectID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user does not have permission")
		}
		return nil, fmt.Errorf("unable to retrieve project: %w", err)
	}

	if !slices.Contains([]entity.CollaboratorRole{
		entity.CollaboratorRoleOwner,
		entity.CollaboratorRoleEditor,
	}, collaborator.Role) {
		return nil, fmt.Errorf("user role does not have permission")
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
		return nil, fmt.Errorf("unable to create category: %w", err)
	}

	return &model.CreateCategoryResponse{
		Success: true,
	}, nil
}

func (d *categoryDomain) GetList(ctx router.Context, req *model.GetListCategoryRequest) (*model.GetListCategoryResponse, error) {
	categoryEntities, err := d.categoryRepo.GetList(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get list categories: %w", err)
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
	return &model.GetListCategoryResponse{
		Data:    data,
		Success: true,
	}, nil
}

func (d *categoryDomain) GeyByID(ctx router.Context, req *model.GetCategoryByIDRequest) (*model.GetCategoryByIDResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (d *categoryDomain) UpdateByID(ctx router.Context, req *model.UpdateCategoryByIDRequest) (*model.UpdateCategoryByIDResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (d *categoryDomain) DeleteByID(ctx router.Context, req *model.DeleteCategoryByIDRequest) (*model.DeleteCategoryByIDResponse, error) {
	panic("not implemented") // TODO: Implement
}
