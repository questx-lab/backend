package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

type CommunityDomain interface {
	Create(context.Context, *model.CreateCommunityRequest) (*model.CreateCommunityResponse, error)
	GetList(context.Context, *model.GetCommunitiesRequest) (*model.GetCommunitiesResponse, error)
	Get(context.Context, *model.GetCommunityRequest) (*model.GetCommunityResponse, error)
	UpdateByID(context.Context, *model.UpdateCommunityRequest) (*model.UpdateCommunityResponse, error)
	UpdateDiscord(context.Context, *model.UpdateCommunityDiscordRequest) (*model.UpdateCommunityDiscordResponse, error)
	DeleteByID(context.Context, *model.DeleteCommunityRequest) (*model.DeleteCommunityResponse, error)
	UploadLogo(context.Context, *model.UploadCommunityLogoRequest) (*model.UploadCommunityLogoResponse, error)
	GetMyInvitedCommunities(context.Context, *model.GetMyInvitedCommunitiesRequest) (*model.GetInvitedCommunitiesResponse, error)
	GetPendingInvitedCommunities(context.Context, *model.GetPendingInviteCommunitiesRequest) (*model.GetPendingInviteCommunitiesResponse, error)
	ApproveInvitedCommunities(context.Context, *model.ApproveInvitedCommunitiesRequest) (*model.ApproveInvitedCommunitiesResponse, error)
	TransferCommunity(context.Context, *model.TransferCommunityRequest) (*model.TransferCommunityResponse, error)
}

type communityDomain struct {
	communityRepo         repository.CommunityRepository
	collaboratorRepo      repository.CollaboratorRepository
	userRepo              repository.UserRepository
	questRepo             repository.QuestRepository
	communityRoleVerifier *common.CommunityRoleVerifier
	globalRoleVerifier    *common.GlobalRoleVerifier
	discordEndpoint       discord.IEndpoint
	storage               storage.Storage
	oauth2Services        []authenticator.IOAuth2Service
}

func NewCommunityDomain(
	communityRepo repository.CommunityRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	questRepo repository.QuestRepository,
	discordEndpoint discord.IEndpoint,
	storage storage.Storage,
	oauth2Services []authenticator.IOAuth2Service,
) CommunityDomain {
	return &communityDomain{
		communityRepo:         communityRepo,
		collaboratorRepo:      collaboratorRepo,
		userRepo:              userRepo,
		questRepo:             questRepo,
		discordEndpoint:       discordEndpoint,
		communityRoleVerifier: common.NewCommunityRoleVerifier(collaboratorRepo, userRepo),
		globalRoleVerifier:    common.NewGlobalRoleVerifier(userRepo),
		storage:               storage,
		oauth2Services:        oauth2Services,
	}
}

func (d *communityDomain) Create(
	ctx context.Context, req *model.CreateCommunityRequest,
) (*model.CreateCommunityResponse, error) {
	if err := checkCommunityDisplayName(req.DisplayName); err != nil {
		return nil, err
	}

	if req.Handle != "" {
		if err := checkCommunityHandle(ctx, req.Handle); err != nil {
			return nil, err
		}

		_, err := d.communityRepo.GetByHandle(ctx, req.Handle)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get community by handle: %v", err)
				return nil, errorx.Unknown
			}

			return nil, errorx.New(errorx.AlreadyExists, "Duplicated handle")
		}
	} else {
		originHandle := generateCommunityHandle(req.DisplayName)
		handle := originHandle
		power := 2
		for {
			if checkCommunityHandle(ctx, handle) == nil {
				_, err := d.communityRepo.GetByHandle(ctx, handle)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					break
				} else if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot get community by handle: %v", err)
					return nil, errorx.Unknown
				}
			}

			// If the handle existed, we will append a random suffix to the
			// origin handle.
			suffix := crypto.RandIntn(int(math.Pow10(power)))
			handle = fmt.Sprintf("%s_%s", originHandle, strconv.Itoa(suffix))
			power++
			continue
		}

		req.Handle = handle
	}

	invitedBy := sql.NullString{Valid: false}
	if req.InviteCode != "" {
		inviteUser, err := d.userRepo.GetByInviteCode(ctx, req.InviteCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Invalid invite code")
			}

			xcontext.Logger(ctx).Errorf("Cannot get invite user: %v", err)
			return nil, errorx.Unknown
		}

		if inviteUser.ID == xcontext.RequestUserID(ctx) {
			return nil, errorx.New(errorx.BadRequest, "Cannot invited by yourself")
		}

		invitedBy = sql.NullString{Valid: true, String: inviteUser.ID}
	}

	userID := xcontext.RequestUserID(ctx)
	community := &entity.Community{
		Base:          entity.Base{ID: uuid.NewString()},
		Introduction:  []byte(req.Introduction),
		Handle:        req.Handle,
		DisplayName:   req.DisplayName,
		WebsiteURL:    req.WebsiteURL,
		Twitter:       req.Twitter,
		CreatedBy:     userID,
		InvitedBy:     invitedBy,
		InvitedStatus: entity.InvitedStatusUnclaimable,
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	if err := d.communityRepo.Create(ctx, community); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create community: %v", err)
		return nil, errorx.Unknown
	}

	err := d.collaboratorRepo.Upsert(ctx, &entity.Collaborator{
		UserID:      userID,
		CommunityID: community.ID,
		Role:        entity.Owner,
		CreatedBy:   userID,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot assign role owner: %v", err)
		return nil, errorx.Unknown
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.CreateCommunityResponse{Handle: community.Handle}, nil
}

func (d *communityDomain) GetList(
	ctx context.Context, req *model.GetCommunitiesRequest,
) (*model.GetCommunitiesResponse, error) {
	apiCfg := xcontext.Configs(ctx).ApiServer
	if req.Limit == 0 {
		req.Limit = apiCfg.DefaultLimit
	}

	if req.Limit == -1 {
		return nil, errorx.New(errorx.BadRequest, "Limit must be positive")
	}

	if req.Limit > apiCfg.MaxLimit {
		return nil, errorx.New(errorx.BadRequest, "Exceed the maximum of limit (%d)", apiCfg.MaxLimit)
	}

	result, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
		Q:          req.Q,
		Offset:     req.Offset,
		Limit:      req.Limit,
		ByTrending: req.ByTrending,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community list: %v", err)
		return nil, errorx.Unknown
	}

	communities := []model.Community{}
	for _, c := range result {
		totalQuests, err := d.questRepo.Count(
			ctx, repository.StatisticQuestFilter{CommunityID: c.ID})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot count quest of community %s: %v", c.ID, err)
			return nil, errorx.Unknown
		}

		communities = append(communities, convertCommunity(&c, int(totalQuests)))
	}

	return &model.GetCommunitiesResponse{Communities: communities}, nil
}

func (d *communityDomain) Get(
	ctx context.Context, req *model.GetCommunityRequest,
) (*model.GetCommunityResponse, error) {
	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get the community: %v", err)
		return nil, errorx.Unknown
	}

	totalQuests, err := d.questRepo.Count(
		ctx, repository.StatisticQuestFilter{CommunityID: community.ID})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot count quest of community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetCommunityResponse{
		Community: convertCommunity(community, int(totalQuests)),
	}, nil
}

func (d *communityDomain) UpdateByID(
	ctx context.Context, req *model.UpdateCommunityRequest,
) (*model.UpdateCommunityResponse, error) {
	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.communityRoleVerifier.Verify(ctx, community.ID, entity.Owner); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Only owner can update community")
	}

	if req.DisplayName != "" {
		if err := checkCommunityDisplayName(req.DisplayName); err != nil {
			xcontext.Logger(ctx).Debugf("Invalid display name: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid display name")
		}
	}

	err = d.communityRepo.UpdateByID(ctx, community.ID, entity.Community{
		DisplayName:  req.DisplayName,
		Introduction: []byte(req.Introduction),
		WebsiteURL:   req.WebsiteURL,
		Twitter:      req.Twitter,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update community: %v", err)
		return nil, errorx.Unknown
	}

	newCommunity, err := d.communityRepo.GetByID(ctx, community.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get new community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateCommunityResponse{Community: convertCommunity(newCommunity, 0)}, nil
}

func (d *communityDomain) UpdateDiscord(
	ctx context.Context, req *model.UpdateCommunityDiscordRequest,
) (*model.UpdateCommunityDiscordResponse, error) {
	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.communityRoleVerifier.Verify(ctx, community.ID, entity.Owner); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Only owner can update discord")
	}

	var service authenticator.IOAuth2Service
	for i := range d.oauth2Services {
		if d.oauth2Services[i].Service() == xcontext.Configs(ctx).Auth.Discord.Name {
			service = d.oauth2Services[i]
		}
	}

	if service == nil {
		xcontext.Logger(ctx).Errorf("Not setup discord oauth2 service")
		return nil, errorx.Unknown
	}

	var discordUserID string
	var oauth2Method string
	if req.AccessToken != "" {
		oauth2Method = "access token"
		discordUserID, err = service.GetUserID(ctx, req.AccessToken)
	} else if req.Code != "" {
		oauth2Method = "authorization code with pkce"
		discordUserID, err = service.VerifyAuthorizationCode(
			ctx, req.Code, req.CodeVerifier, req.RedirectURI)
	} else if req.IDToken != "" {
		oauth2Method = "id token"
		discordUserID, err = service.VerifyIDToken(ctx, req.IDToken)
	}

	if oauth2Method == "" {
		return nil, errorx.New(errorx.BadRequest, "Please provide at least one method to authorize")
	}

	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot verify %s: %v", oauth2Method, err)
		return nil, errorx.Unknown
	}

	guild, err := d.discordEndpoint.GetGuild(ctx, req.ServerID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get discord server: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid discord server")
	}

	tag, rawID, found := strings.Cut(discordUserID, "_")
	if !found || tag != xcontext.Configs(ctx).Auth.Discord.Name {
		xcontext.Logger(ctx).Errorf("Invalid discord user id in database")
		return nil, errorx.Unknown
	}

	if guild.OwnerID != rawID {
		return nil, errorx.New(errorx.PermissionDenied, "You are not server's owner")
	}

	hasAddedBot, err := d.discordEndpoint.HasAddedBot(ctx, req.ServerID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot check has added bot: %v", err)
		return nil, errorx.Unknown
	}

	if !hasAddedBot {
		return nil, errorx.New(errorx.Unavailable, "The server has not added bot yet")
	}

	err = d.communityRepo.UpdateByID(ctx, community.ID, entity.Community{Discord: guild.ID})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateCommunityDiscordResponse{}, nil
}

func (d *communityDomain) DeleteByID(
	ctx context.Context, req *model.DeleteCommunityRequest,
) (*model.DeleteCommunityResponse, error) {
	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.communityRoleVerifier.Verify(ctx, community.ID, entity.Owner); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Only owner can delete community")
	}

	if err := d.communityRepo.DeleteByID(ctx, community.ID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteCommunityResponse{}, nil
}

func (d *communityDomain) GetFollowing(
	ctx context.Context, req *model.GetFollowingCommunitiesRequest,
) (*model.GetFollowingCommunitiesResponse, error) {
	userID := xcontext.RequestUserID(ctx)
	result, err := d.communityRepo.GetFollowingList(ctx, userID, req.Offset, req.Limit)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community list: %v", err)
		return nil, errorx.Unknown
	}

	communities := []model.Community{}
	for _, c := range result {
		communities = append(communities, convertCommunity(&c, 0))
	}

	return &model.GetFollowingCommunitiesResponse{Communities: communities}, nil
}

func (d *communityDomain) UploadLogo(
	ctx context.Context, req *model.UploadCommunityLogoRequest,
) (*model.UploadCommunityLogoResponse, error) {
	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	image, err := common.ProcessImage(ctx, d.storage, "image")
	if err != nil {
		return nil, err
	}

	communityHandle := xcontext.HTTPRequest(ctx).PostFormValue("community_handle")
	community, err := d.communityRepo.GetByHandle(ctx, communityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.communityRoleVerifier.Verify(ctx, community.ID, entity.Owner); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	updatedCommunity := entity.Community{LogoPicture: image.Url}
	if err := d.communityRepo.UpdateByID(ctx, community.ID, updatedCommunity); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update community logo: %v", err)
		return nil, errorx.Unknown
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.UploadCommunityLogoResponse{}, nil
}

func (d *communityDomain) GetMyInvitedCommunities(
	ctx context.Context, req *model.GetMyInvitedCommunitiesRequest,
) (*model.GetInvitedCommunitiesResponse, error) {
	communities, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
		InvitedBy: xcontext.RequestUserID(ctx),
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get invited communities: %v", err)
		return nil, errorx.Unknown
	}

	numberOfClaimableCommunities := 0
	numberOfPendingCommunities := 0
	for _, p := range communities {
		if p.InvitedStatus == entity.InvitedStatusClaimable {
			numberOfClaimableCommunities++
		} else if p.InvitedStatus == entity.InvitedStatusPending {
			numberOfPendingCommunities++
		}
	}

	return &model.GetInvitedCommunitiesResponse{
		TotalClaimableCommunities: numberOfClaimableCommunities,
		TotalPendingCommunities:   numberOfPendingCommunities,
		RewardAmount:              xcontext.Configs(ctx).Quest.InviteCommunityRewardAmount,
	}, nil
}

func (d *communityDomain) GetPendingInvitedCommunities(
	ctx context.Context, req *model.GetPendingInviteCommunitiesRequest,
) (*model.GetPendingInviteCommunitiesResponse, error) {
	if err := d.globalRoleVerifier.Verify(ctx, entity.GlobalAdminRoles...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied to get pending invited communities: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	communities, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
		InvitedStatus: entity.InvitedStatusPending,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get invited communities: %v", err)
		return nil, errorx.Unknown
	}

	invitedCommunities := []model.Community{}
	for _, c := range communities {
		invitedCommunities = append(invitedCommunities, convertCommunity(&c, 0))
	}

	return &model.GetPendingInviteCommunitiesResponse{Communities: invitedCommunities}, nil
}

func (d *communityDomain) ApproveInvitedCommunities(
	ctx context.Context, req *model.ApproveInvitedCommunitiesRequest,
) (*model.ApproveInvitedCommunitiesResponse, error) {
	if err := d.globalRoleVerifier.Verify(ctx, entity.GlobalAdminRoles...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission deined to approve invited community: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	communities, err := d.communityRepo.GetByHandles(ctx, req.CommunityHandles)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get invited communities: %v", err)
		return nil, errorx.Unknown
	}

	for _, p := range communities {
		if p.InvitedStatus != entity.InvitedStatusPending {
			return nil, errorx.New(errorx.BadRequest, "Community %s cannot available to approve", p.ID)
		}
	}

	err = d.communityRepo.UpdateInvitedStatusByHandles(ctx, req.CommunityHandles, entity.InvitedStatusClaimable)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update invited status by ids: %v", err)
		return nil, errorx.Unknown
	}

	return &model.ApproveInvitedCommunitiesResponse{}, nil
}

func (d *communityDomain) TransferCommunity(ctx context.Context, req *model.TransferCommunityRequest) (*model.TransferCommunityResponse, error) {
	if err := d.globalRoleVerifier.Verify(ctx, entity.GlobalAdminRoles...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission deined to transfer community: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if _, err := d.userRepo.GetByID(ctx, req.ToID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found user")
		}

		xcontext.Logger(ctx).Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	if err := d.collaboratorRepo.DeleteOldOwnerByCommunityID(ctx, community.ID); err != nil {
		return nil, errorx.Unknown
	}

	if err := d.collaboratorRepo.Upsert(ctx, &entity.Collaborator{
		UserID:      req.ToID,
		CommunityID: community.ID,
		Role:        entity.Owner,
		CreatedBy:   xcontext.RequestUserID(ctx),
	}); err != nil {
		return nil, errorx.Unknown
	}
	xcontext.WithCommitDBTransaction(ctx)

	return &model.TransferCommunityResponse{}, nil
}
