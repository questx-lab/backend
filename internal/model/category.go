package model

type CreateCategoryRequest struct {
	ProjectID string `json:"project_id,omitempty"`
	Name      string `json:"name,omitempty"`
}

type CreateCategoryResponse struct {
	ID      string `json:"id,omitempty"`
	Success bool   `json:"success,omitempty"`
}

type GetListCategoryRequest struct {
	ProjectID string `json:"project_id,omitempty"`
}

type Category struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}
type GetListCategoryResponse struct {
	Data    []*Category `json:"data,omitempty"`
	Success bool        `json:"success,omitempty"`
}

type GetCategoryByIDRequest struct {
	ID string `json:"id,omitempty"`
}

type GetCategoryByIDResponse struct {
	Data    *Category `json:"data,omitempty"`
	Success bool      `json:"success,omitempty"`
}

type UpdateCategoryByIDRequest struct {
	ID string `json:"id,omitempty"`
}

type UpdateCategoryByIDResponse struct {
	Success bool `json:"success,omitempty"`
}

type DeleteCategoryByIDRequest struct {
	ID string `json:"id,omitempty"`
}

type DeleteCategoryByIDResponse struct {
	Success bool `json:"success,omitempty"`
}
