package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/scylladb/gocqlx/qb"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/table"
)

type ChatReactionStatisticRepository interface {
	IncreaseCount(ctx context.Context, messageID int64, emoji entity.Emoji) error
	GetListByMessages(ctx context.Context, messageIDs []int64) ([]entity.ChatReactionStatistic, error)
}

type chatReactionStatisticRepository struct {
	session gocqlx.Session
	tbl     *table.Table
}

func NewChatMessageReactionStatisticRepository(session gocqlx.Session) ChatReactionStatisticRepository {
	e := &entity.ChatReactionStatistic{}
	m := table.Metadata{
		Name:    e.TableName(),
		Columns: reflectutil.GetColumnNames(e),
		PartKey: []string{"message_id"},
		SortKey: []string{"emoji"},
	}

	return &chatReactionStatisticRepository{
		session: session,
		tbl:     table.New(m),
	}
}

func (r *chatReactionStatisticRepository) IncreaseCount(ctx context.Context, messageID int64, emoji entity.Emoji) error {
	stmt, names := r.tbl.UpdateBuilder().Add("count").ToCql()
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(1, messageID, emoji).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func (r *chatReactionStatisticRepository) GetListByMessages(
	ctx context.Context, messageIDs []int64,
) ([]entity.ChatReactionStatistic, error) {
	var result []entity.ChatReactionStatistic
	metadata := r.tbl.Metadata()

	stmt, names := qb.Select(metadata.Name).Columns(metadata.Columns...).
		Where(qb.In("message_id")).ToCql()
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(messageIDs).SelectRelease(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
