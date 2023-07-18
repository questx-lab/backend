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

type ChatMessageReactionStatisticRepository interface {
	Create(ctx context.Context, data *entity.ChatMessageReactionStatistic) error
	Increment(ctx context.Context, messageID, id string) error
	GetListByMessages(ctx context.Context, messageIDs []string) ([]*entity.ChatMessageReactionStatistic, error)
}

type chatMessageReactionStatisticRepository struct {
	session gocqlx.Session
	tbl     *table.Table
}

func NewChatMessageReactionStatisticRepository(session gocqlx.Session) ChatMessageReactionStatisticRepository {
	e := &entity.ChatMessageReactionStatistic{}
	m := table.Metadata{
		Name:    e.TableName(),
		Columns: gocqlutil.GetTableNames(e),
		PartKey: []string{"message_id"},
		SortKey: []string{"reaction_id"},
	}

	return &chatMessageReactionStatisticRepository{session: session, tbl: table.New(m)}
}

func (r *chatMessageReactionStatisticRepository) Create(ctx context.Context, data *entity.ChatMessageReactionStatistic) error {
	if err := gocqlutil.Insert(r.session, r.tbl, data); err != nil {
		return err
	}

	return nil
}
func (r *chatMessageReactionStatisticRepository) Increment(ctx context.Context, messageID, id string) error {
	e := &entity.ChatMessageReactionStatistic{
		MessageID:  messageID,
		ReactionID: id,
	}
	stmt := fmt.Sprintf(`
		UPDATE %s
		SET quantity = quantity + 1
		WHERE message_id = $1 AND reaction_id = $2
	`, e.TableName())
	_, names := r.tbl.Update()
	err := gocqlx.Session.Query(r.session, stmt, names).BindStruct(e).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func (r *chatMessageReactionStatisticRepository) GetListByMessages(ctx context.Context, messageIDs []string) ([]*entity.ChatMessageReactionStatistic, error) {
	var result []*entity.ChatMessageReactionStatistic
	metadata := r.tbl.Metadata()

	stmt, names := qb.Select(metadata.Name).Columns(metadata.Columns...).Where(qb.In("message_id")).ToCql()
	err := gocqlx.Session.Query(r.session, stmt,
		names).BindStruct(map[string]any{
		"message_id": messageIDs,
	}).GetRelease(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
