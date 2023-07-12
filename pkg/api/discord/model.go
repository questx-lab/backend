package discord

import "time"

const AdministratorRoleFlag = 8

type Guild struct {
	ID      string
	OwnerID string
}

type Role struct {
	ID          string
	Name        string
	Position    int
	BotID       string
	Permissions int
}

type User struct {
	ID string
}

type Member struct {
	User
	RoleIDs []string
}

type InviteCode struct {
	Code      string
	Uses      int
	MaxUses   int
	MaxAge    time.Duration
	CreatedAt time.Time
	Inviter   User
}
