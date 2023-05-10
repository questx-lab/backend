package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"

	"github.com/urfave/cli/v2"
)

func (s *srv) startApi(ct *cli.Context) error {
	server.loadConfig()
	server.loadLogger()
	server.loadEndpoint()
	server.loadDatabase()
	server.loadStorage()
	server.loadRepos()
	server.loadBadgeManager()
	server.loadDomains()
	server.loadRouter()

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.ApiServer.Port),
		Handler: s.router.Handler(),
	}

	log.Printf("Starting server on port: %s\n", s.configs.ApiServer.Port)
	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
	log.Printf("server stop")
	return nil
}

func (s *srv) loadRouter() {
	s.router = router.New(s.db, *s.configs, s.logger)
	s.router.Static("/", "./web")
	s.router.AddCloser(middleware.Logger())

	// Auth API
	authRouter := s.router.Branch()
	authRouter.After(middleware.HandleSaveSession())
	{
		router.GET(authRouter, "/oauth2/verify", s.authDomain.OAuth2Verify)
		router.GET(authRouter, "/wallet/login", s.authDomain.WalletLogin)
		router.GET(authRouter, "/wallet/verify", s.authDomain.WalletVerify)
		router.POST(authRouter, "/refresh", s.authDomain.Refresh)
	}

	// These following APIs need authentication with only Access Token.
	onlyTokenAuthRouter := s.router.Branch()
	authVerifier := middleware.NewAuthVerifier().WithAccessToken()
	onlyTokenAuthRouter.Before(authVerifier.Middleware())
	{
		// Link account API
		router.POST(onlyTokenAuthRouter, "/linkOAuth2", s.authDomain.OAuth2Link)
		router.POST(onlyTokenAuthRouter, "/linkWallet", s.authDomain.WalletLink)
		router.POST(onlyTokenAuthRouter, "/linkTelegram", s.authDomain.TelegramLink)

		// User API
		router.GET(onlyTokenAuthRouter, "/getUser", s.userDomain.GetUser)
		router.POST(onlyTokenAuthRouter, "/follow", s.userDomain.FollowProject)

		// Project API
		router.GET(onlyTokenAuthRouter, "/getMyListProject", s.projectDomain.GetMyList)
		router.GET(onlyTokenAuthRouter, "/getFollowingProjects", s.projectDomain.GetFollowing)
		router.POST(onlyTokenAuthRouter, "/createProject", s.projectDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateProjectByID", s.projectDomain.UpdateByID)
		router.POST(onlyTokenAuthRouter, "/deleteProjectByID", s.projectDomain.DeleteByID)
		router.POST(onlyTokenAuthRouter, "/updateDiscord", s.projectDomain.UpdateDiscord)

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
		router.POST(onlyTokenAuthRouter, "/updateQuest", s.questDomain.Update)

		// Category API
		router.GET(onlyTokenAuthRouter, "/getListCategory", s.categoryDomain.GetList)
		router.POST(onlyTokenAuthRouter, "/createCategory", s.categoryDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateCategoryByID", s.categoryDomain.UpdateByID)
		router.POST(onlyTokenAuthRouter, "/deleteCategoryByID", s.categoryDomain.DeleteByID)

		// Collaborator API
		router.GET(onlyTokenAuthRouter, "/getListCollaborator", s.collaboratorDomain.GetList)
		router.POST(onlyTokenAuthRouter, "/createCollaborator", s.collaboratorDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateCollaboratorByID", s.collaboratorDomain.UpdateRole)
		router.POST(onlyTokenAuthRouter, "/deleteCollaboratorByID", s.collaboratorDomain.Delete)

		// Claimed Quest API
		router.POST(onlyTokenAuthRouter, "/claim", s.claimedQuestDomain.Claim)

		// Image API
		router.POST(onlyTokenAuthRouter, "/uploadImage", s.fileDomain.UploadImage)
		router.POST(onlyTokenAuthRouter, "/uploadAvatar", s.fileDomain.UploadAvatar)

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
	router.GET(s.router, "/getQuest", s.questDomain.Get)
	router.GET(s.router, "/getListQuest", s.questDomain.GetList)
	router.GET(s.router, "/getListProject", s.projectDomain.GetList)
	router.GET(s.router, "/getProjectByID", s.projectDomain.GetByID)
	router.GET(s.router, "/getInvite", s.userDomain.GetInvite)
	router.GET(s.router, "/getLeaderBoard", s.statisticDomain.GetLeaderBoard)
	router.GET(s.router, "/getListProjectByUserID", s.projectDomain.GetListByUserID)
	router.GET(s.router, "/getBadges", s.userDomain.GetBadges)
}
