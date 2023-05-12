package config

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/pkg/storage"
)

type Configs struct {
	Env string

	Database        DatabaseConfigs
	ApiServer       APIServerConfigs
	GameProxyServer ServerConfigs
	Auth            AuthConfigs
	Session         SessionConfigs
	Storage         storage.S3Configs
	File            FileConfigs
	Quest           QuestConfigs
	Redis           RedisConfigs
	Kafka           KafkaConfigs
	Game            GameConfigs
}

type DatabaseConfigs struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

func (d *DatabaseConfigs) ConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
	)
}

type ServerConfigs struct {
	Host string
	Port string
}

type APIServerConfigs struct {
	ServerConfigs

	MaxLimit     int
	DefaultLimit int
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
	Name      string
	VerifyURL string
	IDField   string
}

type TelegramConfigs struct {
	Name            string
	BotToken        string
	LoginExpiration time.Duration
}

type TokenConfigs struct {
	Name       string
	Expiration time.Duration
}

type FileConfigs struct {
	MaxSize int64
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
	BotToken string
	BotID    string
}

type QuizConfigs struct {
	MaxQuestions int
	MaxOptions   int
}

type QuestConfigs struct {
	Twitter  TwitterConfigs
	Dicord   DiscordConfigs
	Telegram TelegramConfigs
	Quiz     QuizConfigs
}

type RedisConfigs struct {
	Addr string
}

type KafkaConfigs struct {
	Addr string
}

type GameConfigs struct {
	GameSaveFrequency time.Duration

	MoveActionDelay time.Duration
	InitActionDelay time.Duration
	JoinActionDelay time.Duration
}
