package model

type GetFollowerRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetFollowerResponse Follower

type GetAllMyFollowersRequest struct{}

type GetAllMyFollowersResponse struct {
	Followers []Follower `json:"followers"`
}

type GetFollowersRequest struct {
	CommunityHandle string `json:"community_handle"`
	Q               string `json:"q"`
	IgnoreUserRole  bool   `json:"ignore_user_role"`
	Offset          int    `json:"offset"`
	Limit           int    `json:"limit"`
}

type GetFollowersResponse struct {
	Followers []Follower `json:"followers"`
}

type SearchMentionRequest struct {
	CommunityHandle string `json:"community_handle"`
	Q               string `json:"q"`
	Cursor          uint64 `json:"cursor"`
	Limit           int    `json:"limit"`
}

type SearchMentionResponse struct {
	Users      []ShortUser `json:"users"`
	NextCursor uint64      `json:"next_cursor"`
}

type GetStreaksRequest struct {
	CommunityHandle string `json:"community_handle"`
	Begin           string `json:"begin"`
	End             string `json:"end"`
}

type GetStreaksResponse struct {
	Records []FollowerStreak `json:"records"`
}
