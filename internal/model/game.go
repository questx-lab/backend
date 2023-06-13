package model

// CreateMapRequest is a form data.
type CreateMapRequest struct{}

type CreateMapResponse struct {
	ID string `json:"id"`
}

type CreateRoomRequest struct {
	CommunityHandle string `json:"community_handle"`
	MapID           string `json:"map_id"`
	Name            string `json:"name"`
}

type CreateRoomResponse struct {
	ID string `json:"id"`
}

type DeleteMapRequest struct {
	ID string `json:"id"`
}

type DeleteMapResponse struct {
}

type DeleteRoomRequest struct {
	ID string `json:"id"`
}

type DeleteRoomResponse struct {
}

type GetRoomsRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetRoomsResponse struct {
	GameRooms []GameRoom `json:"game_rooms"`
}

type GetMapsRequest struct{}

type GetMapsResponse struct {
	GameMaps []GameMap `json:"game_maps"`
}
