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
}

type Collaborator struct {
	ID        string `json:"id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Role      string `json:"name,omitempty"`
}

type GetListCollaboratorResponse struct {
	Data    []*Collaborator `json:"data,omitempty"`
	Success bool            `json:"success,omitempty"`
}

type GetCollaboratorByIDRequest struct {
	ID string `json:"id,omitempty"`
}

type GetCollaboratorByIDResponse struct {
	Data    *Collaborator `json:"data,omitempty"`
	Success bool          `json:"success,omitempty"`
}

type UpdateCollaboratorByIDRequest struct {
	ID   string `json:"id,omitempty"`
	Role string `json:"role,omitempty"`
}

type UpdateCollaboratorByIDResponse struct {
	Success bool `json:"success,omitempty"`
}

type DeleteCollaboratorByIDRequest struct {
	ID string `json:"id,omitempty"`
}

type DeleteCollaboratorByIDResponse struct {
	Success bool `json:"success,omitempty"`
}
