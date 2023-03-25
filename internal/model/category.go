package model

type CreateCategoryRequest struct {
	ProjectID string `json:"project_id,omitempty"`
	Name      string `json:"name,omitempty"`
}

type CreateCategoryResponse struct {
	Success bool `json:"success,omitempty"`
}

type GetListCategoryRequest struct {
	ProjectID string `json:"project_id,omitempty"`
}

type GetListCategoryResponse struct {
}

type GetCategoryByIDRequest struct {
	ID string `json:"id,omitempty"`
}

type GetCategoryByIDResponse struct {
}

type UpdateCategoryByIDRequest struct {
	ID string `json:"id,omitempty"`
}

type UpdateCategoryByIDResponse struct {
}

type DeleteCategoryByIDRequest struct {
	ID string `json:"id,omitempty"`
}

type DeleteCategoryByIDResponse struct {
}
