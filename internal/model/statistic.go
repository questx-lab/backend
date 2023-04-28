package model

type GetLeaderBoardRequest struct {
	Range     string `json:"range"`
	ProjectID string `json:"project_id"`

	Type string `json:"type"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type UserAggregate struct {
	UserID      string `json:"user_id"`
	TotalTask   uint64 `json:"total_task"`
	TotalPoint  uint64 `json:"total_point"`
	PrevRank    uint64 `json:"prev_rank"`
	CurrentRank uint64 `json:"current_rank"`
}

type GetLeaderBoardResponse struct {
	Data []UserAggregate `json:"data"`
}
