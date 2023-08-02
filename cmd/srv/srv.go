package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/migration"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/telegram"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
	"github.com/scylladb/gocqlx/v2"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type srv struct {
	ctx context.Context

	userRepo              repository.UserRepository
	oauth2Repo            repository.OAuth2Repository
	communityRepo         repository.CommunityRepository
	questRepo             repository.QuestRepository
	categoryRepo          repository.CategoryRepository
	claimedQuestRepo      repository.ClaimedQuestRepository
	followerRepo          repository.FollowerRepository
	followerRoleRepo      repository.FollowerRoleRepository
	fileRepo              repository.FileRepository
	apiKeyRepo            repository.APIKeyRepository
	refreshTokenRepo      repository.RefreshTokenRepository
	badgeRepo             repository.BadgeRepository
	badgeDetailRepo       repository.BadgeDetailRepository
	payRewardRepo         repository.PayRewardRepository
	blockchainRepo        repository.BlockChainRepository
	roleRepo              repository.RoleRepository
	chatMessageRepo       repository.ChatMessageRepository
	chatChannelRepo       repository.ChatChannelRepository
	chatMemberRepo        repository.ChatMemberRepository
	chatReactionRepo      repository.ChatReactionRepository
	chatChannelBucketRepo repository.ChatChannelBucketRepository
	lotteryRepo           repository.LotteryRepository

	userDomain         domain.UserDomain
	authDomain         domain.AuthDomain
	communityDomain    domain.CommunityDomain
	questDomain        domain.QuestDomain
	categoryDomain     domain.CategoryDomain
	roleDomain         domain.RoleDomain
	claimedQuestDomain domain.ClaimedQuestDomain
	fileDomain         domain.FileDomain
	apiKeyDomain       domain.APIKeyDomain
	statisticDomain    domain.StatisticDomain
	followerDomain     domain.FollowerDomain
	payRewardDomain    domain.PayRewardDomain
	badgeDomain        domain.BadgeDomain
	blockchainDomain   domain.BlockchainDomain
	chatDomain         domain.ChatDomain
	lotteryDomain      domain.LotteryDomain

	roleVerifier    *common.CommunityRoleVerifier
	questFactory    questclaim.Factory
	publisher       pubsub.Publisher
	storage         storage.Storage
	scyllaDBSession gocqlx.Session

	leaderboard      statistic.Leaderboard
	badgeManager     *badge.Manager
	twitterEndpoint  twitter.IEndpoint
	discordEndpoint  discord.IEndpoint
	telegramEndpoint telegram.IEndpoint

	redisClient xredis.Client
}

func (s *srv) loadConfig() config.Configs {
	return config.Configs{
		Env:              getEnv("ENV", "local"),
		LogLevel:         parseLogLevel(getEnv("LOG_LEVEL", "INFO")),
		DomainNameSuffix: getEnv("K8S_DOMAIN_NAME_SUFFIX", ""),
		ApiServer: config.APIServerConfigs{
			MaxLimit:             parseInt(getEnv("API_MAX_LIMIT", "50")),
			DefaultLimit:         parseInt(getEnv("API_DEFAULT_LIMIT", "1")),
			NeedApproveCommunity: parseBool(getEnv("API_NEED_APPROVE_COMMUNITY", "false")),
			ServerConfigs: config.ServerConfigs{
				Host:      getEnv("API_HOST", ""),
				Port:      getEnv("API_PORT", "8080"),
				AllowCORS: parseArray(getEnv("API_ALLOW_CORS", "http://localhost:3000")),
			},
		},
		SearchServer: config.SearchServerConfigs{
			RPCServerConfigs: config.RPCServerConfigs{
				ServerConfigs: config.ServerConfigs{
					Host:     getEnv("SEARCH_SERVER_HOST", ""),
					Port:     getEnv("SEARCH_SERVER_PORT", "8082"),
					Endpoint: getEnv("SEARCH_SERVER_ENDPOINT", "http://localhost:8082"),
				},
				RPCName: "searchIndexer",
			},
			IndexDir: getEnv("SEARCH_SERVER_INDEX_DIR", "searchindex"),
		},
		Blockchain: config.BlockchainConfigs{
			RPCServerConfigs: config.RPCServerConfigs{
				ServerConfigs: config.ServerConfigs{
					Host:     getEnv("BLOCKCHAIN_HOST", ""),
					Port:     getEnv("BLOCKCHAIN_PORT", "8086"),
					Endpoint: getEnv("BLOCKCHAIN_ENDPOINT", "http://localhost:8086"),
				},
				RPCName: "blockchain",
			},
			SecretKey:                  getEnv("BLOCKCHAIN_SECRET_KEY", "eth_super_super_secret_key_should_be_32_bytes"),
			RefreshConnectionFrequency: parseDuration(getEnv("BLOCKCHAIN_REFRESH_CONENCTION_FREQUENCY", "5m")),
		},
		Notification: config.NotificationConfigs{
			EngineRPCServer: config.RPCServerConfigs{
				ServerConfigs: config.ServerConfigs{
					Host:     getEnv("NOTIFICATION_ENGINE_RPC_HOST", ""),
					Port:     getEnv("NOTIFICATION_ENGINE_RPC_PORT", "8087"),
					Endpoint: getEnv("NOTIFICATION_ENGINE_RPC_ENDPOINT", "http://localhost:8087"),
				},
				RPCName: "notificationEngine",
			},
			EngineWSServer: config.ServerConfigs{
				Host:     getEnv("NOTIFICATION_ENGINE_WS_HOST", ""),
				Port:     getEnv("NOTIFICATION_ENGINE_WS_PORT", "8088"),
				Endpoint: getEnv("NOTIFICATION_ENGINE_WS_ENDPOINT", "ws://localhost:8088/proxy"),
			},
			ProxyServer: config.ServerConfigs{
				Host:      getEnv("NOTIFICATION_PROXY_HOST", ""),
				Port:      getEnv("NOTIFICATION_PROXY_PORT", "8089"),
				AllowCORS: strings.Split("NOTIFICATION_PROXY_ALLOW_CORS", "http://localhost:4000"),
			},
		},
		Auth: config.AuthConfigs{
			TokenSecret: getEnv("TOKEN_SECRET", "token_secret"),
			AccessToken: config.TokenConfigs{
				Name:       "access_token",
				Expiration: parseDuration(getEnv("ACCESS_TOKEN_DURATION", "5m")),
			},
			RefreshToken: config.TokenConfigs{
				Name:       "refresh_token",
				Expiration: parseDuration(getEnv("REFRESH_TOKEN_DURATION", "20m")),
			},
			Google: config.OAuth2Config{
				Name:      "google",
				VerifyURL: "https://www.googleapis.com/oauth2/v1/userinfo",
				IDField:   "email",
				ClientID:  getEnv("GOOGLE_CLIENT_ID", "google-client-id"),
				Issuer:    "https://accounts.google.com",
			},
			Twitter: config.OAuth2Config{
				Name:          "twitter",
				VerifyURL:     "https://api.twitter.com/2/users/me",
				IDField:       "data.id",
				UsernameField: "data.username",
				ClientID:      getEnv("TWITTER_CLIENT_ID", "twitter-client-id"),
				TokenURL:      "https://api.twitter.com/2/oauth2/token",
			},
			Discord: config.OAuth2Config{
				Name:      "discord",
				VerifyURL: "https://discord.com/api/users/@me",
				IDField:   "id",
			},
			Telegram: config.TelegramConfigs{
				Name:            "telegram",
				BotToken:        getEnv("TELEGRAM_BOT_TOKEN", "telegram-bot-token"),
				LoginExpiration: parseDuration(getEnv("TELEGRAM_LOGIN_EXPIRATION", "10s")),
			},
		},
		Database: config.DatabaseConfigs{
			Host:     getEnv("MYSQL_HOST", "mysql"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			User:     getEnv("MYSQL_USER", "mysql"),
			Password: getEnv("MYSQL_PASSWORD", "mysql"),
			Database: getEnv("MYSQL_DATABASE", "questx"),
			LogLevel: getEnv("DATABASE_LOG_LEVEL", "error"),
		},
		Session: config.SessionConfigs{
			Secret: getEnv("AUTH_SESSION_SECRET", "secret"),
			Name:   "auth_session",
		},
		Storage: config.S3Configs{
			Region:         getEnv("STORAGE_REGION", "auto"),
			Endpoint:       getEnv("STORAGE_ENDPOINT", "http://localhost:9000"),
			PublicEndpoint: getEnv("STORAGE_PUBLIC_ENDPOINT", "http://localhost:9000"),
			AccessKey:      getEnv("STORAGE_ACCESS_KEY", "access_key"),
			SecretKey:      getEnv("STORAGE_SECRET_KEY", "secret_key"),
			SSLDisabled:    parseBool(getEnv("STORAGE_SSL_DISABLE", "true")),
		},
		File: config.FileConfigs{
			MaxMemory:        int64(parseSizeToByte(getEnv("MAX_MEMORY_MULTIPART_FORM", "2M"))),
			MaxSize:          int64(parseSizeToByte(getEnv("MAX_FILE_SIZE", "2M"))),
			AvatarCropHeight: uint(parseInt(getEnv("AVATAR_CROP_HEIGHT", "512"))),
			AvatarCropWidth:  uint(parseInt(getEnv("AVATAR_CROP_WIDTH", "512"))),
		},
		Quest: config.QuestConfigs{
			Twitter: config.TwitterConfigs{
				APIEndpoints:      parseArray(getEnv("TWITTER_SCRAPER_ENDPOINTS", "http://localhost:5000")),
				ReclaimDelay:      parseDuration(getEnv("TWITTER_RECLAIM_DELAY", "15m")),
				AppAccessToken:    getEnv("TWITTER_APP_ACCESS_TOKEN", "app_access_token"),
				ConsumerAPIKey:    getEnv("TWITTER_CONSUMER_API_KEY", "consumer_key"),
				ConsumerAPISecret: getEnv("TWITTER_CONSUMER_API_SECRET", "comsumer_secret"),
				AccessToken:       getEnv("TWITTER_ACCESS_TOKEN", "access_token"),
				AccessTokenSecret: getEnv("TWITTER_ACCESS_TOKEN_SECRET", "access_token_secret"),
			},
			Dicord: config.DiscordConfigs{
				ReclaimDelay: parseDuration(getEnv("DISCORD_RECLAIM_DELAY", "15m")),
				BotToken:     getEnv("DISCORD_BOT_TOKEN", "discord_bot_token"),
				BotID:        getEnv("DISCORD_BOT_ID", "discord_bot_id"),
			},
			Telegram: config.TelegramConfigs{
				ReclaimDelay: parseDuration(getEnv("TELEGRAM_RECLAIM_DELAY", "15m")),
				BotToken:     getEnv("TELEGRAM_BOT_TOKEN", "telegram-bot-token"),
			},
			QuizMaxQuestions:                 parseInt(getEnv("QUIZ_MAX_QUESTIONS", "10")),
			QuizMaxOptions:                   parseInt(getEnv("QUIZ_MAX_OPTIONS", "10")),
			InviteReclaimDelay:               parseDuration(getEnv("INVITE_RECLAIM_DELAY", "1m")),
			InviteCommunityRequiredFollowers: parseInt(getEnv("INVITE_COMMUNITY_REQUIRED_FOLLOWERS", "10000")),
			InviteCommunityRewardChain: getEnv("INVITE_COMMUNITY_REWARD_CHAIN",
				"avaxc-testnet"),
			InviteCommunityRewardTokenAddress: getEnv("INVITE_COMMUNITY_REWARD_TOKEN_ADDRESS",
				"0x251AA5624b902a8183C6E991832dA0f0Fd18D5aB"),
			InviteCommunityRewardAmount: parseFloat64(getEnv("INVITE_COMMUNITY_REWARD_AMOUNT", "50")),
		},
		Redis: config.RedisConfigs{
			Addr: getEnv("REDIS_ADDRESS", "localhost:6379"),
		},
		Kafka: config.KafkaConfigs{
			Addr: getEnv("KAFKA_ADDRESS", "localhost:9092"),
		},
		ScyllaDB: config.ScyllaDBConfigs{
			Addr:     getEnv("SCYLLA_DB_ADDRESS", "localhost:9042"),
			KeySpace: getEnv("SCYLLA_DB_KEY_SPACE", "xquest"),
		},
		Cache: config.CacheConfigs{
			TTL: parseDuration(getEnv("CACHE_TTL", "1h")),
		},
		PrometheusServer: config.ServerConfigs{
			Host:     getEnv("PROMETHEUS_HOST", ""),
			Port:     getEnv("PROMETHEUS_PORT", "9000"),
			Endpoint: getEnv("PROMETHEUS_ENDPOINT", "http://localhost:9000"),
		},
		Chat: config.ChatConfigs{
			MessageXP:      parseInt(getEnv("CHAT_MESSAGE_XP", "1")),
			ImageMessageXP: parseInt(getEnv("CHAT_IMAGE_MESSAGE_XP", "2")),
			VideoMessageXP: parseInt(getEnv("CHAT_VIDEO_MESSAGE_XP", "3")),
			ReactionXP:     parseInt(getEnv("CHAT_REACTION_XP", "1")),
		},
	}
}

func (s *srv) newDatabase() *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       xcontext.Configs(s.ctx).Database.ConnectionString(), // data source name
		DefaultStringSize:         256,                                                 // default size for string fields
		DisableDatetimePrecision:  true,                                                // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,                                                // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,                                                // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false,                                               // auto configure based on currently MySQL version
	}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(parseDatabaseLogLevel(xcontext.Configs(s.ctx).Database.LogLevel)),
	})
	if err != nil {
		panic(err)
	}

	return db
}

func (s *srv) migrateDB() {
	if err := migration.Migrate(s.ctx, s.twitterEndpoint); err != nil {
		panic(err)
	}
}

func (s *srv) loadScyllaDB() {
	retryPolicy := &gocql.ExponentialBackoffRetryPolicy{
		Min:        time.Second,
		Max:        10 * time.Second,
		NumRetries: 5,
	}
	cluster := gocql.NewCluster(xcontext.Configs(s.ctx).ScyllaDB.Addr)
	cluster.Keyspace = xcontext.Configs(s.ctx).ScyllaDB.KeySpace
	cluster.Timeout = 5 * time.Second
	cluster.RetryPolicy = retryPolicy
	cluster.Consistency = gocql.Quorum
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		panic(err)
	}

	s.scyllaDBSession = session
	if err := migration.MigrateScyllaDB(s.ctx, s.scyllaDBSession); err != nil {
		panic(err)
	}
}

func (s *srv) loadStorage() {
	s.storage = storage.NewS3Storage(xcontext.Configs(s.ctx).Storage)
}

func (s *srv) loadEndpoint() {
	s.twitterEndpoint = twitter.New(xcontext.Configs(s.ctx).Quest.Twitter)
	s.discordEndpoint = discord.New(xcontext.Configs(s.ctx).Quest.Dicord)
	s.telegramEndpoint = telegram.New(xcontext.Configs(s.ctx).Quest.Telegram)
}

func (s *srv) loadRedisClient() {
	var err error
	s.redisClient, err = xredis.NewClient(s.ctx)
	if err != nil {
		panic(err)
	}
}

func (s *srv) loadLeaderboard() {
	s.leaderboard = statistic.New(s.claimedQuestRepo, s.redisClient)
}

func (s *srv) loadRepos(searchCaller client.SearchCaller) {
	s.userRepo = repository.NewUserRepository(s.redisClient)
	s.oauth2Repo = repository.NewOAuth2Repository()
	s.communityRepo = repository.NewCommunityRepository(searchCaller)
	s.questRepo = repository.NewQuestRepository(searchCaller)
	s.categoryRepo = repository.NewCategoryRepository()
	s.roleRepo = repository.NewRoleRepository()
	s.claimedQuestRepo = repository.NewClaimedQuestRepository()
	s.followerRepo = repository.NewFollowerRepository()
	s.followerRoleRepo = repository.NewFollowerRoleRepository()
	s.fileRepo = repository.NewFileRepository()
	s.apiKeyRepo = repository.NewAPIKeyRepository()
	s.refreshTokenRepo = repository.NewRefreshTokenRepository()
	s.badgeRepo = repository.NewBadgeRepository()
	s.badgeDetailRepo = repository.NewBadgeDetailRepository()
	s.payRewardRepo = repository.NewPayRewardRepository()
	s.blockchainRepo = repository.NewBlockChainRepository()
	s.chatMessageRepo = repository.NewChatMessageRepository(s.scyllaDBSession)
	s.chatChannelRepo = repository.NewChatChannelRepository()
	s.chatMemberRepo = repository.NewChatMemberRepository()
	s.chatReactionRepo = repository.NewChatReactionRepository(s.scyllaDBSession)
	s.chatChannelBucketRepo = repository.NewChatBucketRepository(s.scyllaDBSession)
	s.lotteryRepo = repository.NewLotteryRepository()
}

func (s *srv) loadBadgeManager() {
	s.badgeManager = badge.NewManager(
		s.badgeRepo,
		s.badgeDetailRepo,
		badge.NewSharpScoutBadgeScanner(s.badgeRepo, s.followerRepo),
		badge.NewRainBowBadgeScanner(s.badgeRepo, s.followerRepo),
		badge.NewQuestWarriorBadgeScanner(s.badgeRepo, s.followerRepo),
	)
}

func (s *srv) loadDomains(
	blockchainCaller client.BlockchainCaller,
	notificationEngineCaller client.NotificationEngineCaller,
) {
	cfg := xcontext.Configs(s.ctx)

	var oauth2Services []authenticator.IOAuth2Service
	oauth2Services = append(oauth2Services, authenticator.NewOAuth2Service(s.ctx, cfg.Auth.Google))
	oauth2Services = append(oauth2Services, authenticator.NewOAuth2Service(s.ctx, cfg.Auth.Twitter))
	oauth2Services = append(oauth2Services, authenticator.NewOAuth2Service(s.ctx, cfg.Auth.Discord))

	s.roleVerifier = common.NewCommunityRoleVerifier(s.followerRoleRepo, s.roleRepo, s.userRepo)
	s.questFactory = questclaim.NewFactory(s.claimedQuestRepo, s.questRepo, s.communityRepo,
		s.followerRepo, s.oauth2Repo, s.userRepo, s.payRewardRepo, s.blockchainRepo,
		s.lotteryRepo, s.twitterEndpoint, s.discordEndpoint, s.telegramEndpoint,
	)

	s.authDomain = domain.NewAuthDomain(s.ctx, s.userRepo, s.refreshTokenRepo, s.oauth2Repo,
		oauth2Services, s.twitterEndpoint, s.storage)
	s.userDomain = domain.NewUserDomain(s.userRepo, s.oauth2Repo, s.followerRepo, s.followerRoleRepo,
		s.communityRepo, s.claimedQuestRepo, s.badgeManager, s.storage, notificationEngineCaller)
	s.communityDomain = domain.NewCommunityDomain(s.communityRepo, s.followerRepo, s.followerRoleRepo,
		s.userRepo, s.questRepo, s.oauth2Repo, s.chatChannelRepo, s.roleRepo,
		s.discordEndpoint, s.storage, oauth2Services, notificationEngineCaller,
		s.roleVerifier)
	s.questDomain = domain.NewQuestDomain(s.questRepo, s.communityRepo, s.categoryRepo,
		s.userRepo, s.claimedQuestRepo, s.followerRepo, s.leaderboard, s.roleVerifier, s.questFactory)
	s.categoryDomain = domain.NewCategoryDomain(s.categoryRepo, s.communityRepo, s.roleVerifier)
	s.claimedQuestDomain = domain.NewClaimedQuestDomain(s.claimedQuestRepo, s.questRepo,
		s.followerRepo, s.followerRoleRepo, s.userRepo, s.communityRepo, s.categoryRepo,
		s.badgeManager, s.leaderboard, s.roleVerifier, notificationEngineCaller, s.questFactory)
	s.fileDomain = domain.NewFileDomain(s.storage, s.fileRepo)
	s.apiKeyDomain = domain.NewAPIKeyDomain(s.apiKeyRepo, s.communityRepo, s.roleVerifier)
	s.statisticDomain = domain.NewStatisticDomain(s.claimedQuestRepo, s.followerRepo, s.userRepo,
		s.communityRepo, s.leaderboard)
	s.followerDomain = domain.NewFollowerDomain(s.followerRepo, s.followerRoleRepo, s.communityRepo,
		s.roleRepo, s.userRepo, s.questRepo, s.roleVerifier)
	s.blockchainDomain = domain.NewBlockchainDomain(s.blockchainRepo, s.communityRepo, blockchainCaller)
	s.payRewardDomain = domain.NewPayRewardDomain(s.payRewardRepo, s.blockchainRepo, s.communityRepo,
		s.lotteryRepo, s.questFactory)
	s.badgeDomain = domain.NewBadgeDomain(s.badgeRepo, s.badgeDetailRepo, s.communityRepo, s.badgeManager)
	s.chatDomain = domain.NewChatDomain(s.communityRepo, s.chatMessageRepo, s.chatChannelRepo,
		s.chatReactionRepo, s.chatMemberRepo, s.chatChannelBucketRepo, s.userRepo, s.followerRepo,
		notificationEngineCaller, s.roleVerifier)
	s.lotteryDomain = domain.NewLotteryDomain(s.lotteryRepo, s.followerRepo, s.communityRepo,
		s.roleVerifier, s.questFactory)
	s.roleDomain = domain.NewRoleDomain(s.roleRepo, s.communityRepo, s.roleVerifier)
}

func (s *srv) loadPublisher() {
	s.publisher = kafka.NewPublisher(uuid.NewString(), []string{xcontext.Configs(s.ctx).Kafka.Addr})
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		value = fallback
	}
	value = strings.Trim(value, " ")
	return strings.Trim(value, "\x0d")
}

func parseDuration(s string) time.Duration {
	duration, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}

	return duration
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i
}

func parseFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}

	return f
}

func parseDatabaseLogLevel(s string) gormlogger.LogLevel {
	switch s {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "warn":
		return gormlogger.Warn
	case "info":
		return gormlogger.Info
	}

	panic(fmt.Sprintf("invalid gorm log level %s", s))
}

func parseLogLevel(s string) int {
	s = strings.ToLower(s)

	switch s {
	case "debug":
		return logger.DEBUG
	case "info":
		return logger.INFO
	case "warn":
		return logger.WARNING
	case "error":
		return logger.ERROR
	case "silent":
		return logger.SILENCE
	}

	panic(fmt.Sprintf("invalid log level %s", s))
}

func parseBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		panic(err)
	}

	return b
}

func parseSizeToByte(s string) int {
	if s[len(s)-1] >= '0' && s[len(s)-1] <= '9' {
		return parseInt(s)
	}

	n := parseInt(s[:len(s)-1])
	switch s[len(s)-1] {
	case 'k', 'K':
		return n * 1024
	case 'm', 'M':
		return n * 1024 * 1024
	case 'g', 'G':
		return n * 1024 * 1024 * 1024
	}

	panic(fmt.Sprintf("Invalid value of %s", s))
}

func parseArray(s string) []string {
	s = strings.Trim(s, " ")
	if s == "" {
		return []string{}
	}

	return strings.Split(s, ",")
}
