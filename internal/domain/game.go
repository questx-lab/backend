package domain

import (
	"errors"
	"io"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type GameDomain interface {
	CreateMap(xcontext.Context, *model.CreateMapRequest) (*model.CreateMapResponse, error)
	CreateRoom(xcontext.Context, *model.CreateRoomRequest) (*model.CreateRoomResponse, error)
	DeleteMap(xcontext.Context, *model.DeleteMapRequest) (*model.DeleteMapResponse, error)
	DeleteRoom(xcontext.Context, *model.DeleteRoomRequest) (*model.DeleteRoomResponse, error)
	GetMapInfo(xcontext.Context, *model.GetMapInfoRequest) (*model.GetMapInfoResponse, error)
}

type gameDomain struct {
	fileRepo      repository.FileRepository
	gameRepo      repository.GameRepository
	userRepo      repository.UserRepository
	storage       storage.Storage
	maxUploadSize int
}

func NewGameDomain(
	gameRepo repository.GameRepository,
	userRepo repository.UserRepository,
	fileRepo repository.FileRepository,
	storage storage.Storage,
	cfg config.FileConfigs,
) *gameDomain {
	return &gameDomain{
		gameRepo:      gameRepo,
		userRepo:      userRepo,
		fileRepo:      fileRepo,
		storage:       storage,
		maxUploadSize: cfg.MaxSize * 1024 * 1024,
	}
}

func (d *gameDomain) CreateMap(
	ctx xcontext.Context, req *model.CreateMapRequest,
) (*model.CreateMapResponse, error) {
	if err := verifyUserRole(ctx, d.userRepo, []string{entity.SuperAdminRole, entity.AdminRole}); err != nil {
		return nil, err
	}

	if err := ctx.Request().ParseMultipartForm(int64(d.maxUploadSize)); err != nil {
		return nil, errorx.New(errorx.BadRequest, "Request must be multipart form")
	}

	mapObject, err := formToStorageObject(ctx, "map", "application/json")
	if err != nil {
		return nil, err
	}

	tileSetObject, err := formToStorageObject(ctx, "tileset", "image/png")
	if err != nil {
		return nil, err
	}

	playerImgObject, err := formToStorageObject(ctx, "player_img", "image/png")
	if err != nil {
		return nil, err
	}

	playerJsonObject, err := formToStorageObject(ctx, "player_json", "application/json")
	if err != nil {
		return nil, err
	}

	_, err = gameengine.ParseGameMap(mapObject.Data)
	if err != nil {
		ctx.Logger().Errorf("Cannot parse game map: %v", err)
		return nil, errorx.New(errorx.BadRequest, "invalid game map")
	}

	resp, err := d.storage.BulkUpload(ctx, []*storage.UploadObject{
		mapObject, tileSetObject, playerImgObject, playerJsonObject,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot upload image: %v", err)
		return nil, errorx.New(errorx.Internal, "Unable to upload image")
	}

	name := ctx.Request().PostFormValue("name")
	if name == "" {
		return nil, errorx.New(errorx.BadRequest, "Not found map name")
	}

	gameMap := &entity.GameMap{
		Base:           entity.Base{ID: uuid.NewString()},
		Name:           name,
		Map:            mapObject.Data,
		MapPath:        resp[0].FileName,
		TileSetPath:    resp[1].FileName,
		PlayerImgPath:  resp[2].FileName,
		PlayerJSONPath: resp[3].FileName,
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
	if err := verifyUserRole(ctx, d.userRepo, []string{entity.SuperAdminRole, entity.AdminRole}); err != nil {
		return nil, err
	}
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

func (d *gameDomain) DeleteMap(ctx xcontext.Context, req *model.DeleteMapRequest) (*model.DeleteMapResponse, error) {
	if err := verifyUserRole(ctx, d.userRepo, []string{entity.SuperAdminRole, entity.AdminRole}); err != nil {
		return nil, err
	}
	if err := d.gameRepo.DeleteMap(ctx, req.ID); err != nil {
		ctx.Logger().Errorf("Cannot create room: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteMapResponse{}, nil
}

func (d *gameDomain) DeleteRoom(ctx xcontext.Context, req *model.DeleteRoomRequest) (*model.DeleteRoomResponse, error) {
	if err := verifyUserRole(ctx, d.userRepo, []string{entity.SuperAdminRole, entity.AdminRole}); err != nil {
		return nil, err
	}
	if err := d.gameRepo.DeleteRoom(ctx, req.ID); err != nil {
		ctx.Logger().Errorf("Cannot create room: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteRoomResponse{}, nil
}

func verifyUserRole(ctx xcontext.Context, userRepo repository.UserRepository, acceptRoles []string) error {
	userID := xcontext.GetRequestUserID(ctx)
	u, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorx.New(errorx.NotFound, "Not found user")
		}

		ctx.Logger().Errorf("Cannot get user: %v", err)
		return errorx.Unknown
	}

	if !slices.Contains(acceptRoles, u.Role) {
		ctx.Logger().Errorf("User doesn't have permission: %v", err)
		return errorx.New(errorx.Unauthenticated, "User doesn't have permission")
	}
	return nil
}

func (d *gameDomain) GetMapInfo(
	ctx xcontext.Context, req *model.GetMapInfoRequest,
) (*model.GetMapInfoResponse, error) {
	room, err := d.gameRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get room: %v", err)
		return nil, errorx.Unknown
	}

	gameMap, err := d.gameRepo.GetMapByID(ctx, room.MapID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get map: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetMapInfoResponse{
		MapPath:        gameMap.MapPath,
		TilesetPath:    gameMap.TileSetPath,
		PlayerImgPath:  gameMap.PlayerImgPath,
		PlayerJsonPath: gameMap.PlayerJSONPath,
	}, nil
}

func formToStorageObject(ctx xcontext.Context, name, mime string) (*storage.UploadObject, error) {
	file, _, err := ctx.Request().FormFile(name)
	if err != nil {
		ctx.Logger().Errorf("Cannot get the %s: %v", name, err)
		return nil, errorx.New(errorx.BadRequest, "Cannot get the %s", name)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		ctx.Logger().Errorf("Cannot read file file: %v", err)
		return nil, errorx.Unknown
	}

	return &storage.UploadObject{
		Bucket:   string(entity.Game),
		Prefix:   "",
		FileName: name,
		Mime:     mime,
		Data:     content,
	}, nil
}
