package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	gocqlutil "github.com/questx-lab/backend/pkg/cqlutil"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/table"
)

type ChatMessageRepository interface {
	Create(ctx context.Context, data *entity.ChatMessage) error
	UpdateByID(ctx context.Context, id int64, newContent string, newAttachments []entity.Attachment) error
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
		Columns: gocqlutil.GetColumnNames(e),
		PartKey: []string{"channel_id"},
		SortKey: []string{"created_at"},
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

func (r *chatMessageRepository) UpdateByID(
	ctx context.Context, id int64, newContent string, newAttachments []entity.Attachment,
) error {
	stmt, names := r.tbl.Update("content", "attachments")
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(newContent, newAttachments, id).ExecRelease()
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
