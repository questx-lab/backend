package domain

import (
	"io"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GameDomain interface {
	CreateMap(xcontext.Context, *model.CreateMapRequest) (*model.CreateMapResponse, error)
	CreateRoom(xcontext.Context, *model.CreateRoomRequest) (*model.CreateRoomResponse, error)
}

type gameDomain struct {
	gameRepo      repository.GameRepository
	maxUploadSize int
}

func NewGameDomain(gameRepo repository.GameRepository, cfg config.FileConfigs) *gameDomain {
	return &gameDomain{gameRepo: gameRepo, maxUploadSize: cfg.MaxSize * 1024 * 1024}
}

func (d *gameDomain) CreateMap(
	ctx xcontext.Context, req *model.CreateMapRequest,
) (*model.CreateMapResponse, error) {
	if err := ctx.Request().ParseMultipartForm(int64(d.maxUploadSize)); err != nil {
		return nil, errorx.New(errorx.BadRequest, "Request must be multipart form")
	}

	tmx, _, err := ctx.Request().FormFile("tmx")
	if err != nil {
		return nil, errorx.New(errorx.BadRequest, "Cannot get the file")
	}
	defer tmx.Close()

	tmxContent, err := io.ReadAll(tmx)
	if err != nil {
		ctx.Logger().Errorf("Cannot read tmx file: %v", err)
		return nil, errorx.Unknown
	}

	name := ctx.Request().PostFormValue("name")
	if name == "" {
		return nil, errorx.New(errorx.BadRequest, "Not found map name")
	}

	gameMap := &entity.GameMap{
		Base:    entity.Base{ID: uuid.NewString()},
		Name:    name,
		Content: tmxContent,
	}

	if err := d.gameRepo.CreateMap(ctx, gameMap); err != nil {
		ctx.Logger().Errorf("Cannot create map: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateMapResponse{ID: gameMap.ID}, nil
}

func (d *gameDomain) CreateRoom(
	ctx xcontext.Context, req *model.CreateRoomRequest,
) (*model.CreateRoomResponse, error) {
	room := &entity.GameRoom{
		Base:  entity.Base{ID: uuid.NewString()},
		MapID: req.MapID,
		Name:  req.Name,
	}

	if err := d.gameRepo.CreateRoom(ctx, room); err != nil {
		ctx.Logger().Errorf("Cannot create room: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateRoomResponse{ID: room.ID}, nil
}
