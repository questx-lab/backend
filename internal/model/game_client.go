package model

type ServeGameClientRequest struct {
	RoomID string `json:"room_id"`
}

type GameActionServerRequest struct {
	UserID string         `json:"user_id"`
	Type   string         `json:"type"`
	Value  map[string]any `json:"value"`
}

type GameActionClientRequest struct {
	Type  string         `json:"type"`
	Value map[string]any `json:"value"`
}

type GameActionResponse struct {
	UserID string         `json:"user_id"`
	To     []string       `json:"-"`
	Type   string         `json:"type"`
	Value  map[string]any `json:"value"`
}

func ClientActionToServerAction(req GameActionClientRequest, userID string) GameActionServerRequest {
	return GameActionServerRequest{
		UserID: userID,
		Type:   req.Type,
		Value:  req.Value,
	}
}
