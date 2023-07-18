package repository

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	gocqlutil "github.com/questx-lab/backend/pkg/cqlutil"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/table"
)

type ChatMessageRepository interface {
	CreateMessage(ctx context.Context, data *entity.ChatMessage) error
	GetMessageByID(ctx context.Context, id int64) (*entity.ChatMessage, error)
	DeleteMessageByID(ctx context.Context, id int64) error
}

type chatMessageRepository struct {
	session gocqlx.Session
	tbl     *table.Table
}

func NewChatMessageRepository(session gocqlx.Session) ChatMessageRepository {
	e := &entity.ChatMessage{}
	m := table.Metadata{
		Name:    e.TableName(),
		Columns: gocqlutil.GetTableNames(e),
		PartKey: []string{"channel_id"},
		SortKey: []string{"created_at"},
	}

	return &chatMessageRepository{
		session: session,
		tbl:     table.New(m),
	}
}

func (r *chatMessageRepository) CreateMessage(ctx context.Context, data *entity.ChatMessage) error {
	if err := gocqlutil.Insert(r.session, r.tbl, data); err != nil {
		return err
	}

	return nil
}

func (r *chatMessageRepository) GetMessageByID(ctx context.Context, id int64) (*entity.ChatMessage, error) {
	result, err := gocqlutil.Select(r.session, r.tbl, &entity.ChatMessage{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("channel not found")
	}

	return result[0], nil
}

func (r *chatMessageRepository) DeleteMessageByID(ctx context.Context, id int64) error {
	if err := gocqlutil.Delete(r.session, r.tbl, &entity.ChatMessage{
		ID: id,
	}); err != nil {
		return err
	}

	return nil
}
