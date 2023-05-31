package model

type AssignCollaboratorRequest struct {
	CommunityHandle string `json:"community_handle"`
	UserID          string `json:"user_id"`
	Role            string `json:"name"`
}

type AssignCollaboratorResponse struct{}

type GetCommunityCollabsRequest struct {
	CommunityHandle string `json:"community_handle"`
	Offset          int    `json:"offset"`
	Limit           int    `json:"limit"`
}

type GetCommunityCollabsResponse struct {
	Collaborators []Collaborator `json:"collaborators"`
}

type DeleteCollaboratorRequest struct {
	CommunityHandle string `json:"community_handle"`
	UserID          string `json:"user_id"`
}

type DeleteCollaboratorResponse struct{}

type GetMyCollabsRequest struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type GetMyCollabsResponse struct {
	Collaborators []Collaborator `json:"collaborators"`
}
