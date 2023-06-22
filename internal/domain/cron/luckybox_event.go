package cron

import (
	"context"
	"encoding/json"
	"time"

	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type LuckyboxEventCronJob struct {
	gameRepo  repository.GameRepository
	publisher pubsub.Publisher
}

func NewLuckyboxEventCronJob(
	gameRepo repository.GameRepository,
	publisher pubsub.Publisher,
) *LuckyboxEventCronJob {
	return &LuckyboxEventCronJob{
		gameRepo:  gameRepo,
		publisher: publisher,
	}
}

func (job *LuckyboxEventCronJob) Do(ctx context.Context) {
	// START EVENTS.
	shouldStartEvents, err := job.gameRepo.GetShouldStartLuckyboxEvent(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get should-start events: %v", err)
		return
	}

	for _, event := range shouldStartEvents {
		room, err := job.gameRepo.GetRoomByID(ctx, event.RoomID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get room: %v", err)
			continue
		}

		if room.StartedBy == "" {
			xcontext.Logger(ctx).Errorf("Game room has not started yet")
			continue
		}

		serverAction := model.GameActionServerRequest{
			UserID: "",
			Type:   gameengine.StartLuckyboxEventAction{}.Type(),
			Value: map[string]any{
				"event_id":      event.ID,
				"amount":        event.Amount,
				"point_per_box": event.PointPerBox,
			},
		}

		b, err := json.Marshal(serverAction)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot marshal event %s: %v", event.ID, err)
			continue
		}

		err = job.gameRepo.MarkLuckyboxEventAsStarted(ctx, event.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot mark event %s as started: %v", event.ID, err)
			continue
		}

		err = job.publisher.Publish(ctx, room.StartedBy, &pubsub.Pack{Key: []byte(room.ID), Msg: b})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot publish create event %s: %v", event.ID, err)
			continue
		}

		xcontext.Logger(ctx).Infof("Start event %s of room %s successfully", event.ID, room.ID)
	}

	// STOP EVENTS.
	shouldStopEvents, err := job.gameRepo.GetShouldStopLuckyboxEvent(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get should-stop events: %v", err)
		return
	}

	for _, event := range shouldStopEvents {
		room, err := job.gameRepo.GetRoomByID(ctx, event.RoomID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get room: %v", err)
			continue
		}

		if room.StartedBy == "" {
			xcontext.Logger(ctx).Errorf("Game room has not started yet")
			continue
		}

		serverAction := model.GameActionServerRequest{
			UserID: "",
			Type:   gameengine.StopLuckyboxEventAction{}.Type(),
			Value:  map[string]any{"event_id": event.ID},
		}

		b, err := json.Marshal(serverAction)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot marshal event %s: %v", event.ID, err)
			continue
		}

		err = job.gameRepo.MarkLuckyboxEventAsStopped(ctx, event.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot mark event %s as stopped: %v", event.ID, err)
			continue
		}

		err = job.publisher.Publish(ctx, room.StartedBy, &pubsub.Pack{Key: []byte(room.ID), Msg: b})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot publish stop event %s: %v", event.ID, err)
			continue
		}

		xcontext.Logger(ctx).Infof("Stop event %s of room %s successfully", event.ID, room.ID)
	}
}

func (job *LuckyboxEventCronJob) RunNow() bool {
	return false
}

func (job *LuckyboxEventCronJob) Next() time.Time {
	return time.Now().Add(time.Minute).Truncate(time.Minute)
}
