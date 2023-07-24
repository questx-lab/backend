package main

import (
	"context"
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startApi(*cli.Context) error {
	cfg := xcontext.Configs(s.ctx)
	rpcSearchClient, err := rpc.DialContext(s.ctx, cfg.SearchServer.Endpoint)
	if err != nil {
		return err
	}

	rpcGameCenterClient, err := rpc.DialContext(s.ctx, cfg.GameCenterServer.Endpoint)
	if err != nil {
		return err
	}

	rpcNotificationEngineClient, err := rpc.DialContext(s.ctx, cfg.Notification.EngineRPCServer.Endpoint)
	if err != nil {
		return err
	}

	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.loadEndpoint()
	s.migrateDB()
	s.loadScyllaDB()
	s.loadPublisher()
	s.loadRedisClient()
	s.loadStorage()
	s.loadRepos(client.NewSearchCaller(rpcSearchClient))
	s.loadLeaderboard()
	s.loadBadgeManager()
	s.loadDomains(
		client.NewGameCenterCaller(rpcGameCenterClient),
		client.NewNotificationEngineCaller(rpcNotificationEngineClient),
	)
	router := s.loadAPIRouter()

	httpSrv := &http.Server{
		Addr:    cfg.ApiServer.Address(),
		Handler: router.Handler(cfg.ApiServer.ServerConfigs),
	}
	xcontext.Logger(s.ctx).Infof("Starting server on port: %s", cfg.ApiServer.Port)
	if err := httpSrv.ListenAndServe(); err != nil {
		return err
	}
	xcontext.Logger(s.ctx).Infof("Server stop")
	return nil
}

func (s *srv) loadAPIRouter() *router.Router {
	cfg := xcontext.Configs(s.ctx)
	defaultRouter := router.New(s.ctx)
	defaultRouter.AddCloser(middleware.Logger(cfg.Env))
	defaultRouter.After(middleware.HandleSaveSession())

	// Auth API
	{
		router.GET(defaultRouter, "/loginWallet", s.authDomain.WalletLogin)
		router.POST(defaultRouter, "/verifyWallet", s.authDomain.WalletVerify)
		router.POST(defaultRouter, "/verifyOAuth2", s.authDomain.OAuth2Verify)
		router.POST(defaultRouter, "/refresh", s.authDomain.Refresh)
	}

	// These following APIs need authentication with only Access Token.
	onlyTokenAuthRouter := defaultRouter.Branch()
	authVerifier := middleware.NewAuthVerifier().WithAccessToken()
	onlyTokenAuthRouter.Before(authVerifier.Middleware())
	{
		// Link account API
		router.POST(onlyTokenAuthRouter, "/linkOAuth2", s.authDomain.OAuth2Link)
		router.POST(onlyTokenAuthRouter, "/linkWallet", s.authDomain.WalletLink)
		router.POST(onlyTokenAuthRouter, "/linkTelegram", s.authDomain.TelegramLink)

		// User API
		router.GET(onlyTokenAuthRouter, "/getMe", s.userDomain.GetMe)
		router.GET(onlyTokenAuthRouter, "/getUser", s.userDomain.GetUser)
		router.GET(onlyTokenAuthRouter, "/getMyBadgeDetails", s.badgeDomain.GetMyBadgeDetails)
		router.GET(onlyTokenAuthRouter, "/getUserBadgeDetails", s.badgeDomain.GetUserBadgeDetails)
		router.POST(onlyTokenAuthRouter, "/follow", s.userDomain.FollowCommunity)
		router.POST(onlyTokenAuthRouter, "/uploadAvatar", s.userDomain.UploadAvatar)
		router.POST(onlyTokenAuthRouter, "/updateUser", s.userDomain.Update)

		// Community API
		router.GET(onlyTokenAuthRouter, "/getMyReferrals", s.communityDomain.GetMyReferral)
		router.GET(onlyTokenAuthRouter, "/getDiscordRoles", s.communityDomain.GetDiscordRole)
		router.GET(onlyTokenAuthRouter, "/getMyOwnCommunities", s.communityDomain.GetMyOwn)
		router.POST(onlyTokenAuthRouter, "/createCommunity", s.communityDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateCommunity", s.communityDomain.UpdateByID)
		router.POST(onlyTokenAuthRouter, "/deleteCommunity", s.communityDomain.DeleteByID)
		router.POST(onlyTokenAuthRouter, "/updateCommunityDiscord", s.communityDomain.UpdateDiscord)
		router.POST(onlyTokenAuthRouter, "/uploadCommunityLogo", s.communityDomain.UploadLogo)

		// Follower API
		router.GET(onlyTokenAuthRouter, "/getMyFollowerInfo", s.followerDomain.Get)
		router.GET(onlyTokenAuthRouter, "/getMyFollowing", s.followerDomain.GetByUserID)
		router.GET(onlyTokenAuthRouter, "/getCommunityFollowers", s.followerDomain.GetByCommunityID)

		// API-Key API
		router.POST(onlyTokenAuthRouter, "/generateAPIKey", s.apiKeyDomain.Generate)
		router.POST(onlyTokenAuthRouter, "/regenerateAPIKey", s.apiKeyDomain.Regenerate)
		router.POST(onlyTokenAuthRouter, "/revokeAPIKey", s.apiKeyDomain.Revoke)

		// Quest API
		router.POST(onlyTokenAuthRouter, "/createQuest", s.questDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateQuest", s.questDomain.Update)
		router.POST(onlyTokenAuthRouter, "/updateQuestCategory", s.questDomain.UpdateCategory)
		router.POST(onlyTokenAuthRouter, "/updateQuestPosition", s.questDomain.UpdatePosition)
		router.POST(onlyTokenAuthRouter, "/deleteQuest", s.questDomain.Delete)
		router.POST(onlyTokenAuthRouter, "/parseTemplate", s.questDomain.ParseTemplate)

		// Category API
		router.POST(onlyTokenAuthRouter, "/createCategory", s.categoryDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateCategory", s.categoryDomain.UpdateByID)
		router.POST(onlyTokenAuthRouter, "/deleteCategory", s.categoryDomain.DeleteByID)

		// Claimed Quest API
		router.POST(onlyTokenAuthRouter, "/claim", s.claimedQuestDomain.Claim)
		router.POST(onlyTokenAuthRouter, "/claimReferral", s.claimedQuestDomain.ClaimReferral)

		// Transaction API
		router.GET(onlyTokenAuthRouter, "/getMyPayRewards", s.payRewardDomain.GetMyPayRewards)

		// Image API
		router.POST(onlyTokenAuthRouter, "/uploadImage", s.fileDomain.UploadImage)

		// Game API
		router.GET(onlyTokenAuthRouter, "/getRoomsByCommunity", s.gameDomain.GetRoomsByCommunity)
		router.GET(onlyTokenAuthRouter, "/getCharacters", s.gameDomain.GetAllCharacters)
		router.GET(onlyTokenAuthRouter, "/getCommunityCharacters", s.gameDomain.GetAllCommunityCharacters)
		router.GET(onlyTokenAuthRouter, "/getMyCharacters", s.gameDomain.GetMyCharacters)
		router.POST(onlyTokenAuthRouter, "/createLuckyboxEvent", s.gameDomain.CreateLuckyboxEvent)
		router.POST(onlyTokenAuthRouter, "/setupCommunityCharacter", s.gameDomain.SetupCommunityCharacter)
		router.POST(onlyTokenAuthRouter, "/buyCharacter", s.gameDomain.BuyCharacter)

		// Chat API
		router.GET(onlyTokenAuthRouter, "/getChannels", s.chatDomain.GetChannles)
		router.POST(onlyTokenAuthRouter, "/createChannel", s.chatDomain.CreateChannel)
		router.POST(onlyTokenAuthRouter, "/createMessage", s.chatDomain.CreateMessage)
		router.POST(onlyTokenAuthRouter, "/addReaction", s.chatDomain.AddReaction)
		router.POST(onlyTokenAuthRouter, "/deleteMessage", s.chatDomain.DeleteMessage)

		// Role API
		router.POST(onlyTokenAuthRouter, "/createRole", s.roleDomain.CreateRole)
		router.POST(onlyTokenAuthRouter, "/updateRole", s.roleDomain.UpdateRole)
		router.POST(onlyTokenAuthRouter, "/deleteRole", s.roleDomain.DeleteRole)

	}

	onlyAdminVerifier := middleware.NewOnlyAdmin(s.userRepo)
	onlyAdminRouter := onlyTokenAuthRouter.Branch()
	onlyAdminRouter.Before(onlyAdminVerifier.Middleware())
	{
		// User API
		router.POST(onlyAdminRouter, "/assignGlobalRole", s.userDomain.Assign)

		// Badge API
		router.POST(onlyAdminRouter, "/updateBadge", s.badgeDomain.UpdateBadge)

		// Community API
		router.GET(onlyAdminRouter, "/getReferrals", s.communityDomain.GetReferral)
		router.GET(onlyAdminRouter, "/getPendingCommunities", s.communityDomain.GetListPending)
		router.POST(onlyAdminRouter, "/approvePendingCommunity", s.communityDomain.ApprovePending)
		router.POST(onlyAdminRouter, "/reviewReferral", s.communityDomain.ReviewReferral)
		router.POST(onlyAdminRouter, "/transferCommunity", s.communityDomain.TransferCommunity)
		router.POST(onlyAdminRouter, "/assignCommunityRole", s.communityDomain.AssignRole)

		// Game API
		router.GET(onlyAdminRouter, "/getMaps", s.gameDomain.GetMaps)
		router.POST(onlyAdminRouter, "/createMap", s.gameDomain.CreateMap)
		router.POST(onlyAdminRouter, "/createRoom", s.gameDomain.CreateRoom)
		router.POST(onlyAdminRouter, "/deleteMap", s.gameDomain.DeleteMap)
		router.POST(onlyAdminRouter, "/deleteRoom", s.gameDomain.DeleteRoom)
		router.POST(onlyAdminRouter, "/createCharacter", s.gameDomain.CreateCharacter)
	}

	// These following APIs support authentication with both Access Token and API Key.
	tokenAndKeyAuthRouter := defaultRouter.Branch()
	authVerifier = middleware.NewAuthVerifier().WithAccessToken().WithAPIKey(s.apiKeyRepo)
	tokenAndKeyAuthRouter.Before(authVerifier.Middleware())
	{
		router.GET(tokenAndKeyAuthRouter, "/getClaimedQuest", s.claimedQuestDomain.Get)
		router.GET(tokenAndKeyAuthRouter, "/getClaimedQuests", s.claimedQuestDomain.GetList)
		router.POST(tokenAndKeyAuthRouter, "/review", s.claimedQuestDomain.Review)
		router.POST(tokenAndKeyAuthRouter, "/reviewAll", s.claimedQuestDomain.ReviewAll)
		router.POST(tokenAndKeyAuthRouter, "/givePoint", s.claimedQuestDomain.GivePoint)
	}

	// Public API
	publicRouter := defaultRouter.Branch()
	optionalAuthVerifier := middleware.NewAuthVerifier().WithAccessToken().WithOptional()
	publicRouter.Before(optionalAuthVerifier.Middleware())
	{
		router.GET(publicRouter, "/", homeHandle)
		router.GET(publicRouter, "/getQuest", s.questDomain.Get)
		router.GET(publicRouter, "/getQuests", s.questDomain.GetList)
		router.GET(publicRouter, "/getTemplates", s.questDomain.GetTemplates)
		router.GET(publicRouter, "/getTemplateCategories", s.categoryDomain.GetTemplate)
		router.GET(publicRouter, "/getCommunities", s.communityDomain.GetList)
		router.GET(publicRouter, "/getCommunity", s.communityDomain.Get)
		router.GET(publicRouter, "/getInvite", s.userDomain.GetInvite)
		router.GET(publicRouter, "/getLeaderBoard", s.statisticDomain.GetLeaderBoard)
		router.GET(publicRouter, "/getAllBadgeNames", s.badgeDomain.GetAllBadgeNames)
		router.GET(publicRouter, "/getAllBadges", s.badgeDomain.GetAllBadges)
		router.GET(publicRouter, "/getMessages", s.chatDomain.GetMessages)
		router.GET(publicRouter, "/getCategories", s.categoryDomain.GetList)
		router.GET(publicRouter, "/getUserReactions", s.chatDomain.GetUserReactions)
		router.GET(publicRouter, "/getRoles", s.roleDomain.GetRoles)
	}

	return defaultRouter
}

type homeRequest struct{}
type homeResponse struct{}

func homeHandle(ctx context.Context, req *homeRequest) (*homeResponse, error) {
	return &homeResponse{}, nil
}
