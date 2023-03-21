package repository

import (
	"context"
	"database/sql"
	"log"

	"github.com/questx-lab/backend/internal/entity"
)

type OAuth2Repository interface {
	Create(ctx context.Context, data *entity.OAuth2) error
}

type oauth2Repository struct {
	db *sql.DB
}

func NewOAuth2Repository(db *sql.DB) OAuth2Repository {
	return &oauth2Repository{db: db}
}

func (r *oauth2Repository) Create(ctx context.Context, data *entity.OAuth2) error {
	log.Println("Create oauth2 row: ", data)
	return nil
}
