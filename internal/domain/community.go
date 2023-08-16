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
	"time"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

type CommunityDomain interface {
	Create(context.Context, *model.CreateCommunityRequest) (*model.CreateCommunityResponse, error)
	GetList(context.Context, *model.GetCommunitiesRequest) (*model.GetCommunitiesResponse, error)
	GetListPending(context.Context, *model.GetPendingCommunitiesRequest) (*model.GetPendingCommunitiesResponse, error)
	Get(context.Context, *model.GetCommunityRequest) (*model.GetCommunityResponse, error)
	GetMyOwn(context.Context, *model.GetMyOwnCommunitiesRequest) (*model.GetMyOwnCommunitiesResponse, error)
	UpdateByID(context.Context, *model.UpdateCommunityRequest) (*model.UpdateCommunityResponse, error)
	UpdateDiscord(context.Context, *model.UpdateCommunityDiscordRequest) (*model.UpdateCommunityDiscordResponse, error)
	DeleteByID(context.Context, *model.DeleteCommunityRequest) (*model.DeleteCommunityResponse, error)
	UploadLogo(context.Context, *model.UploadCommunityLogoRequest) (*model.UploadCommunityLogoResponse, error)
	GetMyReferral(context.Context, *model.GetMyReferralRequest) (*model.GetMyReferralResponse, error)
	GetReferral(context.Context, *model.GetReferralRequest) (*model.GetReferralResponse, error)
	ReviewReferral(context.Context, *model.ReviewReferralRequest) (*model.ReviewReferralResponse, error)
	TransferCommunity(context.Context, *model.TransferCommunityRequest) (*model.TransferCommunityResponse, error)
	ReviewPending(context.Context, *model.ReviewPendingCommunityRequest) (*model.ReviewPendingCommunityResponse, error)
	GetDiscordRole(context.Context, *model.GetDiscordRoleRequest) (*model.GetDiscordRoleResponse, error)
	AssignRole(context.Context, *model.AssignRoleRequest) (*model.AssignRoleResponse, error)
	DeleteUserCommunityRole(context.Context, *model.DeleteUserCommunityRoleRequest) (*model.DeleteUserCommunityRoleResponse, error)
}

type communityDomain struct {
	communityRepo            repository.CommunityRepository
	followerRepo             repository.FollowerRepository
	followerRoleRepo         repository.FollowerRoleRepository
	userRepo                 repository.UserRepository
	questRepo                repository.QuestRepository
	oauth2Repo               repository.OAuth2Repository
	chatChannelRepo          repository.ChatChannelRepository
	communityRoleVerifier    *common.CommunityRoleVerifier
	globalRoleVerifier       *common.GlobalRoleVerifier
	discordEndpoint          discord.IEndpoint
	storage                  storage.Storage
	oauth2Services           []authenticator.IOAuth2Service
	roleRepo                 repository.RoleRepository
	notificationEngineCaller client.NotificationEngineCaller
	redisClient              xredis.Client
}

func NewCommunityDomain(
	communityRepo repository.CommunityRepository,
	followerRepo repository.FollowerRepository,
	followerRoleRepo repository.FollowerRoleRepository,
	userRepo repository.UserRepository,
	questRepo repository.QuestRepository,
	oauth2Repo repository.OAuth2Repository,
	chatChannelRepo repository.ChatChannelRepository,
	roleRepo repository.RoleRepository,
	discordEndpoint discord.IEndpoint,
	storage storage.Storage,
	oauth2Services []authenticator.IOAuth2Service,
	notificationEngineCaller client.NotificationEngineCaller,
	communityRoleVerifier *common.CommunityRoleVerifier,
	redisClient xredis.Client,
) CommunityDomain {
	return &communityDomain{
		communityRepo:            communityRepo,
		followerRepo:             followerRepo,
		followerRoleRepo:         followerRoleRepo,
		userRepo:                 userRepo,
		questRepo:                questRepo,
		oauth2Repo:               oauth2Repo,
		chatChannelRepo:          chatChannelRepo,
		roleRepo:                 roleRepo,
		discordEndpoint:          discordEndpoint,
		communityRoleVerifier:    communityRoleVerifier,
		globalRoleVerifier:       common.NewGlobalRoleVerifier(userRepo),
		storage:                  storage,
		oauth2Services:           oauth2Services,
		notificationEngineCaller: notificationEngineCaller,
		redisClient:              redisClient,
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

	walletNonce, err := crypto.GenerateRandomString()
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate wallet nonce: %v", err)
		return nil, errorx.Unknown
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
		WalletNonce:    walletNonce,
		OwnerEmail:     req.OwnerEmail,
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

	err = d.followerRoleRepo.Create(ctx, &entity.FollowerRole{
		UserID:      userID,
		CommunityID: community.ID,
		RoleID:      entity.OwnerBaseRole,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create owner of community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.chatChannelRepo.Create(ctx, &entity.ChatChannel{
		SnowFlakeBase: entity.SnowFlakeBase{ID: xcontext.SnowFlake(ctx).Generate().Int64()},
		CommunityID:   community.ID,
		Name:          "general",
	}); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create general channel: %v", err)
		return nil, errorx.Unknown
	}

	ctx = xcontext.WithCommitDBTransaction(ctx)

	err = FollowCommunity(
		ctx, d.userRepo, d.communityRepo, d.followerRepo, d.followerRoleRepo,
		d.notificationEngineCaller, d.redisClient, userID, community.ID, "",
	)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot follow community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateCommunityResponse{Handle: community.Handle}, nil
}

func (d *communityDomain) GetList(
	ctx context.Context, req *model.GetCommunitiesRequest,
) (*model.GetCommunitiesResponse, error) {
	// 1. Get all suitable active communities and order by trending score.
	result, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
		Q:          req.Q,
		ByTrending: req.ByTrending,
		Status:     entity.CommunityActive,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community list: %v", err)
		return nil, errorx.Unknown
	}

	communityIDs := []string{}
	for _, c := range result {
		communityIDs = append(communityIDs, c.ID)
	}

	// 2. Check the user is platform's admin or super admin. If it is, this
	// method will include sensitive information of community owners.
	requestUser, err := d.userRepo.GetByID(ctx, xcontext.RequestUserID(ctx))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get the request user: %v", err)
		return nil, errorx.Unknown
	}

	isAdmin := slices.Contains(entity.GlobalAdminRoles, requestUser.Role)

	communityToOwnerUserID := map[string]string{}
	ownerUserMap := map[string]entity.User{}
	oauth2Map := map[string][]entity.OAuth2{}
	// 2a. Get the community owners information only if the request user is
	// admin or super admin.
	if isAdmin {
		// Get owners of all communities.
		owners, err := d.followerRoleRepo.GetOwnerByCommunityIDs(ctx, communityIDs...)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get owners of community list: %v", err)
			return nil, errorx.Unknown
		}

		ownerUserIDs := []string{}
		for _, owner := range owners {
			ownerUserIDs = append(ownerUserIDs, owner.UserID)
			communityToOwnerUserID[owner.CommunityID] = owner.UserID
		}

		// Get owner basic info.
		ownerUsers, err := d.userRepo.GetByIDs(ctx, ownerUserIDs)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get owners info of pending community list: %v", err)
			return nil, errorx.Unknown
		}

		for _, u := range ownerUsers {
			ownerUserMap[u.ID] = u
		}

		// Get owner linked services info.
		ownerOAuth2Records, err := d.oauth2Repo.GetAllByUserIDs(ctx, ownerUserIDs...)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get owners oauth2 records of pending community list: %v", err)
			return nil, errorx.Unknown
		}

		for _, oauth2 := range ownerOAuth2Records {
			oauth2Map[oauth2.UserID] = append(oauth2Map[oauth2.UserID], oauth2)
		}
	}

	// 3. Convert community entities to response.
	communities := []model.Community{}
	for _, c := range result {
		totalQuests, err := d.questRepo.Count(
			ctx, repository.StatisticQuestFilter{CommunityID: c.ID})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot count quest of community %s: %v", c.ID, err)
			return nil, errorx.Unknown
		}
		clientCommunity := model.ConvertCommunity(&c, int(totalQuests))

		if isAdmin {
			clientCommunity.OwnerEmail = c.OwnerEmail

			ownerUserID, ok := communityToOwnerUserID[c.ID]
			if !ok {
				xcontext.Logger(ctx).Errorf("Not found owner user ID of community %s", c.ID)
				return nil, errorx.Unknown
			}

			owner, ok := ownerUserMap[ownerUserID]
			if !ok {
				xcontext.Logger(ctx).Errorf("Not found owner of community %s in owner map", c.ID)
				return nil, errorx.Unknown
			}

			oauth2 := oauth2Map[owner.ID]
			clientCommunity.Owner = model.ConvertUser(&owner, oauth2, true, "")
		}

		communities = append(communities, clientCommunity)
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

	communityIDs := []string{}
	for _, c := range result {
		communityIDs = append(communityIDs, c.ID)
	}

	owners, err := d.followerRoleRepo.GetOwnerByCommunityIDs(ctx, communityIDs...)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get owners of pending community list: %v", err)
		return nil, errorx.Unknown
	}

	ownerUserIDs := []string{}
	communityToOwnerUserID := map[string]string{}
	for _, owner := range owners {
		ownerUserIDs = append(ownerUserIDs, owner.UserID)
		communityToOwnerUserID[owner.CommunityID] = owner.UserID
	}

	ownerUsers, err := d.userRepo.GetByIDs(ctx, ownerUserIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get owners info of pending community list: %v", err)
		return nil, errorx.Unknown
	}

	ownerUserMap := map[string]entity.User{}
	for _, u := range ownerUsers {
		ownerUserMap[u.ID] = u
	}

	ownerOAuth2Records, err := d.oauth2Repo.GetAllByUserIDs(ctx, ownerUserIDs...)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get owners oauth2 records of pending community list: %v", err)
		return nil, errorx.Unknown
	}

	oauth2Map := map[string][]entity.OAuth2{}
	for _, oauth2 := range ownerOAuth2Records {
		oauth2Map[oauth2.UserID] = append(oauth2Map[oauth2.UserID], oauth2)
	}

	communities := []model.Community{}
	for _, c := range result {
		clientCommunity := model.ConvertCommunity(&c, 0)

		// Only this API is allowed including owner email.
		clientCommunity.OwnerEmail = c.OwnerEmail

		ownerUserID, ok := communityToOwnerUserID[c.ID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found owner user ID of community %s", c.ID)
			return nil, errorx.Unknown
		}

		owner, ok := ownerUserMap[ownerUserID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found owner of community %s in owner map", c.ID)
			return nil, errorx.Unknown
		}

		oauth2 := oauth2Map[owner.ID]
		clientCommunity.Owner = model.ConvertUser(&owner, oauth2, true, "")
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
		Community: model.ConvertCommunity(community, int(totalQuests)),
	}, nil
}

func (d *communityDomain) GetMyOwn(
	ctx context.Context, req *model.GetMyOwnCommunitiesRequest,
) (*model.GetMyOwnCommunitiesResponse, error) {
	followerRoles, err := d.followerRoleRepo.GetOwnersByUserID(ctx, xcontext.RequestUserID(ctx))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get follower role: %v", err)
		return nil, errorx.Unknown
	}

	if len(followerRoles) == 0 {
		return &model.GetMyOwnCommunitiesResponse{}, nil
	}

	communityIDs := []string{}
	for _, fr := range followerRoles {
		communityIDs = append(communityIDs, fr.CommunityID)
	}

	communities, err := d.communityRepo.GetByIDs(ctx, communityIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community role: %v", err)
		return nil, errorx.Unknown
	}

	clientCommunities := []model.Community{}
	for _, c := range communities {
		clientCommunities = append(clientCommunities, model.ConvertCommunity(&c, 0))
	}

	return &model.GetMyOwnCommunitiesResponse{Communities: clientCommunities}, nil
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

	if community.Discord == "" && req.DiscordInviteLink != "" {
		return nil, errorx.New(errorx.Unavailable, "Please connect to discord before setup invite link")
	}

	if req.DiscordInviteLink != "" {
		inviteCode, err := common.ParseInviteDiscordURL(req.DiscordInviteLink)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Invalid invite link: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid discord invite link")
		}

		code, err := d.discordEndpoint.GetCode(ctx, community.Discord, inviteCode)
		if err != nil {
			xcontext.Logger(ctx).Warnf("Cannot get invite code info: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Not found the invite link in your discord server")
		}

		if code.MaxAge != 0 && code.CreatedAt.Add(code.MaxAge).Before(time.Now()) {
			return nil, errorx.New(errorx.Unavailable, "Discord invite link is expired")
		}

		if code.MaxUses != 0 && code.Uses >= code.MaxUses {
			return nil, errorx.New(errorx.Unavailable, "Discord invite link exceeded the max uses")
		}
	}

	if err := d.communityRoleVerifier.Verify(ctx, community.ID); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Only owner can update community")
	}

	if req.DisplayName != "" {
		if err := checkCommunityDisplayName(req.DisplayName); err != nil {
			xcontext.Logger(ctx).Debugf("Invalid display name: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid display name")
		}
	}

	err = d.communityRepo.UpdateByID(ctx, community.ID, entity.Community{
		DisplayName:       req.DisplayName,
		Introduction:      []byte(req.Introduction),
		WebsiteURL:        req.WebsiteURL,
		Twitter:           req.Twitter,
		DiscordInviteLink: req.DiscordInviteLink,
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

	return &model.UpdateCommunityResponse{Community: model.ConvertCommunity(newCommunity, 0)}, nil
}

func (d *communityDomain) ReviewPending(
	ctx context.Context, req *model.ReviewPendingCommunityRequest,
) (*model.ReviewPendingCommunityResponse, error) {
	status, err := enum.ToEnum[entity.CommunityStatus](req.Status)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid status: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid status %s", req.Status)
	}

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

	if community.Status == entity.CommunityRejected {
		return nil, errorx.New(errorx.Unavailable, "Community has been already rejected")
	}

	err = d.communityRepo.UpdateByID(ctx, community.ID, entity.Community{Status: status})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.ReviewPendingCommunityResponse{}, nil
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

	if err := d.communityRoleVerifier.Verify(ctx, community.ID); err != nil {
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

	if err := d.communityRoleVerifier.Verify(ctx, community.ID); err != nil {
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
		totalQuests, err := d.questRepo.Count(
			ctx, repository.StatisticQuestFilter{CommunityID: c.ID})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot count quest of community %s: %v", c.ID, err)
			return nil, errorx.Unknown
		}

		communities = append(communities, model.ConvertCommunity(&c, int(totalQuests)))
	}

	return &model.GetFollowingCommunitiesResponse{Communities: communities}, nil
}

func (d *communityDomain) UploadLogo(
	ctx context.Context, req *model.UploadCommunityLogoRequest,
) (*model.UploadCommunityLogoResponse, error) {
	communityHandle := xcontext.HTTPRequest(ctx).PostFormValue("community_handle")
	community, err := d.communityRepo.GetByHandle(ctx, communityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	image, err := common.ProcessFormDataImage(ctx, d.storage, "image")
	if err != nil {
		return nil, err
	}

	if err := d.communityRoleVerifier.Verify(ctx, community.ID); err != nil {
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
		communitiesByReferralUser[key] = append(communitiesByReferralUser[key], model.ConvertCommunity(&c, 0))
	}

	referrals := []model.Referral{}
	for referredBy, communities := range communitiesByReferralUser {
		referredByUser, ok := referredUserMap[referredBy]
		if !ok {
			xcontext.Logger(ctx).Errorf("Invalid referred user %s: %v", referredBy, err)
		}

		oauth2Servies, err := d.oauth2Repo.GetAllByUserIDs(ctx, referredBy)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get all oauth2 services: %v", err)
			return nil, errorx.Unknown
		}

		referrals = append(referrals, model.Referral{
			ReferredBy:  model.ConvertUser(referredByUser, oauth2Servies, false, ""),
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

	err = d.communityRepo.UpdateReferralStatusByID(ctx, community.ID, referralStatus)
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

	ownerFollowerRole, err := d.followerRoleRepo.GetFirstByRole(ctx, community.ID, entity.OwnerBaseRole)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get owner role: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.globalRoleVerifier.Verify(ctx, entity.GlobalAdminRoles...); err != nil {
		if ownerFollowerRole.UserID != xcontext.RequestUserID(ctx) {
			return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
		}
	}

	if _, err := d.userRepo.GetByID(ctx, req.ToUserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found user")
		}

		xcontext.Logger(ctx).Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	err = d.followerRoleRepo.Delete(ctx, ownerFollowerRole.UserID, community.ID, entity.OwnerBaseRole)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete owner role: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.followerRoleRepo.Create(ctx, &entity.FollowerRole{
		UserID:      ownerFollowerRole.UserID,
		CommunityID: community.ID,
		RoleID:      entity.UserBaseRole,
	}); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update owner to user: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.followerRoleRepo.Create(ctx, &entity.FollowerRole{
		UserID:      req.ToUserID,
		CommunityID: community.ID,
		RoleID:      entity.OwnerBaseRole,
	}); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update new owner: %v", err)
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
	if err := d.communityRoleVerifier.Verify(ctx, community.ID); err != nil {
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
		isSuitablePosition := req.IncludeAll || role.Position < botRolePosition
		if isSuitablePosition && role.Name != "@everyone" && role.BotID == "" {
			clientRoles = append(clientRoles, model.ConvertDiscordRole(role))
		}
	}

	return &model.GetDiscordRoleResponse{Roles: clientRoles}, nil
}

func (d *communityDomain) AssignRole(ctx context.Context, req *model.AssignRoleRequest) (*model.AssignRoleResponse, error) {
	if slices.Contains([]string{entity.OwnerBaseRole, entity.UserBaseRole}, req.RoleID) {
		return nil, errorx.New(errorx.Unavailable, "Unable to assign base role")
	}

	if xcontext.RequestUserID(ctx) == req.UserID {
		return nil, errorx.New(errorx.PermissionDenied, "Can not assign by yourself")
	}

	role, err := d.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get role: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.communityRoleVerifier.Verify(ctx, role.CommunityID.String, req.RoleID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if err := d.followerRoleRepo.Create(ctx, &entity.FollowerRole{
		UserID:      req.UserID,
		CommunityID: role.CommunityID.String,
		RoleID:      req.RoleID,
	}); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot assign role for community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.AssignRoleResponse{}, nil
}

func (d *communityDomain) DeleteUserCommunityRole(ctx context.Context, req *model.DeleteUserCommunityRoleRequest) (*model.DeleteUserCommunityRoleResponse, error) {
	if xcontext.RequestUserID(ctx) == req.UserID {
		return nil, errorx.New(errorx.PermissionDenied, "Can not delete role by yourself")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	roles, err := d.roleRepo.GetByIDs(ctx, req.RoleIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get role: %v", err)
		return nil, errorx.Unknown
	}

	for _, role := range roles {
		if !role.CommunityID.Valid {
			return nil, errorx.New(errorx.BadRequest, "Cannot delete base roles")
		}

		if role.CommunityID.String != community.ID {
			return nil, errorx.New(errorx.BadRequest, "Role %s not exists in community", role.Name)
		}
	}

	if err := d.communityRoleVerifier.Verify(ctx, community.ID, req.RoleIDs...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if err := d.followerRoleRepo.DeleteByRoles(ctx, req.UserID,
		community.ID,
		req.RoleIDs,
	); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete user role for community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteUserCommunityRoleResponse{}, nil
}
