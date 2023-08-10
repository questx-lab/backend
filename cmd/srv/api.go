package main

import (
	"context"
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/prometheus"
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

	rpcBlockchainClient, err := rpc.DialContext(s.ctx, cfg.Blockchain.Endpoint)
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
		client.NewBlockchainCaller(rpcBlockchainClient),
		client.NewNotificationEngineCaller(rpcNotificationEngineClient),
	)

	go func() {
		promHandler := prometheus.NewHandler()

		httpSrv := &http.Server{
			Addr:    cfg.PrometheusServer.Address(),
			Handler: promHandler,
		}
		xcontext.Logger(s.ctx).Infof("Starting prometheus on port: %s", cfg.PrometheusServer.Port)
		if err := httpSrv.ListenAndServe(); err != nil {
			panic(err)
		}
		xcontext.Logger(s.ctx).Infof("Server prometheus stop")
	}()

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
	defaultRouter.Before(middleware.WithStartTime())
	defaultRouter.AddCloser(middleware.Logger(cfg.Env))
	defaultRouter.AddCloser(middleware.Prometheus())
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
		router.POST(onlyTokenAuthRouter, "/unfollow", s.userDomain.UnFollowCommunity)
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
		router.GET(onlyTokenAuthRouter, "/searchMention", s.followerDomain.SearchMention)
		router.GET(onlyTokenAuthRouter, "/getStreaks", s.followerDomain.GetStreaks)

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

		// Image API
		router.POST(onlyTokenAuthRouter, "/uploadImage", s.fileDomain.UploadImage)

		// Blockchain API
		router.GET(onlyTokenAuthRouter, "/getWalletAddress", s.blockchainDomain.GetWalletAddress)
		router.GET(onlyTokenAuthRouter, "/getMyPayRewards", s.payRewardDomain.GetMyPayRewards)
		router.GET(onlyTokenAuthRouter, "/getClaimableRewards", s.payRewardDomain.GetClaimableRewards)

		// Chat API
		router.GET(onlyTokenAuthRouter, "/getChannels", s.chatDomain.GetChannels)
		router.GET(onlyTokenAuthRouter, "/deleteChannel", s.chatDomain.DeleteChannel)
		router.POST(onlyTokenAuthRouter, "/createChannel", s.chatDomain.CreateChannel)
		router.POST(onlyTokenAuthRouter, "/updateChannel", s.chatDomain.UpdateChannel)
		router.POST(onlyTokenAuthRouter, "/createMessage", s.chatDomain.CreateMessage)
		router.POST(onlyTokenAuthRouter, "/addReaction", s.chatDomain.AddReaction)
		router.POST(onlyTokenAuthRouter, "/removeReaction", s.chatDomain.RemoveReaction)
		router.POST(onlyTokenAuthRouter, "/deleteMessage", s.chatDomain.DeleteMessage)

		router.POST(onlyTokenAuthRouter, "/assignCommunityRole", s.communityDomain.AssignRole)
		router.POST(onlyTokenAuthRouter, "/deleteUserCommunityRole", s.communityDomain.DeleteUserCommunityRole)

		// Role API
		router.POST(onlyTokenAuthRouter, "/createRole", s.roleDomain.CreateRole)
		router.POST(onlyTokenAuthRouter, "/updateRole", s.roleDomain.UpdateRole)
		router.POST(onlyTokenAuthRouter, "/deleteRole", s.roleDomain.DeleteRole)

		// Lottery API
		router.GET(onlyTokenAuthRouter, "/getLotteryEvent", s.lotteryDomain.GetLotteryEvent)
		router.POST(onlyTokenAuthRouter, "/createLotteryEvent", s.lotteryDomain.CreateLotteryEvent)
		router.POST(onlyTokenAuthRouter, "/buyLotteryTickets", s.lotteryDomain.BuyTicket)
		router.POST(onlyTokenAuthRouter, "/claimLotteryWinner", s.lotteryDomain.Claim)
	}

	onlyAdminVerifier := middleware.NewOnlyAdmin(s.userRepo)
	onlyAdminRouter := onlyTokenAuthRouter.Branch()
	onlyAdminRouter.Before(onlyAdminVerifier.Middleware())
	{
		// User API
		router.GET(onlyAdminRouter, "/getTotalUsers", s.userDomain.CountTotalUsers)
		router.POST(onlyAdminRouter, "/assignGlobalRole", s.userDomain.Assign)

		// Badge API
		router.POST(onlyAdminRouter, "/updateBadge", s.badgeDomain.UpdateBadge)

		// Community API
		router.GET(onlyAdminRouter, "/getReferrals", s.communityDomain.GetReferral)
		router.GET(onlyAdminRouter, "/getPendingCommunities", s.communityDomain.GetListPending)
		router.GET(onlyAdminRouter, "/getCommunityRecords", s.communityDomain.GetRecords)
		router.POST(onlyAdminRouter, "/reviewPendingCommunity", s.communityDomain.ReviewPending)
		router.POST(onlyAdminRouter, "/reviewReferral", s.communityDomain.ReviewReferral)
		router.POST(onlyAdminRouter, "/transferCommunity", s.communityDomain.TransferCommunity)

		// Blockchain API
		router.GET(onlyAdminRouter, "/getBlockchain", s.blockchainDomain.GetChain)
		router.POST(onlyAdminRouter, "/createBlockchain", s.blockchainDomain.CreateChain)
		router.POST(onlyAdminRouter, "/createBlockchainConnection", s.blockchainDomain.CreateConnection)
		router.POST(onlyAdminRouter, "/deleteBlockchainConnection", s.blockchainDomain.DeleteConnection)
		router.POST(onlyAdminRouter, "/createBlockchainToken", s.blockchainDomain.CreateToken)
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
