package model

type CreateGameMapRequest struct {
	// This request includes these following fields in form-data format:
	// name: text
	// init_x: integer
	// init_y: integer
	// map: application/json
	// collision_layers: text separated by colon
}

type CreateGameMapResponse struct {
	ID string `json:"id"`
}

type UpdateGameMapTilesetRequest struct {
	// This request includes these following fields in form-data format:
	// game_map_id: text
	// tileset: application/png
}

type UpdateGameMapTilesetResponse struct{}

type UpdateGameMapPlayerRequest struct {
	// This request includes these following fields in form-data format:
	// game_map_id: text
	// name: text
	// player_img: application/png
	// player_cfg: application/json
}

type UpdateGameMapPlayerResponse struct{}

type CreateGameRoomRequest struct {
	CommunityHandle string `json:"community_handle"`
	MapID           string `json:"map_id"`
	Name            string `json:"name"`
}

type CreateGameRoomResponse struct {
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

type GetRoomsByCommunityRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetRoomsByCommunityResponse struct {
	Community Community  `json:"community"`
	GameRooms []GameRoom `json:"game_rooms"`
}

type GetMapsRequest struct{}

type GetMapsResponse struct {
	GameMaps []GameMap `json:"game_maps"`
}
