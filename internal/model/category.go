package model

type CreateCategoryRequest struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
}

type CreateCategoryResponse struct {
	ID string `json:"id,omitempty"`
}

type GetListCategoryRequest struct {
	ProjectID string `json:"project_id"`
}

type GetListCategoryResponse struct {
	Categories []Category `json:"categories,omitempty"`
}

type GetCategoryByIDRequest struct {
	ID string `json:"id"`
}

type GetCategoryByIDResponse struct {
	Category
}

type UpdateCategoryByIDRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateCategoryByIDResponse struct{}

type DeleteCategoryByIDRequest struct {
	ID string `json:"id"`
}

type DeleteCategoryByIDResponse struct{}
