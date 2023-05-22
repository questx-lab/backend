package model

type GetLeaderBoardRequest struct {
	Range     string `json:"range"`
	ProjectID string `json:"project_id"`

	Type string `json:"type"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type GetLeaderBoardResponse struct {
	LeaderBoard []UserAggregate `json:"leaderboard"`
}
