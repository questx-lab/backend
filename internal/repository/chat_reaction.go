package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocql/gocql"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/reflectutil"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
	"golang.org/x/exp/slices"
)

type ChatReactionRepository interface {
	Add(ctx context.Context, messageID int64, emoji entity.Emoji, userID string) error
	Remove(ctx context.Context, messageID int64, emoji entity.Emoji, userID string) error
	RemoveByMessageID(ctx context.Context, messageID int64) error
	CheckUserReaction(ctx context.Context, userID string, messageID int64, emoji entity.Emoji) (bool, error)
	Get(ctx context.Context, messageID int64, emoji entity.Emoji) (*entity.ChatReaction, error)
	GetByMessageID(ctx context.Context, messageID int64) ([]entity.ChatReaction, error)
	GetByMessageIDs(ctx context.Context, messageIDs []int64) ([]entity.ChatReaction, error)
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
		PartKey: []string{"message_id"},
		SortKey: []string{"emoji"},
	}

	return &chatReactionRepository{
		session: session,
		tbl:     table.New(m),
	}
}

func (r *chatReactionRepository) Add(
	ctx context.Context, messageID int64, emoji entity.Emoji, userID string,
) error {
	stmt, names := r.tbl.UpdateBuilder().AddLit("user_ids", fmt.Sprintf("{'%s'}", userID)).ToCql()
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(messageID, emoji).ExecRelease()
	if err != nil {
		return err
	}

	return nil
}

func (r *chatReactionRepository) Get(ctx context.Context, messageID int64, emoji entity.Emoji) (*entity.ChatReaction, error) {
	var result entity.ChatReaction
	stmt, names := r.tbl.Get()
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(messageID, emoji).GetRelease(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *chatReactionRepository) CheckUserReaction(
	ctx context.Context, userID string, messageID int64, emoji entity.Emoji,
) (bool, error) {
	stmt, names := qb.Select(r.tbl.Name()).
		Columns("user_ids").Where(qb.Eq("message_id"), qb.Eq("emoji")).ToCql()

	var userIDs []string
	err := gocqlx.Session.Query(r.session, stmt, names).
		BindMap(map[string]any{"message_id": messageID, "emoji": emoji}).GetRelease(&userIDs)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	return slices.Contains(userIDs, userID), nil
}

func (r *chatReactionRepository) GetByMessageID(ctx context.Context, messageID int64) ([]entity.ChatReaction, error) {
	var result []entity.ChatReaction
	stmt, names := r.tbl.Select()
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(messageID).SelectRelease(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *chatReactionRepository) GetByMessageIDs(ctx context.Context, messageIDs []int64) ([]entity.ChatReaction, error) {
	var result []entity.ChatReaction
	stmt, names := qb.Select(r.tbl.Name()).Columns(r.tbl.Metadata().Columns...).
		Where(qb.In("message_id")).ToCql()
	err := gocqlx.Session.Query(r.session, stmt, names).Bind(messageIDs).SelectRelease(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *chatReactionRepository) Remove(ctx context.Context, messageID int64, emoji entity.Emoji, userID string) error {
	stmt, names := r.tbl.UpdateBuilder().RemoveLit("user_ids", fmt.Sprintf("{'%s'}", userID)).ToCql()
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

func (r *chatReactionRepository) RemoveByMessageID(ctx context.Context, messageID int64) error {
	stmt, names := qb.Delete(r.tbl.Name()).Where(qb.Eq("message_id")).ToCql()
	if err := gocqlx.Session.Query(r.session, stmt, names).BindStruct(&entity.ChatReaction{
		MessageID: messageID,
	}).ExecRelease(); err != nil {
		return err
	}

	return nil
}
