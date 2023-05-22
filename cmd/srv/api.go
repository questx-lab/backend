package main

import (
	"fmt"
	"net/http"

	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startApi(*cli.Context) error {
	s.loadEndpoint()
	s.loadStorage()
	s.loadRepos()
	s.loadBadgeManager()
	s.loadDomains()
	s.loadRouter()

	cfg := xcontext.Configs(s.ctx)
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ApiServer.Port),
		Handler: s.router.Handler(cfg.ApiServer.ServerConfigs),
	}

	xcontext.Logger(s.ctx).Infof("Starting server on port: %s", cfg.ApiServer.Port)
	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
	xcontext.Logger(s.ctx).Infof("Server stop")
	return nil
}

const updateUserPattern = "/updateUser"

func (s *srv) loadRouter() {
	s.router = router.New(s.ctx)
	s.router.Static("/", "./web")
	s.router.AddCloser(middleware.Logger())
	s.router.After(middleware.HandleSaveSession())

	// Auth API
	{
		router.GET(s.router, "/loginWallet", s.authDomain.WalletLogin)
		router.POST(s.router, "/verifyWallet", s.authDomain.WalletVerify)
		router.POST(s.router, "/verifyOAuth2", s.authDomain.OAuth2Verify)
		router.POST(s.router, "/refresh", s.authDomain.Refresh)
	}

	// These following APIs need authentication with only Access Token.
	onlyTokenAuthRouter := s.router.Branch()
	authVerifier := middleware.NewAuthVerifier().WithAccessToken()
	onlyTokenAuthRouter.Before(authVerifier.Middleware())
	//TODO: onlyTokenAuthRouter.Before(middleware.MustUpdateUsername(s.userRepo, updateUserPattern))
	{
		// Link account API
		router.POST(onlyTokenAuthRouter, "/linkOAuth2", s.authDomain.OAuth2Link)
		router.POST(onlyTokenAuthRouter, "/linkWallet", s.authDomain.WalletLink)
		router.POST(onlyTokenAuthRouter, "/linkTelegram", s.authDomain.TelegramLink)

		// User API
		router.GET(onlyTokenAuthRouter, "/getMe", s.userDomain.GetUser)
		router.GET(onlyTokenAuthRouter, "/getMyBadges", s.userDomain.GetMyBadges)
		router.POST(onlyTokenAuthRouter, "/follow", s.userDomain.FollowCommunity)
		router.POST(onlyTokenAuthRouter, "/assignGlobalRole", s.userDomain.Assign)
		router.POST(onlyTokenAuthRouter, "/uploadAvatar", s.userDomain.UploadAvatar)
		router.POST(onlyTokenAuthRouter, updateUserPattern, s.userDomain.Update)

		// Community API
		router.GET(onlyTokenAuthRouter, "/getFollowingCommunities", s.communityDomain.GetFollowing)
		router.GET(onlyTokenAuthRouter, "/getMyReferrals", s.communityDomain.GetMyReferral)
		router.GET(onlyTokenAuthRouter, "/getPendingReferrals", s.communityDomain.GetPendingReferral)
		router.POST(onlyTokenAuthRouter, "/createCommunity", s.communityDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateCommunity", s.communityDomain.UpdateByID)
		router.POST(onlyTokenAuthRouter, "/deleteCommunity", s.communityDomain.DeleteByID)
		router.POST(onlyTokenAuthRouter, "/updateCommunityDiscord", s.communityDomain.UpdateDiscord)
		router.POST(onlyTokenAuthRouter, "/uploadCommunityLogo", s.communityDomain.UploadLogo)
		router.POST(onlyTokenAuthRouter, "/approveReferrals", s.communityDomain.ApproveReferral)

		// Follower API
		router.GET(onlyTokenAuthRouter, "/getMyFollowerInfo", s.followerDomain.Get)
		router.GET(onlyTokenAuthRouter, "/getFollowers", s.followerDomain.GetList)

		// API-Key API
		router.POST(onlyTokenAuthRouter, "/generateAPIKey", s.apiKeyDomain.Generate)
		router.POST(onlyTokenAuthRouter, "/regenerateAPIKey", s.apiKeyDomain.Regenerate)
		router.POST(onlyTokenAuthRouter, "/revokeAPIKey", s.apiKeyDomain.Revoke)

		// Quest API
		router.POST(onlyTokenAuthRouter, "/createQuest", s.questDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateQuest", s.questDomain.Update)
		router.POST(onlyTokenAuthRouter, "/deleteQuest", s.questDomain.Delete)
		router.POST(onlyTokenAuthRouter, "/parseTemplate", s.questDomain.ParseTemplate)

		// Category API
		router.GET(onlyTokenAuthRouter, "/getCategories", s.categoryDomain.GetList)
		router.POST(onlyTokenAuthRouter, "/createCategory", s.categoryDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateCategory", s.categoryDomain.UpdateByID)
		router.POST(onlyTokenAuthRouter, "/deleteCategory", s.categoryDomain.DeleteByID)

		// Collaborator API
		router.GET(onlyTokenAuthRouter, "/getMyCollaborators", s.collaboratorDomain.GetMyCollabs)
		router.GET(onlyTokenAuthRouter, "/getCommunityCollaborators", s.collaboratorDomain.GetCommunityCollabs)
		router.POST(onlyTokenAuthRouter, "/assignCollaborator", s.collaboratorDomain.Assign)
		router.POST(onlyTokenAuthRouter, "/deleteCollaborator", s.collaboratorDomain.Delete)

		// Claimed Quest API
		router.POST(onlyTokenAuthRouter, "/claim", s.claimedQuestDomain.Claim)
		router.POST(onlyTokenAuthRouter, "/claimReferral", s.claimedQuestDomain.ClaimReferral)

		// Transaction API
		router.GET(onlyTokenAuthRouter, "/getMyTransactions", s.transactionDomain.GetMyTransactions)

		// Image API
		router.POST(onlyTokenAuthRouter, "/uploadImage", s.fileDomain.UploadImage)

		// Game API
		router.GET(onlyTokenAuthRouter, "/getMap", s.gameDomain.GetMapInfo)
		router.POST(onlyTokenAuthRouter, "/createMap", s.gameDomain.CreateMap)
		router.POST(onlyTokenAuthRouter, "/createRoom", s.gameDomain.CreateRoom)
		router.POST(onlyTokenAuthRouter, "/deleteMap", s.gameDomain.DeleteMap)
		router.POST(onlyTokenAuthRouter, "/deleteRoom", s.gameDomain.DeleteRoom)
	}

	// These following APIs support authentication with both Access Token and API Key.
	tokenAndKeyAuthRouter := s.router.Branch()
	authVerifier = middleware.NewAuthVerifier().WithAccessToken().WithAPIKey(s.apiKeyRepo)
	tokenAndKeyAuthRouter.Before(authVerifier.Middleware())
	{
		router.GET(tokenAndKeyAuthRouter, "/getClaimedQuest", s.claimedQuestDomain.Get)
		router.GET(tokenAndKeyAuthRouter, "/getClaimedQuests", s.claimedQuestDomain.GetList)
		router.POST(tokenAndKeyAuthRouter, "/review", s.claimedQuestDomain.Review)
		router.POST(tokenAndKeyAuthRouter, "/reviewAll", s.claimedQuestDomain.ReviewAll)
		router.POST(tokenAndKeyAuthRouter, "/giveReward", s.claimedQuestDomain.GiveReward)
	}

	// Public API.
	publicRouter := s.router.Branch()
	optionalAuthVerifier := middleware.NewAuthVerifier().WithAccessToken().WithOptional()
	publicRouter.Before(optionalAuthVerifier.Middleware())
	{
		router.GET(publicRouter, "/getQuest", s.questDomain.Get)
		router.GET(publicRouter, "/getQuests", s.questDomain.GetList)
		router.GET(publicRouter, "/getTemplates", s.questDomain.GetTemplates)
		router.GET(publicRouter, "/getCommunities", s.communityDomain.GetList)
		router.GET(publicRouter, "/getCommunity", s.communityDomain.Get)
		router.GET(publicRouter, "/getInvite", s.userDomain.GetInvite)
		router.GET(publicRouter, "/getLeaderBoard", s.statisticDomain.GetLeaderBoard)
		router.GET(publicRouter, "/getBadges", s.userDomain.GetBadges)
	}
}
