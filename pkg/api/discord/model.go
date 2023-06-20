package discord

import "time"

type Guild struct {
	ID      string
	OwnerID string
}

type Role struct {
	ID       string
	Name     string
	Position int
	BotID    string
}

type User struct {
	ID    string
	Roles []string
}

type InviteCode struct {
	Code      string
	Uses      int
	MaxUses   int
	MaxAge    time.Duration
	CreatedAt time.Time
	Inviter   User
}
