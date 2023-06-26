package model

import "time"

type CreateGameMapRequest struct {
	// This request includes these following fields in form-data format:
	// config_file: application/json
	// id (optional): string, if exists, update the map with given id, else
	//                create a new map.
}

type CreateGameMapResponse struct {
	ID string `json:"id"`
}

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

type CreateLuckyboxEventRequest struct {
	RoomID        string        `json:"room_id"`
	NumberOfBoxes int           `json:"number_of_boxes"`
	PointPerBox   int           `json:"point_per_box"`
	IsRandom      bool          `json:"is_random"`
	StartTime     time.Time     `json:"start_time"`
	Duration      time.Duration `json:"duration"`
}

type CreateLuckyboxEventResponse struct{}
