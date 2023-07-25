package model

type ServeGameClientRequest struct {
	RoomID string `json:"room_id"`
}

type ServeGameProxyRequest struct {
	RoomID  string `json:"room_id"`
	ProxyID string `json:"proxy_id"`
}

type EnginePingCenterResponse struct{}

type GameActionServerRequest struct {
	UserID string         `json:"user_id"`
	Type   string         `json:"type"`
	Value  map[string]any `json:"value"`
}

type GameActionClientRequest struct {
	Type  string         `json:"type"`
	Value map[string]any `json:"value"`
}

type GameActionServerResponse struct {
	UserID string         `json:"user_id"`
	To     []string       `json:"to"`
	Type   string         `json:"type"`
	Value  map[string]any `json:"value"`
}

type GameActionClientResponse struct {
	UserID string         `json:"user_id"`
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

func ServerActionToClientAction(resp GameActionServerResponse) GameActionClientResponse {
	return GameActionClientResponse{
		UserID: resp.UserID,
		Type:   resp.Type,
		Value:  resp.Value,
	}
}
