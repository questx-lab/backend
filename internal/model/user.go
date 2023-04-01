package model

type GetUserRequest struct{}

type GetUserResponse struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Name    string `json:"name"`
}

type JoinProjectRequest struct {
	ProjectID string `json:"project_id"`
}
type JoinProjectResponse struct{}

type GetPointsRequest struct {
	ProjectID string `json:"project_id"`
}
type GetPointsResponse struct {
	Points uint64 `json:"points,omitempty"`
}
