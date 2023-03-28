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
	ProjectID string `json:"project_id,omitempty"`
	Pagination
}

type Collaborator struct {
	ID          string `json:"id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Role        string `json:"name,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	UserName    string `json:"user_name,omitempty"`
}

type GetListCollaboratorResponse struct {
	Collaborators []Collaborator `json:"collaborators,omitempty"`
}

type GetCollaboratorRequest struct {
	ProjectID string `json:"project_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

type GetCollaboratorResponse struct {
	Collaborator
}

type UpdateCollaboratorRoleRequest struct {
	ProjectID string `json:"project_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Role      string `json:"role,omitempty"`
}

type UpdateCollaboratorRoleResponse struct {
}

type DeleteCollaboratorRequest struct {
	ProjectID string `json:"project_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

type DeleteCollaboratorResponse struct{}
