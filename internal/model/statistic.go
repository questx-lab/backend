package model

type GetLeaderBoardRequest struct {
	Range     string `json:"range"`
	ProjectID string `json:"project_id"`

	Type string `json:"type"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}
type Achievement struct {
	UserID    string `json:"user_id"`
	TotalTask int64  `json:"total_task"`
	TotalExp  int64  `json:"total_exp"`
}

type GetLeaderBoardResponse struct {
	Data []Achievement `json:"data"`
}
