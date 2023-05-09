package model

type CreateCollaboratorRequest struct {
	ProjectID string `json:"project_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Role      string `json:"name,omitempty"`
}

type CreateCollaboratorResponse struct {
	ID string `json:"id,omitempty"`
}

type GetListCollaboratorRequest struct {
	ProjectID string `json:"project_id"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
}

type GetListCollaboratorResponse struct {
	Collaborators []Collaborator `json:"collaborators,omitempty"`
}

type GetCollaboratorRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}

type GetCollaboratorResponse struct {
	Collaborator
}

type UpdateCollaboratorRoleRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
}

type UpdateCollaboratorRoleResponse struct {
}

type DeleteCollaboratorRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}

type DeleteCollaboratorResponse struct{}
