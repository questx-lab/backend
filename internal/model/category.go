package model

type CreateCategoryRequest struct {
	CommunityHandle string `json:"community_handle"`
	Name            string `json:"name"`
}

type CreateCategoryResponse struct {
	Category Category `json:"category"`
}

type GetListCategoryRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetListCategoryResponse struct {
	Categories []Category `json:"categories"`
}

type GetTemplateCategoryRequest struct {
}

type GetTemplateCategoryResponse struct {
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

type UpdateCategoryByIDResponse struct {
	Category Category `json:"category"`
}

type DeleteCategoryByIDRequest struct {
	ID string `json:"id"`
}

type DeleteCategoryByIDResponse struct{}
