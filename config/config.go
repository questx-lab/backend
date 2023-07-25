package config

import (
	"fmt"
	"time"
)

type Configs struct {
	Env      string
	LogLevel int

	DomainNameSuffix    string
	Database            DatabaseConfigs
	ApiServer           APIServerConfigs
	GameProxyServer     ServerConfigs
	GameEngineRPCServer RPCServerConfigs
	GameEngineWSServer  ServerConfigs
	GameCenterServer    RPCServerConfigs
	Auth                AuthConfigs
	Session             SessionConfigs
	Storage             S3Configs
	File                FileConfigs
	Quest               QuestConfigs
	Redis               RedisConfigs
	Kafka               KafkaConfigs
	ScyllaDB            ScyllaDBConfigs
	Game                GameConfigs
	SearchServer        SearchServerConfigs
	Blockchain          BlockchainConfigs
	Notification        NotificationConfigs
	Cache               CacheConfigs
}

type DatabaseConfigs struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
	LogLevel string
}

func (d DatabaseConfigs) ConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
	)
}

type ServerConfigs struct {
	Host      string
	Port      string
	Endpoint  string
	AllowCORS []string
}

func (c ServerConfigs) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type APIServerConfigs struct {
	ServerConfigs
	MaxLimit             int
	DefaultLimit         int
	NeedApproveCommunity bool
}

type SessionConfigs struct {
	Secret string
	Name   string
}

type AuthConfigs struct {
	TokenSecret  string
	AccessToken  TokenConfigs
	RefreshToken TokenConfigs

	Google   OAuth2Config
	Twitter  OAuth2Config
	Discord  OAuth2Config
	Telegram TelegramConfigs
}

type OAuth2Config struct {
	Name          string
	VerifyURL     string
	IDField       string
	UsernameField string

	// Only for verifying id token or authorization code. Leave as empty to
	// disable this feature. The issuer must follow OpenID interface.
	Issuer string
	// Use this field to verify authorization code in case the issuer doesn't
	// follow OpenID interface.
	// NOTE: This field cannot be used to verify id token.
	TokenURL string
	ClientID string
}

type TelegramConfigs struct {
	ReclaimDelay time.Duration

	Name            string
	BotToken        string
	LoginExpiration time.Duration
}

type TokenConfigs struct {
	Name       string
	Expiration time.Duration
}

type FileConfigs struct {
	MaxMemory int64
	MaxSize   int64

	AvatarCropHeight uint
	AvatarCropWidth  uint
}

type TwitterConfigs struct {
	ReclaimDelay time.Duration

	AppAccessToken string

	ConsumerAPIKey    string
	ConsumerAPISecret string
	AccessToken       string
	AccessTokenSecret string
}

type DiscordConfigs struct {
	ReclaimDelay time.Duration

	BotToken string
	BotID    string
}

type QuestConfigs struct {
	Twitter  TwitterConfigs
	Dicord   DiscordConfigs
	Telegram TelegramConfigs

	QuizMaxQuestions                 int
	QuizMaxOptions                   int
	InviteReclaimDelay               time.Duration
	InviteCommunityRequiredFollowers int

	InviteCommunityRewardChain        string
	InviteCommunityRewardTokenAddress string
	InviteCommunityRewardAmount       float64
}

type RedisConfigs struct {
	Addr string
}

type KafkaConfigs struct {
	Addr string
}

type ScyllaDBConfigs struct {
	Addr     string
	KeySpace string
}
type GameConfigs struct {
	GameCenterJanitorFrequency     time.Duration
	GameCenterLoadBalanceFrequency time.Duration
	GameEnginePingFrequency        time.Duration
	GameSaveFrequency              time.Duration
	ProxyClientBatchingFrequency   time.Duration

	MaxUsers                 int
	LuckyboxGenerateMaxRetry int

	JoinActionDelay            time.Duration
	MessageActionDelay         time.Duration
	CollectLuckyboxActionDelay time.Duration
	MinLuckyboxEventDuration   time.Duration
	MaxLuckyboxEventDuration   time.Duration
	MaxLuckyboxPerEvent        int

	MessageHistoryLength int
}

type RPCServerConfigs struct {
	ServerConfigs
	RPCName string
}

type SearchServerConfigs struct {
	RPCServerConfigs
	IndexDir string
}

type S3Configs struct {
	Region         string
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	SSLDisabled    bool
}

type BlockchainConfigs struct {
	RPCServerConfigs

	SecretKey                  string
	RefreshConnectionFrequency time.Duration
}

type NotificationConfigs struct {
	EngineRPCServer RPCServerConfigs
	EngineWSServer  ServerConfigs
	ProxyServer     ServerConfigs
}

type CacheConfigs struct {
	TTL time.Duration
}
