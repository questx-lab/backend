package domain

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/pubsub"
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

	// Game luckybox
	CreateLuckyboxEvent(context.Context, *model.CreateLuckyboxEventRequest) (*model.CreateLuckyboxEventResponse, error)

	// Game character
	CreateCharacter(context.Context, *model.CreateGameCharacterRequest) (*model.CreateGameCharacterResponse, error)
	GetAllCharacters(context.Context, *model.GetAllGameCharactersRequest) (*model.GetAllGameCharactersResponse, error)
	SetupCommunityCharacter(context.Context, *model.SetupCommunityCharacterRequest) (*model.SetupCommunityCharacterResponse, error)
	GetAllCommunityCharacters(context.Context, *model.GetAllCommunityCharactersRequest) (*model.GetAllCommunityCharactersResponse, error)
	BuyCharacter(context.Context, *model.BuyCharacterRequest) (*model.BuyCharacterResponse, error)
	GetMyCharacters(context.Context, *model.GetMyCharactersRequest) (*model.GetMyCharactersResponse, error)
}

type gameDomain struct {
	fileRepo          repository.FileRepository
	gameRepo          repository.GameRepository
	gameLuckyboxRepo  repository.GameLuckyboxRepository
	gameCharacterRepo repository.GameCharacterRepository
	userRepo          repository.UserRepository
	communityRepo     repository.CommunityRepository
	followerRepo      repository.FollowerRepository
	storage           storage.Storage
	publisher         pubsub.Publisher
	roleVerifier      *common.CommunityRoleVerifier
}

func NewGameDomain(
	gameRepo repository.GameRepository,
	gameLuckyboxRepo repository.GameLuckyboxRepository,
	gameCharacterRepo repository.GameCharacterRepository,
	userRepo repository.UserRepository,
	fileRepo repository.FileRepository,
	communityRepo repository.CommunityRepository,
	collaboratorRepo repository.CollaboratorRepository,
	followerRepo repository.FollowerRepository,
	storage storage.Storage,
	publisher pubsub.Publisher,
	cfg config.FileConfigs,
) *gameDomain {
	return &gameDomain{
		gameRepo:          gameRepo,
		gameLuckyboxRepo:  gameLuckyboxRepo,
		gameCharacterRepo: gameCharacterRepo,
		userRepo:          userRepo,
		fileRepo:          fileRepo,
		communityRepo:     communityRepo,
		followerRepo:      followerRepo,
		storage:           storage,
		publisher:         publisher,
		roleVerifier:      common.NewCommunityRoleVerifier(collaboratorRepo, userRepo),
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

	happenInRangeEvents, err := d.gameLuckyboxRepo.GetLuckyboxEventsHappenInRange(
		ctx, room.ID, luckyboxEvent.StartTime, luckyboxEvent.EndTime)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get luckybox events happen in range: %v", err)
		return nil, errorx.Unknown
	}

	numberOfBoxesAtMinute := make([]int, req.Duration/time.Minute)
	for i := 0; i < int(req.Duration/time.Minute); i++ {
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

		start := int(startTime.Sub(luckyboxEvent.StartTime) / time.Minute)
		end := int(endTime.Sub(luckyboxEvent.StartTime) / time.Minute)
		for i := start; i < end; i++ {
			numberOfBoxesAtMinute[i] += event.Amount
			if numberOfBoxesAtMinute[i] > xcontext.Configs(ctx).Game.MaxLuckyboxPerEvent {
				return nil, errorx.New(errorx.Unavailable, "Cannot create more event in that time")
			}
		}
	}

	err = d.gameLuckyboxRepo.CreateLuckyboxEvent(ctx, luckyboxEvent)
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

func (d *gameDomain) CreateCharacter(
	ctx context.Context, req *model.CreateGameCharacterRequest,
) (*model.CreateGameCharacterResponse, error) {
	if req.Name == "" {
		return nil, errorx.New(errorx.BadRequest, "Require a name")
	}

	if req.Level < 0 {
		return nil, errorx.New(errorx.BadRequest, "Require a non-negative number of level")
	}

	data, err := d.storage.DownloadFromURL(ctx, req.ConfigURL)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot download config url: %v", err)
		return nil, errorx.New(errorx.Unavailable, "Cannot download config url")
	}

	_, err = gameengine.ParseCharacter(data)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot parse config data: %v", err)
		return nil, errorx.New(errorx.Unavailable, "Cannot parse config data")
	}

	spriteWidthRatio := 0.5
	spriteHeightRatio := 0.2
	if req.SpriteWidthRatio != 0 {
		spriteWidthRatio = req.SpriteWidthRatio
	}
	if req.SpriteHeightRatio != 0 {
		spriteHeightRatio = req.SpriteHeightRatio
	}

	character := &entity.GameCharacter{
		Base:              entity.Base{ID: uuid.NewString()},
		Name:              req.Name,
		Level:             req.Level,
		ConfigURL:         req.ConfigURL,
		ImageURL:          req.ImageURL,
		SpriteWidthRatio:  spriteWidthRatio,
		SpriteHeightRatio: spriteHeightRatio,
	}

	err = d.gameCharacterRepo.Create(ctx, character)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create game character: %v", err)
		return nil, errorx.Unknown
	}

	err = d.publisher.Publish(ctx, model.CreateCharacterTopic, &pubsub.Pack{Key: []byte(character.ID)})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot publish create character event: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateGameCharacterResponse{}, nil
}

func (d *gameDomain) GetAllCharacters(
	ctx context.Context, req *model.GetAllGameCharactersRequest,
) (*model.GetAllGameCharactersResponse, error) {
	characters, err := d.gameCharacterRepo.GetAll(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all characters: %v", err)
		return nil, errorx.Unknown
	}

	clientCharacters := []model.GameCharacter{}
	for _, c := range characters {
		clientCharacters = append(clientCharacters, convertGameCharacter(&c))
	}

	return &model.GetAllGameCharactersResponse{GameCharacters: clientCharacters}, nil
}

func (d *gameDomain) SetupCommunityCharacter(
	ctx context.Context, req *model.SetupCommunityCharacterRequest,
) (*model.SetupCommunityCharacterResponse, error) {
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow community handle")
	}

	if req.Points < 0 {
		return nil, errorx.New(errorx.BadRequest, "Not allow a negative points")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, community.ID, entity.Owner); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	err = d.gameCharacterRepo.CreateCommunityCharacter(ctx, &entity.GameCommunityCharacter{
		CommunityID: community.ID,
		CharacterID: req.CharacterID,
		Points:      req.Points,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create community character: %v", err)
		return nil, errorx.Unknown
	}

	return &model.SetupCommunityCharacterResponse{}, nil
}

func (d *gameDomain) GetAllCommunityCharacters(
	ctx context.Context, req *model.GetAllCommunityCharactersRequest,
) (*model.GetAllCommunityCharactersResponse, error) {
	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	communityCharacters, err := d.gameCharacterRepo.GetAllCommunityCharacters(ctx, community.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all community characters: %v", err)
		return nil, errorx.Unknown
	}

	communityCharacterMap := map[string]entity.GameCommunityCharacter{}
	for _, c := range communityCharacters {
		communityCharacterMap[c.CharacterID] = c
	}

	characters, err := d.gameCharacterRepo.GetAll(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all characters: %v", err)
		return nil, errorx.Unknown
	}

	clientCharacters := []model.GameCommunityCharacter{}
	for _, character := range characters {
		communityCharacter, ok := communityCharacterMap[character.ID]
		if !ok {
			communityCharacter = entity.GameCommunityCharacter{
				CommunityID: community.ID,
				CharacterID: character.ID,
				Points:      math.MaxInt,
			}
		}

		clientCharacters = append(
			clientCharacters,
			convertGameCommunityCharacter(&communityCharacter, convertGameCharacter(&character)),
		)
	}

	return &model.GetAllCommunityCharactersResponse{CommunityCharacters: clientCharacters}, nil
}

func (d *gameDomain) BuyCharacter(
	ctx context.Context, req *model.BuyCharacterRequest,
) (*model.BuyCharacterResponse, error) {
	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	userID := xcontext.RequestUserID(ctx)
	userCharacters, err := d.gameCharacterRepo.GetAllUserCharacters(ctx, userID, community.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get user character: %v", err)
		return nil, errorx.Unknown
	}

	for _, uc := range userCharacters {
		if uc.CharacterID == req.CharacterID {
			return nil, errorx.New(errorx.AlreadyExists, "User had already bought this character before")
		}
	}

	character, err := d.gameCharacterRepo.GetByID(ctx, req.CharacterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found character")
		}

		xcontext.Logger(ctx).Errorf("Cannot get character: %v", err)
		return nil, errorx.Unknown
	}

	var communityCharacter *entity.GameCommunityCharacter
	// In this case, user had already bought a free character. We need to check
	// if:
	// - user has bought the previous level before.
	// - this character is for sale now.
	// - user has enough points.
	if len(userCharacters) > 0 {
		// User must buy the previous level character before.
		if character.Level > 1 {
			previousCharacter, err := d.gameCharacterRepo.Get(ctx, character.Name, character.Level-1)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Not found the previous character of %s: %v", character.ID, err)
				return nil, errorx.Unknown
			}

			buyPreviousCharacter := false
			for _, uc := range userCharacters {
				if uc.CharacterID == previousCharacter.ID {
					buyPreviousCharacter = true
					break
				}
			}

			if !buyPreviousCharacter {
				return nil, errorx.New(errorx.Unavailable, "You must buy the previous level before")
			}
		}

		communityCharacter, err = d.gameCharacterRepo.GetCommunityCharacter(
			ctx, community.ID, req.CharacterID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "This character is not for sale yet")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community character: %v", err)
			return nil, errorx.Unknown
		}

		if communityCharacter.Points > 0 {
			follower, err := d.followerRepo.Get(ctx, userID, community.ID)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get follower info: %v", err)
				return nil, errorx.New(errorx.Unavailable, "User has not follow the community yet")
			}

			if follower.Points < uint64(communityCharacter.Points) {
				return nil, errorx.New(errorx.Unavailable, "Not enough points")
			}

			// User must buy this character if he chooses a free character before.
			err = d.followerRepo.DecreasePoint(
				ctx, userID, community.ID, uint64(communityCharacter.Points), false)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot decrease point of user: %v", err)
				return nil, errorx.Unknown
			}
		}
	} else {
		// When user has no character in game, we can give user a free 0-level
		// character without any limitation.
		if character.Level != 0 {
			return nil, errorx.New(errorx.Unavailable, "Please choose a free character of level 0")
		}
	}

	err = d.gameCharacterRepo.CreateUserCharacter(ctx, &entity.GameUserCharacter{
		UserID:      userID,
		CommunityID: community.ID,
		CharacterID: req.CharacterID,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create user character: %v", err)
		return nil, errorx.Unknown
	}

	ctx = xcontext.WithCommitDBTransaction(ctx)

	gameRooms, err := d.gameRepo.GetRoomsByUserCommunity(ctx, userID, community.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get user by community: %v", err)
		return nil, errorx.Unknown
	}

	serverAction := model.GameActionServerRequest{
		UserID: "",
		Type:   gameengine.BuyCharacterAction{}.Type(),
		Value: map[string]any{
			"buy_user_id":  userID,
			"character_id": req.CharacterID,
		},
	}

	b, err := json.Marshal(serverAction)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot marshal action: %v", err)
		return nil, errorx.Unknown
	}

	for _, room := range gameRooms {
		err := d.publisher.Publish(ctx, room.StartedBy, &pubsub.Pack{
			Key: []byte(room.ID),
			Msg: b,
		})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot publish the buy character topic: %v", err)
			return nil, errorx.Unknown
		}
	}

	return &model.BuyCharacterResponse{}, nil
}

func (d *gameDomain) GetMyCharacters(
	ctx context.Context, req *model.GetMyCharactersRequest,
) (*model.GetMyCharactersResponse, error) {
	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	userCharacters, err := d.gameCharacterRepo.GetAllUserCharacters(
		ctx, xcontext.RequestUserID(ctx), community.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get user characters: %v", err)
		return nil, errorx.Unknown
	}

	characters, err := d.gameCharacterRepo.GetAll(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get characters: %v", err)
		return nil, errorx.Unknown
	}

	characterMap := map[string]entity.GameCharacter{}
	for _, c := range characters {
		characterMap[c.ID] = c
	}

	clientCharacters := []model.GameUserCharacter{}
	for _, uc := range userCharacters {
		character, ok := characterMap[uc.CharacterID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Cannot get character info of %s", uc.CharacterID)
			continue
		}

		clientCharacters = append(clientCharacters,
			convertGameUserCharacter(&uc, convertGameCharacter(&character)),
		)
	}

	return &model.GetMyCharactersResponse{UserCharacters: clientCharacters}, nil
}
