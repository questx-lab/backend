package event

type FollowCommunityEvent struct {
	CommunityID     string `json:"community_id"`
	CommunityHandle string `json:"community_handle"`
}

func (FollowCommunityEvent) Op() string {
	return "follow_community"
}
