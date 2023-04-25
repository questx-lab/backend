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

type DeleteMapRequest struct {
	ID string `json:"id,omitempty"`
}

type DeleteMapResponse struct {
}

type DeleteRoomRequest struct {
	ID string `json:"id,omitempty"`
}

type DeleteRoomResponse struct {
}

type GetMapInfoRequest struct {
	RoomID string `json:"room_id"`
}

type GetMapInfoResponse struct {
	MapPath        string `json:"map_path,omitempty"`
	TilesetPath    string `json:"tileset_path,omitempty"`
	PlayerImgPath  string `json:"player_img_path,omitempty"`
	PlayerJsonPath string `json:"player_json_path,omitempty"`
}
