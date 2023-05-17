package model

type CreateProjectRequest struct {
	Name               string   `json:"name"`
	Introduction       string   `json:"introduction"`
	WebsiteURL         string   `json:"website_url"`
	DevelopmentStage   string   `json:"development_stage"`
	TeamSize           int      `json:"team_size"`
	SharedContentTypes []string `json:"shared_content_types"`
	Twitter            string   `json:"twitter"`
	ReferralCode       string   `json:"referral_code"`
}

type CreateProjectResponse struct {
	ID string `json:"id"`
}

type GetListProjectRequest struct {
	Q      string `json:"q"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

type GetListProjectResponse struct {
	Projects []Project `json:"projects"`
}

type GetProjectByIDRequest struct {
	ID string `json:"id"`
}

type GetProjectByIDResponse struct {
	Project `json:"project"`
}

type UpdateProjectByIDRequest struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Introduction       string   `json:"introduction"`
	WebsiteURL         string   `json:"website_url"`
	DevelopmentStage   string   `json:"development_stage"`
	TeamSize           int      `json:"team_size"`
	SharedContentTypes []string `json:"shared_content_types"`
	Twitter            string   `json:"twitter"`
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
	Projects []Project `json:"projects"`
}

type UploadProjectLogoRequest struct {
	// Logo data is included in form-data.
}

type UploadProjectLogoResponse struct{}

type GetMyReferralProjectsRequest struct{}

type GetMyReferralProjectsResponse struct {
	Projects []Project `json:"projects"`
}

type GetPendingReferralProjectsRequest struct{}

type GetPendingReferralProjectsResponse struct {
	Projects []Project `json:"projects"`
}

type ApproveReferralProjectsRequest struct {
	ProjectIDs []string `json:"project_ids"`
}

type ApproveReferralProjectsResponse struct{}
