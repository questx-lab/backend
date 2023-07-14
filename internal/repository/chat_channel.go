package repository

import (
	"context"

	"github.com/gocql/gocql"
	"github.com/questx-lab/backend/internal/entity"
	gocqlutil "github.com/questx-lab/backend/pkg/cqlutil"
	"github.com/scylladb/gocqlx/table"
)

type ChatChannelRepository interface{}

type chatChannelRepository struct {
	session *gocql.Session
	tbl     *table.Table
}

func NewChatChannelRepository(session *gocql.Session) ChatChannelRepository {
	e := &entity.ChatChannel{}
	m := table.Metadata{
		Name:    e.TableName(),
		Columns: gocqlutil.GetTableNames(e),
		PartKey: []string{"id"},
		SortKey: []string{"created_at"},
	}

	return &chatChannelRepository{
		session: session,
		tbl:     table.New(m),
	}
}

func (r *chatChannelRepository) CreateChannel(ctx context.Context, data *entity.ChatChannel) error {
	if err := gocqlutil.Insert(r.session, r.tbl, data); err != nil {
		return err
	}

	return nil
}
