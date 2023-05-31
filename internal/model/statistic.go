package model

type GetLeaderBoardRequest struct {
	Period          string `json:"period"`
	CommunityHandle string `json:"community_handle"`
	OrderedBy       string `json:"ordered_by"`
	Offset          int    `json:"offset"`
	Limit           int    `json:"limit"`
}

type GetLeaderBoardResponse struct {
	LeaderBoard []UserStatistic `json:"leaderboard"`
}
