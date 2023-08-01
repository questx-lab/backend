package event

type FollowCommunityEvent struct {
	CommunityID     string `json:"community_id"`
	CommunityHandle string `json:"community_handle"`
}

func (FollowCommunityEvent) Op() string {
	return "follow_community"
}

type UserStatus string

const (
	Online  = UserStatus("online")
	Offline = UserStatus("offline")
)

type ChangeUserStatusEvent struct {
	UserID string     `json:"user_id"`
	Status UserStatus `json:"status"`
}

func (ChangeUserStatusEvent) Op() string {
	return "change_status"
}
