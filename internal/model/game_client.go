package model

type ServeGameClientRequest struct {
	RoomID string `json:"room_id"`
}

type WSGameClientRequest struct {
	RoomID string `json:"room_id"`
}

type GameMapResponse struct {
	ID      string `json:"id,omitempty"`
	Content string `json:"content,omitempty"`
}

type GameUserPosition struct {
	UserID    string `json:"user_id,omitempty"`
	ObjectID  string `json:"object_id,omitempty"`
	X         int    `json:"x,omitempty"`
	Y         int    `json:"y,omitempty"`
	Direction string `json:"direction,omitempty"`
}

type GameStateResponse struct {
	// ID indicates how many actions this state applied. If client receives a
	// game state with ID=t, please ignore all actions whose id is less than or
	// equal to t.
	ID    int                `json:"id,omitempty"`
	Users []GameUserPosition `json:"users,omitempty"`
}

type GameActionRouterRequest struct {
	RoomID string         `json:"room_id"`
	UserID string         `json:"user_id"`
	Type   string         `json:"type"`
	Value  map[string]any `json:"value"`
}

type GameActionClientRequest struct {
	Type  string         `json:"type"`
	Value map[string]any `json:"value"`
}

type GameActionServerRequest struct {
	Type   string         `json:"type"`
	Value  map[string]any `json:"value"`
	UserID string         `json:"user_id,omitempty"`
}

type GameActionClientResponse struct {
	ID     int            `json:"id"`
	Type   string         `json:"type,omitempty"`
	Value  map[string]any `json:"value,omitempty"`
	UserID string         `json:"user_id"`
}

type GameActionServerResponse struct {
	// ID indicates the order of action when it is applied into game state.
	// Action with ID=t is only applied into game state with ID=t-1.
	ID     int            `json:"id,omitempty"`
	UserID string         `json:"user_id,omitempty"`
	Type   string         `json:"type,omitempty"`
	Value  map[string]any `json:"value,omitempty"`
}
