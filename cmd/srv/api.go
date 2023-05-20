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
	server.loadEndpoint()
	server.loadStorage()
	server.loadRepos()
	server.loadBadgeManager()
	server.loadDomains()
	server.loadRouter()
	server.setupTrendingPoints()

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

func (s *srv) setupTrendingPoints() {
	s.projectDomain.RunPeriodicResetTrendingPoints(s.ctx)
}

func (s *srv) loadRouter() {
	s.router = router.New(s.ctx)
	s.router.Static("/", "./web")
	s.router.AddCloser(middleware.Logger())
	s.router.After(middleware.HandleSaveSession())

	// Auth API
	{
		router.GET(s.router, "/oauth2/verify", s.authDomain.OAuth2Verify)
		router.GET(s.router, "/wallet/login", s.authDomain.WalletLogin)
		router.GET(s.router, "/wallet/verify", s.authDomain.WalletVerify)
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
		router.GET(onlyTokenAuthRouter, "/getUser", s.userDomain.GetUser)
		router.GET(onlyTokenAuthRouter, "/getMyBadges", s.userDomain.GetMyBadges)
		router.POST(onlyTokenAuthRouter, "/follow", s.userDomain.FollowProject)
		router.POST(onlyTokenAuthRouter, "/assignGlobalRole", s.userDomain.Assign)
		router.POST(onlyTokenAuthRouter, "/uploadAvatar", s.userDomain.UploadAvatar)
		router.POST(onlyTokenAuthRouter, updateUserPattern, s.userDomain.Update)

		// Project API
		router.GET(onlyTokenAuthRouter, "/getFollowingProjects", s.projectDomain.GetFollowing)
		router.GET(onlyTokenAuthRouter, "/getMyReferralInfo", s.projectDomain.GetMyReferral)
		router.GET(onlyTokenAuthRouter, "/getPendingReferralProjects", s.projectDomain.GetPendingReferral)
		router.POST(onlyTokenAuthRouter, "/createProject", s.projectDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateProjectByID", s.projectDomain.UpdateByID)
		router.POST(onlyTokenAuthRouter, "/deleteProjectByID", s.projectDomain.DeleteByID)
		router.POST(onlyTokenAuthRouter, "/updateProjectDiscord", s.projectDomain.UpdateDiscord)
		router.POST(onlyTokenAuthRouter, "/uploadProjectLogo", s.projectDomain.UploadLogo)
		router.POST(onlyTokenAuthRouter, "/approveReferralProjects", s.projectDomain.ApproveReferral)

		// Participant API
		router.GET(onlyTokenAuthRouter, "/getParticipant", s.participantDomain.Get)
		router.GET(onlyTokenAuthRouter, "/getListParticipant", s.participantDomain.GetList)

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
		router.GET(onlyTokenAuthRouter, "/getListCategory", s.categoryDomain.GetList)
		router.POST(onlyTokenAuthRouter, "/createCategory", s.categoryDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateCategoryByID", s.categoryDomain.UpdateByID)
		router.POST(onlyTokenAuthRouter, "/deleteCategoryByID", s.categoryDomain.DeleteByID)

		// Collaborator API
		router.GET(onlyTokenAuthRouter, "/getMyCollaborators", s.collaboratorDomain.GetMyCollabs)
		router.GET(onlyTokenAuthRouter, "/getProjectCollaborators", s.collaboratorDomain.GetProjectCollabs)
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
		router.POST(onlyTokenAuthRouter, "/createMap", s.gameDomain.CreateMap)
		router.POST(onlyTokenAuthRouter, "/createRoom", s.gameDomain.CreateRoom)
		router.POST(onlyTokenAuthRouter, "/deleteMap", s.gameDomain.DeleteMap)
		router.POST(onlyTokenAuthRouter, "/deleteRoom", s.gameDomain.DeleteRoom)
		router.GET(onlyTokenAuthRouter, "/getMap", s.gameDomain.GetMapInfo)
	}

	// These following APIs support authentication with both Access Token and API Key.
	tokenAndKeyAuthRouter := s.router.Branch()
	authVerifier = middleware.NewAuthVerifier().WithAccessToken().WithAPIKey(s.apiKeyRepo)
	tokenAndKeyAuthRouter.Before(authVerifier.Middleware())
	{
		router.GET(tokenAndKeyAuthRouter, "/getClaimedQuest", s.claimedQuestDomain.Get)
		router.GET(tokenAndKeyAuthRouter, "/getListClaimedQuest", s.claimedQuestDomain.GetList)
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
		router.GET(publicRouter, "/getListQuest", s.questDomain.GetList)
		router.GET(publicRouter, "/getTemplates", s.questDomain.GetTemplates)
		router.GET(publicRouter, "/getListProject", s.projectDomain.GetList)
		router.GET(publicRouter, "/getProjectByID", s.projectDomain.GetByID)
		router.GET(publicRouter, "/getInvite", s.userDomain.GetInvite)
		router.GET(publicRouter, "/getLeaderBoard", s.statisticDomain.GetLeaderBoard)
		router.GET(publicRouter, "/getBadges", s.userDomain.GetBadges)
	}
}
