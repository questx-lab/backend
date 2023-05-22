package model

type CreateCategoryRequest struct {
	CommunityID string `json:"community_id"`
	Name        string `json:"name"`
}

type CreateCategoryResponse struct {
	ID string `json:"id"`
}

type GetListCategoryRequest struct {
	CommunityID string `json:"community_id"`
}

type GetListCategoryResponse struct {
	Categories []Category `json:"categories"`
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
