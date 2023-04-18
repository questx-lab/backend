package domain

import (
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GameClientDomain interface {
	WSServe(ctx xcontext.Context, conn *ws.Connection, req *model.WSGameClientRequest) error
}

type gameClientDomain struct {
	gameRepo repository.GameRepository
}

func NewGameClientDomain(gameRepo repository.GameRepository) *gameClientDomain {
	return &gameClientDomain{
		gameRepo: gameRepo,
	}
}

func (d *gameClientDomain) WSServe(
	ctx xcontext.Context, conn *ws.Connection, req *model.WSGameClientRequest,
) error {
	room, err := d.gameRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		ctx.Logger().Debugf("Cannot get room with id=%s", req.RoomID)
		return errorx.New(errorx.BadRequest, "invalid room")
	}

	gameMap, err := d.gameRepo.GetMapByID(ctx, room.MapID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get map with id=%s and room_id=%s", room.MapID, room.ID)
		return errorx.Unknown
	}

	// Send the map content to client.
	err = conn.Write(gameMap.Content)
	if err != nil {
		ctx.Logger().Errorf("Cannot send game map to client")
		return errorx.Unknown
	}

	for {
		select {
		case message, ok := <-conn.R:
			if !ok {
				// The connection is closed or cannot read from connection anymore.
				return nil
			}

			// TODO: Currently, it sends what it receives.
			err := conn.Write(message)
			if err != nil {
				return err
			}
		}
	}
}
