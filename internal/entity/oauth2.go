package entity

import "github.com/google/uuid"

type OAuth2 struct {
	UserID        uuid.UUID
	Service       string
	ServiceUserID string
}
