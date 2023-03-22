package model

import "github.com/questx-lab/backend/internal/entity"

type Response struct {
	Code    int    `json:"code,omitempty"`
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
}

type Pagination struct {
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`
}

type CreateProjectRequest struct {
	Name     string `json:"name,omitempty"`
	Twitter  string `json:"twitter,omitempty"`
	Discord  string `json:"discord,omitempty"`
	Telegram string `json:"telegram,omitempty"`
}

type CreateProjectResponse struct {
	Response
}

type GetListProjectRequest struct {
	Pagination
}

type GetListProjectResponse struct {
	Response
	Data []*entity.Project `json:"data,omitempty"`
}

type GetProjectByIDRequest struct {
	ID string `json:"id,omitempty"`
}

type GetProjectByIDResponse struct {
	Response
	Data *entity.Project `json:"data,omitempty"`
}

type UpdateProjectByIDRequest struct {
	ID       string `json:"id,omitempty"`
	Twitter  string `json:"twitter,omitempty"`
	Discord  string `json:"discord,omitempty"`
	Telegram string `json:"telegram,omitempty"`
}

type UpdateProjectByIDResponse struct {
	Response
}

type DeleteProjectByIDRequest struct {
	ID string `json:"id,omitempty"`
}

type DeleteProjectByIDResponse struct {
	Response
}
