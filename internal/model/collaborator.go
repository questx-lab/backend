package model

type CreateCollaboratorRequest struct {
	ProjectID string `json:"project_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Role      string `json:"name,omitempty"`
}

type CreateCollaboratorResponse struct {
	Success bool   `json:"success,omitempty"`
	ID      string `json:"id,omitempty"`
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
	Data    []*Collaborator `json:"data,omitempty"`
	Success bool            `json:"success,omitempty"`
}

type GetCollaboratorRequest struct {
	ProjectID string `json:"project_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

type GetCollaboratorResponse struct {
	Data    *Collaborator `json:"data,omitempty"`
	Success bool          `json:"success,omitempty"`
}

type UpdateCollaboratorRoleRequest struct {
	ProjectID string `json:"project_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Role      string `json:"role,omitempty"`
}

type UpdateCollaboratorRoleResponse struct {
	Success bool `json:"success,omitempty"`
}

type DeleteCollaboratorRequest struct {
	ProjectID string `json:"project_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

type DeleteCollaboratorResponse struct {
	Success bool `json:"success,omitempty"`
}
