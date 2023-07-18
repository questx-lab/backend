package repository

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	gocqlutil "github.com/questx-lab/backend/pkg/cqlutil"

	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

type ChatMessageRepository interface {
	CreateMessage(ctx context.Context, data *entity.ChatMessage) error
	DeleteMessage(ctx context.Context, id string, bucket int64) error
	GetListByLastMessage(ctx context.Context, filter *LastMessageFilter) ([]*entity.ChatMessage, error)
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
		PartKey: []string{"channel_id", "bucket"},
		SortKey: []string{"id"},
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

func (r *chatMessageRepository) DeleteMessage(ctx context.Context, id string, bucket int64) error {
	if err := gocqlutil.Delete(r.session, r.tbl, &entity.ChatMessage{
		ID:     id,
		Bucket: bucket,
	}); err != nil {
		return err
	}

	return nil
}

type LastMessageFilter struct {
	ChannelID     string
	LastMessageID string
	FromBucket    int64
	ToBucket      int64
	Limit         int64
}

func (r *chatMessageRepository) GetListByLastMessage(ctx context.Context, filter *LastMessageFilter) ([]*entity.ChatMessage, error) {
	cmp := []qb.Cmp{qb.Eq("channel_id"), qb.Eq("bucket"), qb.Lt("message_id")}

	var result []*entity.ChatMessage

	for bucket := filter.FromBucket; bucket >= filter.ToBucket; bucket-- {
		messages, err := gocqlutil.Select(r.session, r.tbl, &entity.ChatMessage{
			ChannelID: filter.ChannelID,
			Message:   filter.LastMessageID,
			Bucket:    bucket,
		}, filter.Limit, cmp...)
		if err != nil {
			return nil, err
		}

		result = append(result, messages...)
		result = result[:filter.Limit]
		if len(result) == int(filter.Limit) {
			break
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("channel not found")
	}

	return result, nil
}
