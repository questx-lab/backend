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
	// ID indicates the order of action when it is applied into game state.
	// Action with ID=t is only applied into game state with ID=t-1.
	ID        int            `json:"id"`
	UserID    string         `json:"user_id,omitempty"`
	OnlyOwner bool           `json:"only_owner"`
	Type      string         `json:"type,omitempty"`
	Value     map[string]any `json:"value,omitempty"`
}

func ClientActionToServerAction(req GameActionClientRequest, userID string) GameActionServerRequest {
	return GameActionServerRequest{
		UserID: userID,
		Type:   req.Type,
		Value:  req.Value,
	}
}
