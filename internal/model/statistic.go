package model

type GetLeaderBoardRequest struct {
	Range     string `json:"range"`
	ProjectID string `json:"project_id"`

	Type string `json:"type"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}
type Achievement struct {
	UserID     string `json:"user_id"`
	TotalTask  uint64 `json:"total_task"`
	TotalPoint uint64 `json:"total_point"`
}

type GetLeaderBoardResponse struct {
	Data []Achievement `json:"data"`
}
