package model

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
	ID string `json:"id,omitempty"`
}

type GetListProjectRequest struct {
	Pagination
}

type Project struct {
	ID        string
	CreatedAt string
	UpdatedAt string

	CreatedBy string
	Name      string
	Twitter   string
	Discord   string
	Telegram  string
}

type GetListProjectResponse struct {
	Projects []Project `json:"projects,omitempty"`
}

type GetProjectByIDRequest struct {
	ID string `json:"id,omitempty"`
}

type GetProjectByIDResponse struct {
	Project
}

type UpdateProjectByIDRequest struct {
	ID       string `json:"id,omitempty"`
	Twitter  string `json:"twitter,omitempty"`
	Discord  string `json:"discord,omitempty"`
	Telegram string `json:"telegram,omitempty"`
}

type UpdateProjectByIDResponse struct{}

type DeleteProjectByIDRequest struct {
	ID string `json:"id,omitempty"`
}

type DeleteProjectByIDResponse struct{}
