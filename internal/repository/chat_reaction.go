package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

type ChatReactionRepository interface {
	// Reactions
	Create(ctx context.Context, data *entity.ChatReaction) error
	Get(ctx context.Context, userID string, messageID int64, emoji entity.Emoji) (*entity.ChatReaction, error)
	Delete(ctx context.Context, messageID int64, userID string, emoji entity.Emoji) error
	DeleteByMessageID(ctx context.Context, messageID int64) error
	GetList(ctx context.Context, messageID int64, emoji entity.Emoji, limit int64) ([]*entity.ChatReaction, error)
}

type chatReactionRepository struct {
	session gocqlx.Session
	tbl     *table.Table
}

func NewChatReactionRepository(session gocqlx.Session) ChatReactionRepository {
	e := &entity.ChatReaction{}
	m := table.Metadata{
		Name:    e.TableName(),
		Columns: reflectutil.GetColumnNames(e),
		PartKey: []string{"message_id", "emoji"},
		SortKey: []string{"user_id"},
	}

	return &chatReactionRepository{
		session: session,
		tbl:     table.New(m),
	}
}

func (r *chatReactionRepository) Create(ctx context.Context, data *entity.ChatReaction) error {
	stmt, names := r.tbl.Insert()
	err := gocqlx.Session.Query(r.session, stmt, names).BindStruct(data).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func (r *chatReactionRepository) Get(
	ctx context.Context, userID string, messageID int64, emoji entity.Emoji,
) (*entity.ChatReaction, error) {
	var result entity.ChatReaction
	stmt, names := r.tbl.Get()
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(messageID, emoji, userID).GetRelease(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *chatReactionRepository) Delete(ctx context.Context, messageID int64, userID string, emoji entity.Emoji) error {
	stmt, names := r.tbl.Delete()
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(messageID, emoji, userID).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func (r *chatReactionRepository) GetList(ctx context.Context, messageID int64, emoji entity.Emoji, limit int64) ([]*entity.ChatReaction, error) {
	var result []*entity.ChatReaction
	stmt, names := r.tbl.SelectBuilder().Limit(uint(limit)).ToCql()
	err := gocqlx.Session.Query(r.session, stmt, names).BindStruct(&entity.ChatReaction{
		MessageID: messageID,
		Emoji:     emoji,
	}).SelectRelease(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *chatReactionRepository) DeleteByMessageID(ctx context.Context, messageID int64) error {
	stmt, names := r.tbl.DeleteBuilder().Where(qb.Eq("message_id")).ToCql()
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(&entity.ChatReaction{
		MessageID: messageID,
	}).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}
