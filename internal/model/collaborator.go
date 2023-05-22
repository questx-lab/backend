package model

type AssignCollaboratorRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"name"`
}

type AssignCollaboratorResponse struct{}

type GetProjectCollabsRequest struct {
	ProjectID string `json:"project_id"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
}

type GetProjectCollabsResponse struct {
	Collaborators []Collaborator `json:"collaborators"`
}

type GetCollaboratorRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}

type GetCollaboratorResponse struct {
	Collaborator
}

type DeleteCollaboratorRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}

type DeleteCollaboratorResponse struct{}

type GetMyCollabsRequest struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type GetMyCollabsResponse struct {
	Collaborators []Collaborator `json:"collaborators"`
}
