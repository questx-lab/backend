package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/numberutil"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/table"
)

type ChatMessageRepository interface {
	Create(ctx context.Context, data *entity.ChatMessage) error
	Get(ctx context.Context, id, channelID int64) (*entity.ChatMessage, error)
	UpdateByID(ctx context.Context, id, channelID int64, content string, attachments []entity.Attachment) error
	DeleteByID(ctx context.Context, id int64) error
}

type chatMessageRepository struct {
	session gocqlx.Session
	tbl     *table.Table
}

func NewChatMessageRepository(session gocqlx.Session) ChatMessageRepository {
	e := &entity.ChatMessage{}
	m := table.Metadata{
		Name:    e.TableName(),
		Columns: reflectutil.GetColumnNames(e),
		PartKey: []string{"channel_id", "bucket"},
		SortKey: []string{"id"},
	}

	return &chatMessageRepository{
		session: session,
		tbl:     table.New(m),
	}
}

func (r *chatMessageRepository) Create(ctx context.Context, data *entity.ChatMessage) error {
	stmt, names := r.tbl.Insert()
	err := gocqlx.Session.Query(r.session, stmt, names).BindStruct(data).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func (r *chatMessageRepository) Get(ctx context.Context, id, channelID int64) (*entity.ChatMessage, error) {
	bucket := numberutil.CreateBucket(id)
	var result entity.ChatMessage
	stmt, names := r.tbl.Get()
	err := gocqlx.Session.
		Query(r.session, stmt, names).
		Bind(channelID, bucket, id).
		GetRelease(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *chatMessageRepository) UpdateByID(
	ctx context.Context, id, channelID int64, content string, attachments []entity.Attachment,
) error {
	bucket := numberutil.CreateBucket(id)
	stmt, names := r.tbl.Update("content", "attachments")
	err := gocqlx.Session.Query(r.session, stmt, names).
		Bind(content, attachments, channelID, bucket, id).
		ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func (r *chatMessageRepository) DeleteByID(ctx context.Context, id int64) error {
	stmt, names := r.tbl.Delete()
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(id).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}
