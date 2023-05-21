package main

import (
	"fmt"
	"net/http"

	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (app *App) startApi(*cli.Context) error {
	app.loadEndpoint()
	app.loadStorage()
	app.loadRepos()
	app.loadBadgeManager()
	app.loadDomains()
	app.loadRouter()

	cfg := xcontext.Configs(app.ctx)
	app.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ApiServer.Port),
		Handler: app.router.Handler(cfg.ApiServer.ServerConfigs),
	}

	xcontext.Logger(app.ctx).Infof("Starting server on port: %s", cfg.ApiServer.Port)
	if err := app.server.ListenAndServe(); err != nil {
		panic(err)
	}
	xcontext.Logger(app.ctx).Infof("Server stop")
	return nil
}

const updateUserPattern = "/updateUser"

func (app *App) loadRouter() {
	app.router = router.New(app.ctx)
	app.router.Static("/", "./web")
	app.router.AddCloser(middleware.Logger())
	app.router.After(middleware.HandleSaveSession())

	// Auth API
	{
		router.GET(app.router, "/oauth2/verify", app.authDomain.OAuth2Verify)
		router.GET(app.router, "/wallet/login", app.authDomain.WalletLogin)
		router.GET(app.router, "/wallet/verify", app.authDomain.WalletVerify)
		router.POST(app.router, "/refresh", app.authDomain.Refresh)
	}

	// These following APIs need authentication with only Access Token.
	onlyTokenAuthRouter := app.router.Branch()
	authVerifier := middleware.NewAuthVerifier().WithAccessToken()
	onlyTokenAuthRouter.Before(authVerifier.Middleware())
	//TODO: onlyTokenAuthRouter.Before(middleware.MustUpdateUsername(s.userRepo, updateUserPattern))
	{
		// Link account API
		router.POST(onlyTokenAuthRouter, "/linkOAuth2", app.authDomain.OAuth2Link)
		router.POST(onlyTokenAuthRouter, "/linkWallet", app.authDomain.WalletLink)
		router.POST(onlyTokenAuthRouter, "/linkTelegram", app.authDomain.TelegramLink)

		// User API
		router.GET(onlyTokenAuthRouter, "/getUser", app.userDomain.GetUser)
		router.GET(onlyTokenAuthRouter, "/getMyBadges", app.userDomain.GetMyBadges)
		router.POST(onlyTokenAuthRouter, "/follow", app.userDomain.FollowProject)
		router.POST(onlyTokenAuthRouter, "/assignGlobalRole", app.userDomain.Assign)
		router.POST(onlyTokenAuthRouter, "/uploadAvatar", app.userDomain.UploadAvatar)
		router.POST(onlyTokenAuthRouter, updateUserPattern, app.userDomain.Update)

		// Project API
		router.GET(onlyTokenAuthRouter, "/getFollowingProjects", app.projectDomain.GetFollowing)
		router.GET(onlyTokenAuthRouter, "/getMyReferralInfo", app.projectDomain.GetMyReferral)
		router.GET(onlyTokenAuthRouter, "/getPendingReferralProjects", app.projectDomain.GetPendingReferral)
		router.POST(onlyTokenAuthRouter, "/createProject", app.projectDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateProjectByID", app.projectDomain.UpdateByID)
		router.POST(onlyTokenAuthRouter, "/deleteProjectByID", app.projectDomain.DeleteByID)
		router.POST(onlyTokenAuthRouter, "/updateProjectDiscord", app.projectDomain.UpdateDiscord)
		router.POST(onlyTokenAuthRouter, "/uploadProjectLogo", app.projectDomain.UploadLogo)
		router.POST(onlyTokenAuthRouter, "/approveReferralProjects", app.projectDomain.ApproveReferral)

		// Participant API
		router.GET(onlyTokenAuthRouter, "/getParticipant", app.participantDomain.Get)
		router.GET(onlyTokenAuthRouter, "/getListParticipant", app.participantDomain.GetList)

		// API-Key API
		router.POST(onlyTokenAuthRouter, "/generateAPIKey", app.apiKeyDomain.Generate)
		router.POST(onlyTokenAuthRouter, "/regenerateAPIKey", app.apiKeyDomain.Regenerate)
		router.POST(onlyTokenAuthRouter, "/revokeAPIKey", app.apiKeyDomain.Revoke)

		// Quest API
		router.POST(onlyTokenAuthRouter, "/createQuest", app.questDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateQuest", app.questDomain.Update)
		router.POST(onlyTokenAuthRouter, "/deleteQuest", app.questDomain.Delete)
		router.POST(onlyTokenAuthRouter, "/parseTemplate", app.questDomain.ParseTemplate)

		// Category API
		router.GET(onlyTokenAuthRouter, "/getListCategory", app.categoryDomain.GetList)
		router.POST(onlyTokenAuthRouter, "/createCategory", app.categoryDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateCategoryByID", app.categoryDomain.UpdateByID)
		router.POST(onlyTokenAuthRouter, "/deleteCategoryByID", app.categoryDomain.DeleteByID)

		// Collaborator API
		router.GET(onlyTokenAuthRouter, "/getMyCollaborators", app.collaboratorDomain.GetMyCollabs)
		router.GET(onlyTokenAuthRouter, "/getProjectCollaborators", app.collaboratorDomain.GetProjectCollabs)
		router.POST(onlyTokenAuthRouter, "/assignCollaborator", app.collaboratorDomain.Assign)
		router.POST(onlyTokenAuthRouter, "/deleteCollaborator", app.collaboratorDomain.Delete)

		// Claimed Quest API
		router.POST(onlyTokenAuthRouter, "/claim", app.claimedQuestDomain.Claim)
		router.POST(onlyTokenAuthRouter, "/claimReferral", app.claimedQuestDomain.ClaimReferral)

		// Transaction API
		router.GET(onlyTokenAuthRouter, "/getMyTransactions", app.transactionDomain.GetMyTransactions)

		// Image API
		router.POST(onlyTokenAuthRouter, "/uploadImage", app.fileDomain.UploadImage)

		// Game API
		router.POST(onlyTokenAuthRouter, "/createMap", app.gameDomain.CreateMap)
		router.POST(onlyTokenAuthRouter, "/createRoom", app.gameDomain.CreateRoom)
		router.POST(onlyTokenAuthRouter, "/deleteMap", app.gameDomain.DeleteMap)
		router.POST(onlyTokenAuthRouter, "/deleteRoom", app.gameDomain.DeleteRoom)
		router.GET(onlyTokenAuthRouter, "/getMap", app.gameDomain.GetMapInfo)
	}

	// These following APIs support authentication with both Access Token and API Key.
	tokenAndKeyAuthRouter := app.router.Branch()
	authVerifier = middleware.NewAuthVerifier().WithAccessToken().WithAPIKey(app.apiKeyRepo)
	tokenAndKeyAuthRouter.Before(authVerifier.Middleware())
	{
		router.GET(tokenAndKeyAuthRouter, "/getClaimedQuest", app.claimedQuestDomain.Get)
		router.GET(tokenAndKeyAuthRouter, "/getListClaimedQuest", app.claimedQuestDomain.GetList)
		router.POST(tokenAndKeyAuthRouter, "/review", app.claimedQuestDomain.Review)
		router.POST(tokenAndKeyAuthRouter, "/reviewAll", app.claimedQuestDomain.ReviewAll)
		router.POST(tokenAndKeyAuthRouter, "/giveReward", app.claimedQuestDomain.GiveReward)
	}

	// Public API.
	publicRouter := app.router.Branch()
	optionalAuthVerifier := middleware.NewAuthVerifier().WithAccessToken().WithOptional()
	publicRouter.Before(optionalAuthVerifier.Middleware())
	{
		router.GET(publicRouter, "/getQuest", app.questDomain.Get)
		router.GET(publicRouter, "/getListQuest", app.questDomain.GetList)
		router.GET(publicRouter, "/getTemplates", app.questDomain.GetTemplates)
		router.GET(publicRouter, "/getListProject", app.projectDomain.GetList)
		router.GET(publicRouter, "/getProjectByID", app.projectDomain.GetByID)
		router.GET(publicRouter, "/getInvite", app.userDomain.GetInvite)
		router.GET(publicRouter, "/getLeaderBoard", app.statisticDomain.GetLeaderBoard)
		router.GET(publicRouter, "/getBadges", app.userDomain.GetBadges)
	}
}
