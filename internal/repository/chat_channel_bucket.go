package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/reflectutil"

	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/table"
)

type ChatChannelBucketRepository interface {
	Increase(ctx context.Context, channelID, bucket int64) error
	Decrease(ctx context.Context, channelID, bucket int64) error
}

type chatChannelBucketRepository struct {
	session gocqlx.Session
	tbl     *table.Table
}

func NewChatBucketRepository(session gocqlx.Session) ChatChannelBucketRepository {
	e := &entity.ChatChannelBucket{}
	m := table.Metadata{
		Name:    e.TableName(),
		Columns: reflectutil.GetColumnNames(e),
		PartKey: []string{"channel_id"},
		SortKey: []string{"bucket"},
	}

	return &chatChannelBucketRepository{
		session: session,
		tbl:     table.New(m),
	}
}

func (r *chatChannelBucketRepository) Increase(ctx context.Context, channelID, bucket int64) error {
	stmt, names := r.tbl.UpdateBuilder().Add("quantity").ToCql()
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(channelID, bucket, 1).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func (r *chatChannelBucketRepository) Decrease(ctx context.Context, channelID, bucket int64) error {
	stmt, names := r.tbl.UpdateBuilder().Remove("quantity").ToCql()
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(channelID, bucket, 1).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}
