package domain

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

type GameDomain interface {
	CreateMap(context.Context, *model.CreateGameMapRequest) (*model.CreateGameMapResponse, error)
	CreateRoom(context.Context, *model.CreateGameRoomRequest) (*model.CreateGameRoomResponse, error)
	UpdateTileset(context.Context, *model.UpdateGameMapTilesetRequest) (*model.UpdateGameMapTilesetResponse, error)
	UpdatePlayer(context.Context, *model.UpdateGameMapPlayerRequest) (*model.UpdateGameMapPlayerResponse, error)
	DeleteMap(context.Context, *model.DeleteMapRequest) (*model.DeleteMapResponse, error)
	DeleteRoom(context.Context, *model.DeleteRoomRequest) (*model.DeleteRoomResponse, error)
	GetMaps(context.Context, *model.GetMapsRequest) (*model.GetMapsResponse, error)
	GetRooms(context.Context, *model.GetRoomsRequest) (*model.GetRoomsResponse, error)
}

type gameDomain struct {
	fileRepo      repository.FileRepository
	gameRepo      repository.GameRepository
	userRepo      repository.UserRepository
	communityRepo repository.CommunityRepository
	storage       storage.Storage
}

func NewGameDomain(
	gameRepo repository.GameRepository,
	userRepo repository.UserRepository,
	fileRepo repository.FileRepository,
	communityRepo repository.CommunityRepository,
	storage storage.Storage,
	cfg config.FileConfigs,
) *gameDomain {
	return &gameDomain{
		gameRepo:      gameRepo,
		userRepo:      userRepo,
		fileRepo:      fileRepo,
		communityRepo: communityRepo,
		storage:       storage,
	}
}

func (d *gameDomain) CreateMap(
	ctx context.Context, req *model.CreateGameMapRequest,
) (*model.CreateGameMapResponse, error) {
	httpReq := xcontext.HTTPRequest(ctx)
	if err := httpReq.ParseMultipartForm(xcontext.Configs(ctx).File.MaxMemory); err != nil {
		return nil, errorx.New(errorx.BadRequest, "Request must be multipart form")
	}

	mapConfig, err := formToGameStorageObject(ctx, "map", "application/json")
	if err != nil {
		return nil, err
	}

	collisionLayers := httpReq.PostFormValue("collision_layers")
	if collisionLayers == "" {
		return nil, errorx.New(errorx.BadRequest, "Not found collision layers")
	}

	_, err = gameengine.ParseGameMap(mapConfig.Data, strings.Split(collisionLayers, ","))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot parse game map: %v", err)
		return nil, errorx.New(errorx.BadRequest, "invalid game map")
	}

	resp, err := d.storage.Upload(ctx, mapConfig)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot upload map config: %v", err)
		return nil, errorx.New(errorx.Internal, "Unable to upload map config")
	}

	name := httpReq.PostFormValue("name")
	if name == "" {
		return nil, errorx.New(errorx.BadRequest, "Not found map name")
	}

	initX, err := strconv.Atoi(httpReq.PostFormValue("init_x"))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot parse init x: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid init x")
	}

	initY, err := strconv.Atoi(httpReq.PostFormValue("init_y"))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot parse init y: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid init y")
	}

	gameMap := &entity.GameMap{
		Base:            entity.Base{ID: uuid.NewString()},
		Name:            name,
		InitX:           initX,
		InitY:           initY,
		ConfigURL:       resp.Url,
		CollisionLayers: collisionLayers,
	}

	if err := d.gameRepo.CreateMap(ctx, gameMap); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create map: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateGameMapResponse{ID: gameMap.ID}, nil
}

func (d *gameDomain) CreateRoom(
	ctx context.Context, req *model.CreateGameRoomRequest,
) (*model.CreateGameRoomResponse, error) {
	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	room := &entity.GameRoom{
		Base:        entity.Base{ID: uuid.NewString()},
		CommunityID: community.ID,
		MapID:       req.MapID,
		Name:        req.Name,
	}

	if err := d.gameRepo.CreateRoom(ctx, room); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create room: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateGameRoomResponse{ID: room.ID}, nil
}

func (d *gameDomain) UpdateTileset(
	ctx context.Context, req *model.UpdateGameMapTilesetRequest,
) (*model.UpdateGameMapTilesetResponse, error) {
	gameMapID := xcontext.HTTPRequest(ctx).PostFormValue("game_map_id")
	if gameMapID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow an empty game map id")
	}

	tileSet, err := formToGameStorageObject(ctx, "tileset", "image/png")
	if err != nil {
		return nil, err
	}

	resp, err := d.storage.Upload(ctx, tileSet)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot upload tileset: %v", err)
		return nil, errorx.New(errorx.Internal, "Unable to upload tileset")
	}

	err = d.gameRepo.CreateGameTileset(ctx, &entity.GameMapTileset{
		Base:       entity.Base{ID: uuid.NewString()},
		GameMapID:  gameMapID,
		TilesetURL: resp.Url,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create tileset: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateGameMapTilesetResponse{}, nil
}

func (d *gameDomain) UpdatePlayer(
	ctx context.Context, req *model.UpdateGameMapPlayerRequest,
) (*model.UpdateGameMapPlayerResponse, error) {
	gameMapID := xcontext.HTTPRequest(ctx).PostFormValue("game_map_id")
	if gameMapID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow an empty game map id")
	}

	name := xcontext.HTTPRequest(ctx).PostFormValue("name")
	if name == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow an empty player name")
	}

	playerImage, err := formToGameStorageObject(ctx, "player_img", "image/png")
	if err != nil {
		return nil, err
	}

	playerConfig, err := formToGameStorageObject(ctx, "player_cfg", "application/json")
	if err != nil {
		return nil, err
	}

	playerImageResp, err := d.storage.Upload(ctx, playerImage)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot upload player image: %v", err)
		return nil, errorx.New(errorx.Internal, "Unable to upload player image")
	}

	playerConfigResp, err := d.storage.Upload(ctx, playerConfig)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot upload player config: %v", err)
		return nil, errorx.New(errorx.Internal, "Unable to upload player config")
	}

	err = d.gameRepo.CreateGamePlayer(ctx, &entity.GameMapPlayer{
		Base:      entity.Base{ID: uuid.NewString()},
		Name:      name,
		GameMapID: gameMapID,
		ConfigURL: playerConfigResp.Url,
		ImageURL:  playerImageResp.Url,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create tileset: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateGameMapPlayerResponse{}, nil
}

func (d *gameDomain) DeleteMap(ctx context.Context, req *model.DeleteMapRequest) (*model.DeleteMapResponse, error) {
	if err := d.gameRepo.DeleteMap(ctx, req.ID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create room: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteMapResponse{}, nil
}

func (d *gameDomain) DeleteRoom(ctx context.Context, req *model.DeleteRoomRequest) (*model.DeleteRoomResponse, error) {
	if err := d.gameRepo.DeleteRoom(ctx, req.ID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create room: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteRoomResponse{}, nil
}

func (d *gameDomain) GetMaps(
	ctx context.Context, req *model.GetMapsRequest,
) (*model.GetMapsResponse, error) {
	maps, err := d.gameRepo.GetMaps(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get maps: %v", err)
		return nil, errorx.Unknown
	}

	clientMaps := []model.GameMap{}
	for _, gameMap := range maps {
		gameMapTilesets, err := d.gameRepo.GetTilesetByMapID(ctx, gameMap.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get tileset by map id: %v", err)
			return nil, errorx.Unknown
		}

		clientTilesets := []model.GameMapTileset{}
		for _, tileset := range gameMapTilesets {
			clientTilesets = append(clientTilesets, convertGameMapTileset(&tileset))
		}

		gameMapPlayers, err := d.gameRepo.GetPlayerByMapID(ctx, gameMap.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get tileset by map id: %v", err)
			return nil, errorx.Unknown
		}

		clientPlayers := []model.GameMapPlayer{}
		for _, player := range gameMapPlayers {
			clientPlayers = append(clientPlayers, convertGameMapPlayer(&player))
		}

		clientMaps = append(clientMaps, convertGameMap(&gameMap, clientTilesets, clientPlayers))
	}

	return &model.GetMapsResponse{GameMaps: clientMaps}, nil
}

func (d *gameDomain) GetRooms(
	ctx context.Context, req *model.GetRoomsRequest,
) (*model.GetRoomsResponse, error) {
	communityID := ""
	if req.CommunityHandle != "" {
		community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found community")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}

		communityID = community.ID
	}

	rooms, err := d.gameRepo.GetRoomsByCommunityID(ctx, communityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get rooms: %v", err)
		return nil, errorx.Unknown
	}

	gameMapSet := map[string]*entity.GameMap{}
	communitySet := map[string]*entity.Community{}
	for _, room := range rooms {
		gameMapSet[room.MapID] = nil
		communitySet[room.CommunityID] = nil
	}

	gameMaps, err := d.gameRepo.GetMapByIDs(ctx, common.MapKeys(gameMapSet))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get game map: %v", err)
		return nil, errorx.Unknown
	}

	for i := range gameMaps {
		gameMapSet[gameMaps[i].ID] = &gameMaps[i]
	}

	communities, err := d.communityRepo.GetByIDs(ctx, common.MapKeys(communitySet))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get communities: %v", err)
		return nil, errorx.Unknown
	}

	for i := range communities {
		communitySet[communities[i].ID] = &communities[i]
	}

	clientRooms := []model.GameRoom{}
	for _, room := range rooms {
		gameMap, ok := gameMapSet[room.MapID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Invalid map %s for room %s: %v", room.MapID, room.ID, err)
			return nil, errorx.Unknown
		}

		community, ok := communitySet[room.CommunityID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Invalid community %s for room %s: %v", room.CommunityID, room.ID, err)
			return nil, errorx.Unknown
		}

		gameMapTilesets, err := d.gameRepo.GetTilesetByMapID(ctx, gameMap.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get tileset by map id: %v", err)
			return nil, errorx.Unknown
		}

		clientTilesets := []model.GameMapTileset{}
		for _, tileset := range gameMapTilesets {
			clientTilesets = append(clientTilesets, convertGameMapTileset(&tileset))
		}

		gameMapPlayers, err := d.gameRepo.GetPlayerByMapID(ctx, gameMap.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get tileset by map id: %v", err)
			return nil, errorx.Unknown
		}

		clientPlayers := []model.GameMapPlayer{}
		for _, player := range gameMapPlayers {
			clientPlayers = append(clientPlayers, convertGameMapPlayer(&player))
		}

		clientRooms = append(
			clientRooms,
			convertGameRoom(
				&room,
				convertGameMap(gameMap, clientTilesets, clientPlayers),
				convertCommunity(community, 0),
			),
		)
	}

	return &model.GetRoomsResponse{GameRooms: clientRooms}, nil
}

func formToGameStorageObject(ctx context.Context, name, mime string) (*storage.UploadObject, error) {
	file, _, err := xcontext.HTTPRequest(ctx).FormFile(name)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get the %s: %v", name, err)
		return nil, errorx.New(errorx.BadRequest, "Cannot get the %s", name)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot read file file: %v", err)
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
