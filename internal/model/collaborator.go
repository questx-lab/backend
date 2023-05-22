package model

type AssignCollaboratorRequest struct {
	CommunityID string `json:"community_id"`
	UserID      string `json:"user_id"`
	Role        string `json:"name"`
}

type AssignCollaboratorResponse struct{}

type GetCommunityCollabsRequest struct {
	Community string `json:"community_id"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
}

type GetCommunityCollabsResponse struct {
	Collaborators []Collaborator `json:"collaborators"`
}

type GetCollaboratorRequest struct {
	CommunityID string `json:"community_id"`
	UserID      string `json:"user_id"`
}

type GetCollaboratorResponse struct {
	Collaborator
}

type DeleteCollaboratorRequest struct {
	CommunityID string `json:"community_id"`
	UserID      string `json:"user_id"`
}

type DeleteCollaboratorResponse struct{}

type GetMyCollabsRequest struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type GetMyCollabsResponse struct {
	Collaborators []Collaborator `json:"collaborators"`
}
