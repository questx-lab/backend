package event

import (
	"encoding/json"

	"github.com/questx-lab/backend/internal/model"
)

type FollowCommunityEvent struct {
	CommunityID     string `json:"community_id"`
	CommunityHandle string `json:"community_handle"`
}

func (FollowCommunityEvent) Op() string {
	return "follow_community"
}

func (e *FollowCommunityEvent) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}

type UserStatus string

const (
	Online  = UserStatus("online")
	Offline = UserStatus("offline")
)

type ChangeUserStatusEvent struct {
	User model.ShortUser `json:"user"`
}

func (ChangeUserStatusEvent) Op() string {
	return "change_status"
}

func (e *ChangeUserStatusEvent) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}
