package model

// CreateMapRequest is a form data.
type CreateMapRequest struct{}

type CreateMapResponse struct {
	ID string `json:"id,omitempty"`
}

type CreateRoomRequest struct {
	MapID string `json:"map_id"`
	Name  string `json:"name"`
}

type CreateRoomResponse struct {
	ID string `json:"id,omitempty"`
}
