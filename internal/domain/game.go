package domain

import (
	"errors"
	"io"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/domain/gamestate"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type GameDomain interface {
	CreateMap(xcontext.Context, *model.CreateMapRequest) (*model.CreateMapResponse, error)
	CreateRoom(xcontext.Context, *model.CreateRoomRequest) (*model.CreateRoomResponse, error)
}

type gameDomain struct {
	gameRepo      repository.GameRepository
	userRepo      repository.UserRepository
	maxUploadSize int
}

func NewGameDomain(
	gameRepo repository.GameRepository,
	userRepo repository.UserRepository,
	cfg config.FileConfigs,
) *gameDomain {
	return &gameDomain{
		gameRepo:      gameRepo,
		userRepo:      userRepo,
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

	file, _, err := ctx.Request().FormFile("file")
	if err != nil {
		return nil, errorx.New(errorx.BadRequest, "Cannot get the file")
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		ctx.Logger().Errorf("Cannot read file file: %v", err)
		return nil, errorx.Unknown
	}

	_, err = gamestate.ParseGameMap(fileContent)
	if err != nil {
		ctx.Logger().Errorf("Cannot parse game map: %v", err)
		return nil, errorx.New(errorx.BadRequest, "invalid game map")
	}

	name := ctx.Request().PostFormValue("name")
	if name == "" {
		return nil, errorx.New(errorx.BadRequest, "Not found map name")
	}

	gameMap := &entity.GameMap{
		Base:    entity.Base{ID: uuid.NewString()},
		Name:    name,
		Content: fileContent,
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
