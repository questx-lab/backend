package repository

import (
	"context"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm/clause"
)

type StatisticGameLuckyboxFilter struct {
	CommunityID string
	UserID      string
	StartTime   time.Time
	EndTime     time.Time
}

type GameLuckyboxRepository interface {
	CreateLuckyboxEvent(context.Context, *entity.GameLuckyboxEvent) error
	GetLuckyboxEventsHappenInRange(ctx context.Context, roomID string, start time.Time, end time.Time) ([]entity.GameLuckyboxEvent, error)
	GetShouldStartLuckyboxEvent(context.Context) ([]entity.GameLuckyboxEvent, error)
	GetShouldStopLuckyboxEvent(context.Context) ([]entity.GameLuckyboxEvent, error)
	MarkLuckyboxEventAsStarted(context.Context, string) error
	MarkLuckyboxEventAsStopped(context.Context, string) error
	UpsertLuckybox(context.Context, *entity.GameLuckybox) error
	GetAvailableLuckyboxesByRoomID(context.Context, string) ([]entity.GameLuckybox, error)
	Statistic(context.Context, StatisticGameLuckyboxFilter) ([]entity.UserStatistic, error)
}

type gameLuckyboxRepository struct{}

func NewGameLuckyboxRepository() *gameLuckyboxRepository {
	return &gameLuckyboxRepository{}
}

func (r *gameLuckyboxRepository) CreateLuckyboxEvent(ctx context.Context, event *entity.GameLuckyboxEvent) error {
	return xcontext.DB(ctx).Create(event).Error
}

func (r *gameLuckyboxRepository) GetLuckyboxEventsHappenInRange(
	ctx context.Context, roomID string, start time.Time, end time.Time,
) ([]entity.GameLuckyboxEvent, error) {
	var result []entity.GameLuckyboxEvent

	tx := xcontext.DB(ctx).Model(&entity.GameLuckyboxEvent{}).
		Where("room_id = ?", roomID)

	tx = tx.Where(
		tx.
			Or("end_time >= ? AND end_time <= ?", start, end).
			Or("start_time >= ? AND start_time <= ?", start, end).
			Or("start_time <= ? AND end_time >= ?", start, end),
	)

	if err := tx.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameLuckyboxRepository) GetShouldStartLuckyboxEvent(ctx context.Context) ([]entity.GameLuckyboxEvent, error) {
	var result []entity.GameLuckyboxEvent
	err := xcontext.DB(ctx).
		Where("start_time <= ? AND is_started=false", time.Now()).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameLuckyboxRepository) GetShouldStopLuckyboxEvent(ctx context.Context) ([]entity.GameLuckyboxEvent, error) {
	var result []entity.GameLuckyboxEvent
	err := xcontext.DB(ctx).
		Where("end_time <= ? AND is_stopped=false", time.Now()).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameLuckyboxRepository) MarkLuckyboxEventAsStarted(ctx context.Context, id string) error {
	return xcontext.DB(ctx).
		Model(&entity.GameLuckyboxEvent{}).
		Where("id=?", id).
		Update("is_started", true).Error
}

func (r *gameLuckyboxRepository) MarkLuckyboxEventAsStopped(ctx context.Context, id string) error {
	return xcontext.DB(ctx).
		Model(&entity.GameLuckyboxEvent{}).
		Where("id=?", id).
		Update("is_stopped", true).Error
}

func (r *gameLuckyboxRepository) UpsertLuckybox(ctx context.Context, luckybox *entity.GameLuckybox) error {
	return xcontext.DB(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: "id"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"collected_by": luckybox.CollectedBy,
				"collected_at": luckybox.CollectedAt,
			}),
		},
	).Create(luckybox).Error
}

func (r *gameLuckyboxRepository) GetAvailableLuckyboxesByRoomID(ctx context.Context, roomID string) ([]entity.GameLuckybox, error) {
	var result []entity.GameLuckybox
	err := xcontext.DB(ctx).Model(&entity.GameLuckybox{}).
		Joins("join game_luckybox_events on game_luckybox_events.id=game_luckyboxes.event_id").
		Where("game_luckybox_events.room_id=?", roomID).
		Where("game_luckybox_events.is_stopped=false").
		Where("game_luckyboxes.collected_by IS NULL").
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *gameLuckyboxRepository) Statistic(
	ctx context.Context, filter StatisticGameLuckyboxFilter,
) ([]entity.UserStatistic, error) {
	tx := xcontext.DB(ctx).Model(&entity.GameLuckybox{}).
		Select("SUM(game_luckyboxes.point) as points, game_rooms.community_id, users.id as user_id").
		Joins("join users on game_luckyboxes.collected_by = users.id").
		Joins("join game_luckybox_events on game_luckyboxes.event_id = game_luckybox_events.id").
		Joins("join game_rooms on game_rooms.id = game_luckybox_events.room_id").
		Group("users.id")

	if filter.CommunityID != "" {
		tx.Where("game_rooms.community_id = ?", filter.CommunityID)
	}

	if !filter.StartTime.IsZero() {
		tx.Where("game_luckyboxes.collected_at >= ?", filter.StartTime)
	}

	if !filter.EndTime.IsZero() {
		tx.Where("game_luckyboxes.collected_at <= ?", filter.EndTime)
	}

	if filter.UserID != "" {
		tx.Where("game_luckyboxes.collected_by = ?", filter.UserID)
	}

	var result []entity.UserStatistic
	if err := tx.Scan(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}
