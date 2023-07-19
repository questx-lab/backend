package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/numberutil"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

type LastMessageFilter struct {
	FromBucket    int64
	ToBucket      int64
	ChannelID     int64
	LastMessageID int64
	Limit         int64
}

type ChatMessageRepository interface {
	Create(ctx context.Context, data *entity.ChatMessage) error
	Get(ctx context.Context, id, channelID int64) (*entity.ChatMessage, error)
	UpdateByID(ctx context.Context, id, channelID int64, content string, attachments []entity.Attachment) error
	Delete(ctx context.Context, channelID, bucket, id int64) error
	GetListByLastMessage(ctx context.Context, filter LastMessageFilter) ([]entity.ChatMessage, error)
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
	bucket := numberutil.BucketFrom(id)
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
	bucket := numberutil.BucketFrom(id)
	stmt, names := r.tbl.Update("content", "attachments")
	err := gocqlx.Session.Query(r.session, stmt, names).
		Bind(content, attachments, channelID, bucket, id).
		ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func (r *chatMessageRepository) Delete(ctx context.Context, channelID, bucket, id int64) error {
	stmt, names := r.tbl.Delete()
	if err := gocqlx.Session.Query(r.session, stmt, names).BindStruct(&entity.ChatMessage{
		ID:        id,
		Bucket:    bucket,
		ChannelID: channelID,
	}).ExecRelease(); err != nil {
		return err
	}

	return nil
}

func (r *chatMessageRepository) GetListByLastMessage(ctx context.Context, filter LastMessageFilter) ([]entity.ChatMessage, error) {
	builder := r.tbl.SelectBuilder().LimitNamed("limit")
	if filter.LastMessageID != 0 {
		builder = builder.Where(qb.Lt("id"))
	}
	stmt, names := builder.ToCql()
	query := gocqlx.Session.Query(r.session, stmt, names)

	var result []entity.ChatMessage
	for bucket := filter.FromBucket; bucket >= filter.ToBucket; bucket-- {
		binder := map[string]any{
			"channel_id": filter.ChannelID,
			"bucket":     bucket,
			"limit":      filter.Limit,
		}
		if filter.LastMessageID != 0 {
			binder["id"] = filter.LastMessageID
		}

		var messages []entity.ChatMessage
		if err := query.BindMap(binder).Select(&messages); err != nil {
			return nil, err
		}

		result = append(result, messages...)
		if len(messages) >= int(filter.Limit) {
			break
		}

		filter.Limit -= int64(len(messages))
	}

	query.Release()

	return result, nil
}
