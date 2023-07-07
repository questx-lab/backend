package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/mail"
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
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

type CommunityDomain interface {
	Create(context.Context, *model.CreateCommunityRequest) (*model.CreateCommunityResponse, error)
	GetList(context.Context, *model.GetCommunitiesRequest) (*model.GetCommunitiesResponse, error)
	GetListPending(context.Context, *model.GetPendingCommunitiesRequest) (*model.GetPendingCommunitiesResponse, error)
	Get(context.Context, *model.GetCommunityRequest) (*model.GetCommunityResponse, error)
	UpdateByID(context.Context, *model.UpdateCommunityRequest) (*model.UpdateCommunityResponse, error)
	UpdateDiscord(context.Context, *model.UpdateCommunityDiscordRequest) (*model.UpdateCommunityDiscordResponse, error)
	DeleteByID(context.Context, *model.DeleteCommunityRequest) (*model.DeleteCommunityResponse, error)
	UploadLogo(context.Context, *model.UploadCommunityLogoRequest) (*model.UploadCommunityLogoResponse, error)
	GetMyReferral(context.Context, *model.GetMyReferralRequest) (*model.GetMyReferralResponse, error)
	GetReferral(context.Context, *model.GetReferralRequest) (*model.GetReferralResponse, error)
	ReviewReferral(context.Context, *model.ReviewReferralRequest) (*model.ReviewReferralResponse, error)
	TransferCommunity(context.Context, *model.TransferCommunityRequest) (*model.TransferCommunityResponse, error)
	ApprovePending(context.Context, *model.ApprovePendingCommunityRequest) (*model.ApprovePendingCommunityRequest, error)
	GetDiscordRole(context.Context, *model.GetDiscordRoleRequest) (*model.GetDiscordRoleResponse, error)
}

type communityDomain struct {
	communityRepo         repository.CommunityRepository
	collaboratorRepo      repository.CollaboratorRepository
	userRepo              repository.UserRepository
	questRepo             repository.QuestRepository
	oauth2Repo            repository.OAuth2Repository
	gameRepo              repository.GameRepository
	communityRoleVerifier *common.CommunityRoleVerifier
	discordEndpoint       discord.IEndpoint
	storage               storage.Storage
	publisher             pubsub.Publisher
	oauth2Services        []authenticator.IOAuth2Service
}

func NewCommunityDomain(
	communityRepo repository.CommunityRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	questRepo repository.QuestRepository,
	oauth2Repo repository.OAuth2Repository,
	gameRepo repository.GameRepository,
	discordEndpoint discord.IEndpoint,
	storage storage.Storage,
	publisher pubsub.Publisher,
	oauth2Services []authenticator.IOAuth2Service,
) CommunityDomain {
	return &communityDomain{
		communityRepo:         communityRepo,
		collaboratorRepo:      collaboratorRepo,
		userRepo:              userRepo,
		questRepo:             questRepo,
		oauth2Repo:            oauth2Repo,
		gameRepo:              gameRepo,
		discordEndpoint:       discordEndpoint,
		communityRoleVerifier: common.NewCommunityRoleVerifier(collaboratorRepo, userRepo),
		storage:               storage,
		publisher:             publisher,
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

	referredBy := sql.NullString{Valid: false}
	if req.ReferralCode != "" {
		referralUser, err := d.userRepo.GetByReferralCode(ctx, req.ReferralCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Invalid referral code")
			}

			xcontext.Logger(ctx).Errorf("Cannot get referral user: %v", err)
			return nil, errorx.Unknown
		}

		if referralUser.ID == xcontext.RequestUserID(ctx) {
			return nil, errorx.New(errorx.BadRequest, "Cannot refer by yourself")
		}

		referredBy = sql.NullString{Valid: true, String: referralUser.ID}
	}

	userID := xcontext.RequestUserID(ctx)
	community := &entity.Community{
		Base:           entity.Base{ID: uuid.NewString()},
		Introduction:   []byte(req.Introduction),
		Handle:         req.Handle,
		DisplayName:    req.DisplayName,
		WebsiteURL:     req.WebsiteURL,
		Twitter:        req.Twitter,
		CreatedBy:      userID,
		ReferredBy:     referredBy,
		ReferralStatus: entity.ReferralUnclaimable,
		Status:         entity.CommunityActive,
	}

	if req.OwnerEmail != "" {
		_, err := mail.ParseAddress(req.OwnerEmail)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Cannot validate owner email address: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid email address of owner")
		}
	}

	if xcontext.Configs(ctx).ApiServer.NeedApproveCommunity {
		if req.OwnerEmail == "" {
			return nil, errorx.New(errorx.Unavailable,
				"We need your email address to contact when your community is approved")
		}

		community.Status = entity.CommunityPending
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

	firstMap, err := d.gameRepo.GetFirstMap(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Not found the first map: %v", err)
		return nil, errorx.New(errorx.Internal, "We has not setup the first map yet")
	}

	room := entity.GameRoom{
		Base:        entity.Base{ID: uuid.NewString()},
		CommunityID: community.ID,
		MapID:       firstMap.ID,
		Name:        fmt.Sprintf("%s-%d", community.Handle, crypto.RandRange(100, 999)),
	}
	if err := d.gameRepo.CreateRoom(ctx, &room); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create room: %v", err)
		return nil, errorx.Unknown
	}

	xcontext.WithCommitDBTransaction(ctx)

	err = d.publisher.Publish(ctx, model.CreateRoomTopic, &pubsub.Pack{
		Key: []byte(room.ID),
		Msg: []byte{},
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot publish create community event: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateCommunityResponse{Handle: community.Handle}, nil
}

func (d *communityDomain) GetList(
	ctx context.Context, req *model.GetCommunitiesRequest,
) (*model.GetCommunitiesResponse, error) {
	result, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
		Q:          req.Q,
		ByTrending: req.ByTrending,
		Status:     entity.CommunityActive,
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

func (d *communityDomain) GetListPending(
	ctx context.Context, req *model.GetPendingCommunitiesRequest,
) (*model.GetPendingCommunitiesResponse, error) {
	result, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
		Status: entity.CommunityPending,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get pending community list: %v", err)
		return nil, errorx.Unknown
	}

	communities := []model.Community{}
	for _, c := range result {
		clientCommunity := convertCommunity(&c, 0)
		clientCommunity.OwnerEmail = c.OwnerEmail
		communities = append(communities, clientCommunity)
	}

	return &model.GetPendingCommunitiesResponse{Communities: communities}, nil
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

func (d *communityDomain) ApprovePending(
	ctx context.Context, req *model.ApprovePendingCommunityRequest,
) (*model.ApprovePendingCommunityRequest, error) {
	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if community.Status == entity.CommunityActive {
		return nil, errorx.New(errorx.Unavailable, "Community has been already approved")
	}

	err = d.communityRepo.UpdateByID(ctx, community.ID, entity.Community{Status: entity.CommunityActive})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.ApprovePendingCommunityRequest{}, nil
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

	var discordUser authenticator.OAuth2User
	var oauth2Method string
	if req.AccessToken != "" {
		oauth2Method = "access token"
		discordUser, err = service.GetUserID(ctx, req.AccessToken)
	} else if req.Code != "" {
		oauth2Method = "authorization code with pkce"
		discordUser, err = service.VerifyAuthorizationCode(
			ctx, req.Code, req.CodeVerifier, req.RedirectURI)
	} else if req.IDToken != "" {
		oauth2Method = "id token"
		discordUser, err = service.VerifyIDToken(ctx, req.IDToken)
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

	tag, rawID, found := strings.Cut(discordUser.ID, "_")
	if !found || tag != xcontext.Configs(ctx).Auth.Discord.Name {
		xcontext.Logger(ctx).Errorf("Invalid discord user id: %s", discordUser.ID)
		return nil, errorx.Unknown
	}

	if guild.OwnerID != rawID {
		member, err := d.discordEndpoint.GetMember(ctx, req.ServerID, rawID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get discord member: %v", err)
			return nil, errorx.Unknown
		}

		if member.ID == "" {
			return nil, errorx.New(errorx.Unavailable, "The user has not joined in server")
		}

		roles, err := d.discordEndpoint.GetRoles(ctx, req.ServerID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get discord server roles: %v", err)
			return nil, errorx.Unknown
		}

		isAdmin := false
		for _, userRoleID := range member.RoleIDs {
			for _, serverRole := range roles {
				if userRoleID == serverRole.ID {
					if serverRole.Permissions&discord.AdministratorRoleFlag != 0 {
						isAdmin = true
						break
					}
				}
			}
		}

		if !isAdmin {
			return nil, errorx.New(errorx.PermissionDenied, "You are not server's owner or admin")
		}
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

	collaborators, err := d.collaboratorRepo.GetListByUserID(ctx, userID, 0, -1)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get collaborator list: %v", err)
		return nil, errorx.Unknown
	}

	collaboratedCommunityIDs := []string{}
	for _, c := range collaborators {
		collaboratedCommunityIDs = append(collaboratedCommunityIDs, c.CommunityID)
	}

	communities := []model.Community{}
	for _, c := range result {
		// Ignore community which this user is collaborated.
		if slices.Contains(collaboratedCommunityIDs, c.ID) {
			continue
		}

		totalQuests, err := d.questRepo.Count(
			ctx, repository.StatisticQuestFilter{CommunityID: c.ID})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot count quest of community %s: %v", c.ID, err)
			return nil, errorx.Unknown
		}

		communities = append(communities, convertCommunity(&c, int(totalQuests)))
	}

	return &model.GetFollowingCommunitiesResponse{Communities: communities}, nil
}

func (d *communityDomain) UploadLogo(
	ctx context.Context, req *model.UploadCommunityLogoRequest,
) (*model.UploadCommunityLogoResponse, error) {
	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	image, err := common.ProcessFormDataImage(ctx, d.storage, "image")
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

func (d *communityDomain) GetMyReferral(
	ctx context.Context, req *model.GetMyReferralRequest,
) (*model.GetMyReferralResponse, error) {
	communities, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
		ReferredBy: xcontext.RequestUserID(ctx),
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get referral communities: %v", err)
		return nil, errorx.Unknown
	}

	numberOfClaimableCommunities := 0
	numberOfPendingCommunities := 0
	for _, p := range communities {
		if p.ReferralStatus == entity.ReferralClaimable {
			numberOfClaimableCommunities++
		} else if p.ReferralStatus == entity.ReferralPending {
			numberOfPendingCommunities++
		}
	}

	return &model.GetMyReferralResponse{
		TotalClaimableCommunities: numberOfClaimableCommunities,
		TotalPendingCommunities:   numberOfPendingCommunities,
		RewardAmount:              xcontext.Configs(ctx).Quest.InviteCommunityRewardAmount,
	}, nil
}

func (d *communityDomain) GetReferral(
	ctx context.Context, req *model.GetReferralRequest,
) (*model.GetReferralResponse, error) {
	communities, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
		OrderByReferredBy: true,
		ReferralStatus: []entity.ReferralStatusType{
			entity.ReferralPending,
			entity.ReferralClaimable,
		},
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get referral communities: %v", err)
		return nil, errorx.Unknown
	}

	referredUserMap := map[string]*entity.User{}
	for _, c := range communities {
		referredUserMap[c.ReferredBy.String] = nil
	}

	referralUsers, err := d.userRepo.GetByIDs(ctx, common.MapKeys(referredUserMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get list referred users: %v", err)
		return nil, errorx.Unknown
	}

	for i := range referralUsers {
		referredUserMap[referralUsers[i].ID] = &referralUsers[i]
	}

	communitiesByReferralUser := map[string][]model.Community{}
	for _, c := range communities {
		key := c.ReferredBy.String
		communitiesByReferralUser[key] = append(communitiesByReferralUser[key], convertCommunity(&c, 0))
	}

	referrals := []model.Referral{}
	for referredBy, communities := range communitiesByReferralUser {
		referredByUser, ok := referredUserMap[referredBy]
		if !ok {
			xcontext.Logger(ctx).Errorf("Invalid referred user %s: %v", referredBy, err)
		}

		oauth2Servies, err := d.oauth2Repo.GetAllByUserID(ctx, referredBy)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get all oauth2 services: %v", err)
			return nil, errorx.Unknown
		}

		referrals = append(referrals, model.Referral{
			ReferredBy:  convertUser(referredByUser, oauth2Servies, false),
			Communities: communities,
		})
	}

	return &model.GetReferralResponse{Referrals: referrals}, nil
}

func (d *communityDomain) ReviewReferral(
	ctx context.Context, req *model.ReviewReferralRequest,
) (*model.ReviewReferralResponse, error) {
	var referralStatus entity.ReferralStatusType
	if req.Action == model.ReviewReferralActionApprove {
		referralStatus = entity.ReferralClaimable
	} else if req.Action == model.ReviewReferralActionReject {
		referralStatus = entity.ReferralRejected
	} else {
		return nil, errorx.New(errorx.BadRequest, "Invalid action %s", req.Action)
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get referral community: %v", err)
		return nil, errorx.Unknown
	}

	if community.ReferralStatus != entity.ReferralPending {
		return nil, errorx.New(errorx.BadRequest, "Community is not pending status of referral")
	}

	err = d.communityRepo.UpdateReferralStatusByIDs(ctx, []string{community.ID}, referralStatus)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update referral status by ids: %v", err)
		return nil, errorx.Unknown
	}

	return &model.ReviewReferralResponse{}, nil
}

func (d *communityDomain) TransferCommunity(ctx context.Context, req *model.TransferCommunityRequest) (*model.TransferCommunityResponse, error) {
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

func (d *communityDomain) GetDiscordRole(
	ctx context.Context, req *model.GetDiscordRoleRequest,
) (*model.GetDiscordRoleResponse, error) {
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

	if community.Discord == "" {
		return nil, errorx.New(errorx.Unavailable, "Community must connect to discord server first")
	}

	// Only owner or editor can get discord roles.
	if err := d.communityRoleVerifier.Verify(ctx, community.ID, entity.AdminGroup...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	roles, err := d.discordEndpoint.GetRoles(ctx, community.Discord)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get roles: %v", err)
		return nil, errorx.Unknown
	}

	botRolePosition := -1
	for _, role := range roles {
		if role.BotID == xcontext.Configs(ctx).Quest.Dicord.BotID {
			botRolePosition = role.Position
		}
	}

	if botRolePosition == -1 {
		return nil, errorx.New(errorx.Unavailable, "Not found questx bot in your discord server")
	}

	clientRoles := []model.DiscordRole{}
	for _, role := range roles {
		if role.Position < botRolePosition && role.Name != "@everyone" && role.BotID == "" {
			clientRoles = append(clientRoles, convertDiscordRole(role))
		}
	}

	return &model.GetDiscordRoleResponse{Roles: clientRoles}, nil
}
