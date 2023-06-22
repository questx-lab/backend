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
	events, err := job.gameRepo.GetAvailableLuckyboxEvent(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get events: %v", err)
		return
	}

	for _, event := range events {
		room, err := job.gameRepo.GetRoomByID(ctx, event.RoomID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get room: %v", err)
			continue
		}

		serverAction := model.GameActionServerRequest{
			UserID: "",
			Type:   gameengine.CreateLuckyboxEventAction{}.Type(),
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
			xcontext.Logger(ctx).Errorf("cannot mark event %s as started: %v", event.ID, err)
			continue
		}

		err = job.publisher.Publish(ctx, room.StartedBy, &pubsub.Pack{Key: []byte(room.ID), Msg: b})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot publish create event %s: %v", event.ID, err)
			continue
		}

		xcontext.Logger(ctx).Infof("Start event %s of room %s successfully", event.ID, room.ID)
	}
}

func (job *LuckyboxEventCronJob) RunNow() bool {
	return false
}

func (job *LuckyboxEventCronJob) Next() time.Time {
	return time.Now().Add(time.Minute).Truncate(time.Minute)
}
