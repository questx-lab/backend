package model

type CreateProjectRequest struct {
	Name         string `json:"name,omitempty"`
	Introduction string `json:"introduction,omitempty"`
	Twitter      string `json:"twitter,omitempty"`
}

type CreateProjectResponse struct {
	ID string `json:"id,omitempty"`
}

type GetListProjectRequest struct {
	Q      string `json:"q"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

type GetListProjectResponse struct {
	Projects []Project `json:"projects,omitempty"`
}

type GetProjectByIDRequest struct {
	ID string `json:"id"`
}

type GetProjectByIDResponse struct {
	Project `json:"project,omitempty"`
}

type UpdateProjectByIDRequest struct {
	ID           string `json:"id"`
	Introduction string `json:"introduction"`
	Twitter      string `json:"twitter"`
}

type UpdateProjectByIDResponse struct{}

type UpdateProjectDiscordRequest struct {
	ID          string `json:"id"`
	ServerID    string `json:"server_id"`
	AccessToken string `json:"access_token"`
}

type UpdateProjectDiscordResponse struct{}

type DeleteProjectByIDRequest struct {
	ID string `json:"id"`
}

type DeleteProjectByIDResponse struct{}

type GetFollowingProjectRequest struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type GetFollowingProjectResponse struct {
	Projects []Project `json:"projects,omitempty"`
}
