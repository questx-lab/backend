package domain

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

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
	DeleteMap(context.Context, *model.DeleteMapRequest) (*model.DeleteMapResponse, error)
	DeleteRoom(context.Context, *model.DeleteRoomRequest) (*model.DeleteRoomResponse, error)
	GetMaps(context.Context, *model.GetMapsRequest) (*model.GetMapsResponse, error)
	GetRoomsByCommunity(context.Context, *model.GetRoomsByCommunityRequest) (*model.GetRoomsByCommunityResponse, error)
	CreateLuckyboxEvent(context.Context, *model.CreateLuckyboxEventRequest) (*model.CreateLuckyboxEventResponse, error)
}

type gameDomain struct {
	fileRepo      repository.FileRepository
	gameRepo      repository.GameRepository
	userRepo      repository.UserRepository
	communityRepo repository.CommunityRepository
	storage       storage.Storage
	roleVerifier  *common.CommunityRoleVerifier
}

func NewGameDomain(
	gameRepo repository.GameRepository,
	userRepo repository.UserRepository,
	fileRepo repository.FileRepository,
	communityRepo repository.CommunityRepository,
	collaboratorRepo repository.CollaboratorRepository,
	storage storage.Storage,
	cfg config.FileConfigs,
) *gameDomain {
	return &gameDomain{
		gameRepo:      gameRepo,
		userRepo:      userRepo,
		fileRepo:      fileRepo,
		communityRepo: communityRepo,
		storage:       storage,
		roleVerifier:  common.NewCommunityRoleVerifier(collaboratorRepo, userRepo),
	}
}

func (d *gameDomain) CreateMap(
	ctx context.Context, req *model.CreateGameMapRequest,
) (*model.CreateGameMapResponse, error) {
	httpReq := xcontext.HTTPRequest(ctx)
	if err := httpReq.ParseMultipartForm(xcontext.Configs(ctx).File.MaxMemory); err != nil {
		return nil, errorx.New(errorx.BadRequest, "Request must be multipart form")
	}

	configFile, err := formToGameStorageObject(ctx, "config_file", "application/json")
	if err != nil {
		return nil, err
	}

	var mapConfig gameengine.MapConfig
	if err := json.Unmarshal(configFile.Data, &mapConfig); err != nil {
		xcontext.Logger(ctx).Debugf("Cannot parse config file: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid config file")
	}

	mapData, err := d.storage.DownloadFromURL(ctx, mapConfig.PathOf(mapConfig.Config))
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot download map data: %v", err)
		return nil, errorx.New(errorx.Unavailable, "Cannot download map data")
	}

	if len(mapConfig.CollisionLayers) == 0 {
		return nil, errorx.New(errorx.BadRequest, "Not found collision layers")
	}

	parsedMap, err := gameengine.ParseGameMap(mapData, mapConfig.CollisionLayers)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot parse game map: %v", err)
		return nil, errorx.New(errorx.BadRequest, "invalid game map")
	}

	initPos := mapConfig.InitPosition
	if parsedMap.IsPointCollision(initPos) {
		return nil, errorx.New(errorx.Unavailable,
			"The initial position is collide with blocked objects")
	}

	for _, character := range mapConfig.CharacterConfigs {
		characterData, err := d.storage.DownloadFromURL(ctx, mapConfig.PathOf(character.Config))
		if err != nil {
			xcontext.Logger(ctx).Debugf("Cannot download character data of %s: %v", character.Name, err)
			return nil, errorx.New(errorx.Unavailable, "Cannot download character data of %s", character.Name)
		}

		parsedCharacter, err := gameengine.ParseCharacter(characterData)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Cannot parse character data of %s: %v", character.Name, err)
			return nil, errorx.New(errorx.Unavailable, "Cannot parse character data of %s", character.Name)
		}

		character := gameengine.Character{
			Size: gameengine.Size{
				Width:  parsedCharacter.Width,
				Height: parsedCharacter.Height,
			},
		}

		if parsedMap.IsCollision(initPos.CenterToTopLeft(character.Size), character.Size) {
			return nil, errorx.New(errorx.Unavailable, "The character is collide with blocked objects")
		}
	}

	name := httpReq.FormValue("name")
	id := httpReq.FormValue("id")
	if id == "" {
		if name == "" {
			return nil, errorx.New(errorx.BadRequest, "Need name parameter if create a new map")
		}

		id = uuid.NewString()
	} else {
		gameMap, err := d.gameRepo.GetMapByID(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found map")
			}

			xcontext.Logger(ctx).Errorf("Cannot get map by id: %v", err)
			return nil, errorx.Unknown
		}

		if name == "" {
			name = gameMap.Name
		}
	}

	resp, err := d.storage.Upload(ctx, configFile)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot upload map config: %v", err)
		return nil, errorx.New(errorx.Internal, "Unable to upload map config")
	}

	gameMap := &entity.GameMap{
		Base:      entity.Base{ID: id},
		Name:      name,
		ConfigURL: resp.Url,
	}

	if err := d.gameRepo.UpsertMap(ctx, gameMap); err != nil {
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
		clientMaps = append(clientMaps, convertGameMap(&gameMap))
	}

	return &model.GetMapsResponse{GameMaps: clientMaps}, nil
}

func (d *gameDomain) GetRoomsByCommunity(
	ctx context.Context, req *model.GetRoomsByCommunityRequest,
) (*model.GetRoomsByCommunityResponse, error) {
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow an empty community handle")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	rooms, err := d.gameRepo.GetRoomsByCommunityID(ctx, community.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get rooms: %v", err)
		return nil, errorx.Unknown
	}

	gameMapSet := map[string]*entity.GameMap{}
	for _, room := range rooms {
		gameMapSet[room.MapID] = nil
	}

	gameMaps, err := d.gameRepo.GetMapByIDs(ctx, common.MapKeys(gameMapSet))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get game map: %v", err)
		return nil, errorx.Unknown
	}

	for i := range gameMaps {
		gameMapSet[gameMaps[i].ID] = &gameMaps[i]
	}

	clientRooms := []model.GameRoom{}
	for _, room := range rooms {
		gameMap, ok := gameMapSet[room.MapID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Invalid map %s for room %s: %v", room.MapID, room.ID, err)
			return nil, errorx.Unknown
		}

		clientRooms = append(clientRooms, convertGameRoom(&room, convertGameMap(gameMap)))
	}

	return &model.GetRoomsByCommunityResponse{Community: convertCommunity(community, 0), GameRooms: clientRooms}, nil
}

func (d *gameDomain) CreateLuckyboxEvent(
	ctx context.Context, req *model.CreateLuckyboxEventRequest,
) (*model.CreateLuckyboxEventResponse, error) {
	if req.RoomID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow an empty room id")
	}

	if req.NumberOfBoxes <= 0 {
		return nil, errorx.New(errorx.BadRequest, "Not allow a non-positive number_of_boxes")
	}

	if req.NumberOfBoxes > xcontext.Configs(ctx).Game.MaxLuckyboxPerEvent {
		return nil, errorx.New(errorx.BadRequest, "Too many boxes")
	}

	if req.PointPerBox <= 0 {
		return nil, errorx.New(errorx.BadRequest, "Not allow a non-positive point_per_box")
	}

	if !req.StartTime.IsZero() && req.StartTime.Before(time.Now()) {
		return nil, errorx.New(errorx.BadRequest, "Invalid start time")
	}

	if req.StartTime.IsZero() {
		req.StartTime = time.Now()
	}

	if req.Duration < xcontext.Configs(ctx).Game.MinLuckyboxEventDuration {
		return nil, errorx.New(errorx.BadRequest, "Event duration must be larger than %s",
			xcontext.Configs(ctx).Game.MinLuckyboxEventDuration)
	}

	if req.Duration > xcontext.Configs(ctx).Game.MaxLuckyboxEventDuration {
		return nil, errorx.New(errorx.BadRequest, "Event duration must be less than %s",
			xcontext.Configs(ctx).Game.MaxLuckyboxEventDuration)
	}

	room, err := d.gameRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found room")
		}

		xcontext.Logger(ctx).Errorf("Cannot get room: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, room.CommunityID, entity.AdminGroup...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denined")
	}

	luckyboxEvent := &entity.GameLuckyboxEvent{
		Base:        entity.Base{ID: uuid.NewString()},
		RoomID:      req.RoomID,
		Amount:      req.NumberOfBoxes,
		PointPerBox: req.PointPerBox,
		IsRandom:    req.IsRandom,
		StartTime:   req.StartTime,
		EndTime:     req.StartTime.Add(req.Duration),
		IsStarted:   false,
		IsStopped:   false,
	}

	happenInRangeEvents, err := d.gameRepo.GetLuckyboxEventsHappenInRange(
		ctx, room.ID, luckyboxEvent.StartTime, luckyboxEvent.EndTime)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get luckybox events happen in range: %v", err)
		return nil, errorx.Unknown
	}

	numberOfBoxesAtMinute := make([]int, req.Duration/time.Second)
	for i := 0; i < int(req.Duration/time.Second); i++ {
		numberOfBoxesAtMinute[i] = luckyboxEvent.Amount
	}

	for _, event := range happenInRangeEvents {
		startTime := event.StartTime
		if event.StartTime.Before(luckyboxEvent.StartTime) {
			startTime = luckyboxEvent.StartTime
		}

		endTime := event.EndTime
		if event.EndTime.After(luckyboxEvent.EndTime) {
			endTime = luckyboxEvent.EndTime
		}

		for i := startTime.Unix(); i < endTime.Unix(); i++ {
			index := i - luckyboxEvent.StartTime.Unix()
			numberOfBoxesAtMinute[index] += event.Amount
			if numberOfBoxesAtMinute[index] > xcontext.Configs(ctx).Game.MaxLuckyboxPerEvent {
				return nil, errorx.New(errorx.Unavailable, "Cannot create more event in that time")
			}
		}
	}

	err = d.gameRepo.CreateLuckyboxEvent(ctx, luckyboxEvent)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create luckybox event")
		return nil, errorx.Unknown
	}

	return &model.CreateLuckyboxEventResponse{}, nil
}

func formToGameStorageObject(ctx context.Context, name, mime string) (*storage.UploadObject, error) {
	file, header, err := xcontext.HTTPRequest(ctx).FormFile(name)
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

	prefix := "common"
	if mime == "application/json" {
		prefix = "configs"
	} else if mime == "application/png" {
		prefix = "images"
	}

	return &storage.UploadObject{
		Bucket:   string(entity.Game),
		Prefix:   prefix,
		FileName: header.Filename,
		Mime:     mime,
		Data:     content,
	}, nil
}
