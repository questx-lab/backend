package event

import "github.com/questx-lab/backend/internal/model"

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
	User model.User `json:"user"`
}

func (ChangeUserStatusEvent) Op() string {
	return "change_status"
}
