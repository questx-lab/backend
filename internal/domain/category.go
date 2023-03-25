package domain

import (
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/router"
)

type CategoryDomain interface {
	Create(ctx router.Context, req *model.CreateCategoryRequest) (*model.CreateCategoryResponse, error)
	GetList(ctx router.Context, req *model.GetListCategoryRequest) (*model.GetListCategoryResponse, error)
	GeyByID(ctx router.Context, req *model.GetCategoryByIDRequest) (*model.GetCategoryByIDResponse, error)
	UpdateByID(ctx router.Context, req *model.UpdateCategoryByIDRequest) (*model.UpdateCategoryByIDResponse, error)
	DeleteByID(ctx router.Context, req *model.DeleteCategoryByIDRequest) (*model.DeleteCategoryByIDResponse, error)
}

type categoryDomain struct {
	categoryRepo repository.CategoryRepository
}
