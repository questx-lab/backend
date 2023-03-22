package model

type GetUserRequest struct {
	ID string `form:"id"`
}

type GetUserResponse struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Name    string `json:"name"`
}
