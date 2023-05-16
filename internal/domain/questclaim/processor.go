package questclaim

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

// URL Processor
type urlProcessor struct{}

func newURLProcessor(xcontext.Context, map[string]any) (*urlProcessor, error) {
	return &urlProcessor{}, nil
}

func (urlProcessor) RetryAfter() time.Duration {
	return 0
}

func (p *urlProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	_, err := url.ParseRequestURI(input)
	if err != nil {
		return Rejected, err
	}

	return NeedManualReview, nil
}

// VisitLink Processor
type visitLinkProcessor struct {
	Link string `mapstructure:"link" structs:"link"`
}

func newVisitLinkProcessor(ctx xcontext.Context, data map[string]any, needParse bool) (*visitLinkProcessor, error) {
	visitLink := visitLinkProcessor{}
	err := mapstructure.Decode(data, &visitLink)
	if err != nil {
		return nil, err
	}

	if needParse {
		if visitLink.Link == "" {
			return nil, errors.New("not found link in validation data")
		}

		_, err = url.ParseRequestURI(visitLink.Link)
		if err != nil {
			return nil, err
		}
	}

	return &visitLink, nil
}

func (visitLinkProcessor) RetryAfter() time.Duration {
	return 0
}

func (v *visitLinkProcessor) GetActionForClaim(xcontext.Context, string) (ActionForClaim, error) {
	return Accepted, nil
}

// Text Processor
// TODO: Add retry_after when the claimed quest is rejected by auto validate.
type textProcessor struct {
	AutoValidate       bool          `mapstructure:"auto_validate" structs:"auto_validate"`
	Answer             string        `mapstructure:"answer" structs:"answer"`
	RetryAfterDuration time.Duration `mapstructure:"retry_after" structs:"retry_after"`
}

func newTextProcessor(ctx xcontext.Context, data map[string]any, needParse bool) (*textProcessor, error) {
	text := textProcessor{}
	err := mapstructure.Decode(data, &text)
	if err != nil {
		return nil, err
	}

	if needParse {
		if text.AutoValidate && text.Answer == "" {
			return nil, errors.New("must provide answer if the quest is automatically validated")
		}
	}

	return &text, nil
}

func (p textProcessor) RetryAfter() time.Duration {
	return p.RetryAfterDuration
}

func (p *textProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	if !p.AutoValidate {
		return NeedManualReview, nil
	}

	if p.Answer != input {
		return Rejected, nil
	}

	return Accepted, nil
}

// Quiz Processor
type quiz struct {
	Question string   `mapstructure:"question" structs:"question"`
	Options  []string `mapstructure:"options" structs:"options"`
	Answers  []string `mapstructure:"answers" structs:"answers"`
}

type quizAnswers struct {
	Answers []string `json:"answers"`
}

type quizProcessor struct {
	Quizs              []quiz        `mapstructure:"quizs" structs:"quizs"`
	RetryAfterDuration time.Duration `mapstructure:"retry_after" structs:"retry_after"`
}

func newQuizProcessor(ctx xcontext.Context, data map[string]any, needParse bool) (*quizProcessor, error) {
	quiz := quizProcessor{}
	err := mapstructure.Decode(data, &quiz)
	if err != nil {
		return nil, err
	}

	if needParse {
		if len(quiz.Quizs) > ctx.Configs().Quest.QuizMaxQuestions {
			return nil, errors.New("too many questions")
		}

		for i, q := range quiz.Quizs {
			if len(q.Options) < 2 {
				return nil, errors.New("provide at least two options")
			}

			if len(q.Options) > ctx.Configs().Quest.QuizMaxOptions {
				return nil, errors.New("too many options")
			}

			if len(q.Answers) == 0 || len(q.Answers) > ctx.Configs().Quest.QuizMaxOptions {
				return nil, fmt.Errorf("invalid number of answers for question %d", i)
			}

			for _, answer := range q.Answers {
				ok := false
				for _, option := range q.Options {
					if answer == option {
						ok = true
						break
					}
				}

				if !ok {
					return nil, errors.New("not found the answer in options")
				}
			}

		}
	}

	return &quiz, nil
}

func (p quizProcessor) RetryAfter() time.Duration {
	return p.RetryAfterDuration
}

func (p *quizProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	answers := quizAnswers{}
	err := json.Unmarshal([]byte(input), &answers)
	if err != nil {
		ctx.Logger().Debugf("Cannot unmarshal input: %v", err)
		return Rejected, errorx.Unknown
	}

	if len(answers.Answers) != len(p.Quizs) {
		return Rejected, errorx.New(errorx.BadRequest, "Invalid number of answers")
	}

	for i, answer := range answers.Answers {
		ok := false
		for _, correctAnswer := range p.Quizs[i].Answers {
			if answer == correctAnswer {
				ok = true
			}
		}

		if !ok {
			return Rejected, nil
		}
	}

	return Accepted, nil
}

// Image Processor
type imageProcessor struct{}

func newImageProcessor(xcontext.Context, map[string]any) (*imageProcessor, error) {
	return &imageProcessor{}, nil
}

func (imageProcessor) RetryAfter() time.Duration {
	return 0
}

func (p *imageProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	// TODO: Input is a link of image, need to validate the image.
	return NeedManualReview, nil
}

// Empty Processor
type emptyProcessor struct{}

func newEmptyProcessor(xcontext.Context, map[string]any) (*emptyProcessor, error) {
	return &emptyProcessor{}, nil
}

func (emptyProcessor) RetryAfter() time.Duration {
	return 0
}

func (p *emptyProcessor) GetActionForClaim(xcontext.Context, string) (ActionForClaim, error) {
	return Accepted, nil
}

// Twitter Follow Processor
type twitterFollowProcessor struct {
	TwitterHandle string `mapstructure:"twitter_handle" structs:"twitter_handle"`

	retryAfter time.Duration
	target     twitterUser
	factory    Factory
}

func newTwitterFollowProcessor(
	ctx xcontext.Context, factory Factory, data map[string]any, needParse bool,
) (*twitterFollowProcessor, error) {
	twitterFollow := twitterFollowProcessor{}
	err := mapstructure.Decode(data, &twitterFollow)
	if err != nil {
		return nil, err
	}

	target, err := parseTwitterUserURL(twitterFollow.TwitterHandle)
	if err != nil {
		return nil, err
	}

	if needParse {
		_, err := factory.twitterEndpoint.GetUser(ctx, target.UserScreenName)
		if err != nil {
			return nil, err
		}
	}

	twitterFollow.retryAfter = ctx.Configs().Quest.Twitter.ReclaimDelay
	twitterFollow.target = target
	twitterFollow.factory = factory
	return &twitterFollow, nil
}

func (p twitterFollowProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *twitterFollowProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	userScreenName := p.factory.getRequestServiceUserID(ctx, ctx.Configs().Auth.Twitter.Name)
	if userScreenName == "" {
		return Rejected, errorx.New(errorx.Unavailable, "User has not connected to twitter")
	}

	b, err := p.factory.twitterEndpoint.CheckFollowing(ctx, userScreenName, p.target.UserScreenName)
	if err != nil {
		if errors.Is(err, twitter.ErrRateLimit) {
			return Rejected, errorx.New(errorx.TooManyRequests, "We are busy now, please try again later")
		}

		ctx.Logger().Debugf("Cannot check following: %v", err)
		return Rejected, errorx.New(errorx.Unavailable, "Invalid twitter response")
	}

	if !b {
		return Rejected, nil
	}

	return Accepted, nil
}

// Twitter Reaction Processsor
type twitterReactionProcessor struct {
	Like    bool `mapstructure:"like" structs:"like"`
	Retweet bool `mapstructure:"retweet" structs:"retweet"`
	Reply   bool `mapstructure:"reply" structs:"reply"`

	TweetURL     string `mapstructure:"tweet_url" structs:"tweet_url"`
	DefaultReply string `mapstructure:"default_reply" structs:"default_reply"`

	retryAfter  time.Duration
	originTweet tweet
	factory     Factory
}

func newTwitterReactionProcessor(
	ctx xcontext.Context, factory Factory, data map[string]any, needParse bool,
) (*twitterReactionProcessor, error) {
	twitterReaction := twitterReactionProcessor{}
	err := mapstructure.Decode(data, &twitterReaction)
	if err != nil {
		return nil, err
	}

	tweet, err := parseTweetURL(twitterReaction.TweetURL)
	if err != nil {
		return nil, err
	}

	if needParse {
		remoteTweet, err := factory.twitterEndpoint.GetTweet(ctx, tweet.TweetID)
		if err != nil {
			return nil, err
		}

		if remoteTweet.AuthorScreenName != tweet.UserScreenName {
			return nil, errors.New("invalid user")
		}
	}

	twitterReaction.retryAfter = ctx.Configs().Quest.Twitter.ReclaimDelay
	twitterReaction.originTweet = tweet
	twitterReaction.factory = factory
	return &twitterReaction, nil
}

func (p twitterReactionProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *twitterReactionProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	userScreenName := p.factory.getRequestServiceUserID(ctx, ctx.Configs().Auth.Twitter.Name)
	if userScreenName == "" {
		return Rejected, errorx.New(errorx.Unavailable, "User has not connected to twitter")
	}

	isLikeAccepted := true
	if p.Like {
		isLikeAccepted = false

		tweets, err := p.factory.twitterEndpoint.GetLikedTweet(ctx, userScreenName)
		if err != nil {
			ctx.Logger().Errorf("Cannot get liked tweet: %v", err)
			return Rejected, errorx.Unknown
		}

		for _, tweet := range tweets {
			if tweet.ID == p.originTweet.TweetID {
				isLikeAccepted = true
			}
		}
	}

	isRetweetAccepted := true
	if p.Retweet {
		isRetweetAccepted = false

		retweets, err := p.factory.twitterEndpoint.GetRetweet(ctx, p.originTweet.TweetID)
		if err != nil {
			ctx.Logger().Errorf("Cannot get retweet: %v", err)
			return Rejected, errorx.Unknown
		}

		for _, retweet := range retweets {
			if retweet.AuthorScreenName == userScreenName {
				isRetweetAccepted = true
			}
		}
	}

	isReplyAccepted := true
	if p.Reply {
		isReplyAccepted = false

		replyTweet, err := parseTweetURL(input)
		if err != nil {
			return Rejected, errorx.New(errorx.BadRequest, "Invalid input")
		}

		if replyTweet.UserScreenName == userScreenName {
			_, err := p.factory.twitterEndpoint.GetTweet(ctx, replyTweet.TweetID)
			if err != nil {
				ctx.Logger().Debugf("Cannot get tweet api: %v", err)
				return Rejected, errorx.Unknown
			}

			isReplyAccepted = true
		}
	}

	if isLikeAccepted && isReplyAccepted && isRetweetAccepted {
		return Accepted, nil
	}

	return Rejected, nil
}

// Twitter Tweet Processor
type twitterTweetProcessor struct {
	IncludedWords []string `mapstructure:"included_words" structs:"included_words"`
	DefaultTweet  string   `mapstructure:"default_tweet" structs:"default_tweet"`

	retryAfter time.Duration
	factory    Factory
}

func newTwitterTweetProcessor(
	ctx xcontext.Context, factory Factory, data map[string]any,
) (*twitterTweetProcessor, error) {
	twitterTweet := twitterTweetProcessor{}
	err := mapstructure.Decode(data, &twitterTweet)
	if err != nil {
		return nil, err
	}

	twitterTweet.retryAfter = ctx.Configs().Quest.Twitter.ReclaimDelay
	twitterTweet.factory = factory
	return &twitterTweet, nil
}

func (p twitterTweetProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *twitterTweetProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	tw, err := parseTweetURL(input)
	if err != nil {
		ctx.Logger().Debugf("Cannot parse tweet url: %v", err)
		return Rejected, errorx.New(errorx.BadRequest, "Invalid tweet url")
	}

	userScreenName := p.factory.getRequestServiceUserID(ctx, ctx.Configs().Auth.Twitter.Name)
	if userScreenName == "" {
		return Rejected, errorx.New(errorx.Unavailable, "User has not connected to twitter")
	}

	if tw.UserScreenName != userScreenName {
		return Rejected, nil
	}

	resp, err := p.factory.twitterEndpoint.GetTweet(ctx, tw.TweetID)
	if err != nil {
		ctx.Logger().Debugf("Cannot get tweet: %v", err)
		return Rejected, nil
	}

	if resp.AuthorScreenName != tw.UserScreenName {
		return Rejected, nil
	}

	for _, word := range p.IncludedWords {
		if !strings.Contains(resp.Text, word) {
			return Rejected, nil
		}
	}

	return Accepted, nil
}

// Twitter Join Space Processsor
type twitterJoinSpaceProcessor struct {
	SpaceURL string `mapstructure:"space_url" structs:"space_url"`

	factory Factory
}

func newTwitterJoinSpaceProcessor(
	ctx xcontext.Context, factory Factory, data map[string]any,
) (*twitterJoinSpaceProcessor, error) {
	twitterJoinSpace := twitterJoinSpaceProcessor{}
	err := mapstructure.Decode(data, &twitterJoinSpace)
	if err != nil {
		return nil, err
	}

	_, err = url.ParseRequestURI(twitterJoinSpace.SpaceURL)
	if err != nil {
		return nil, err
	}

	twitterJoinSpace.factory = factory
	return &twitterJoinSpace, nil
}

func (p twitterJoinSpaceProcessor) RetryAfter() time.Duration {
	return 0
}

func (p *twitterJoinSpaceProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	return Accepted, nil
}

// Join Discord Processor
type joinDiscordProcessor struct {
	InviteLink string `mapstructure:"invite_link" structs:"invite_link"`
	GuildID    string `mapstructure:"guild_id" structs:"guild_id"`

	retryAfter time.Duration
	factory    Factory
}

func newJoinDiscordProcessor(
	ctx xcontext.Context,
	factory Factory,
	quest entity.Quest,
	data map[string]any,
	needParse bool,
) (*joinDiscordProcessor, error) {
	joinDiscord := joinDiscordProcessor{}
	err := mapstructure.Decode(data, &joinDiscord)
	if err != nil {
		return nil, err
	}

	if needParse {
		project, err := factory.projectRepo.GetByID(ctx, quest.ProjectID.String)
		if err != nil {
			return nil, err
		}

		if project.Discord == "" {
			return nil, errors.New("not yet connected to discord server")
		}

		hasAddBot, err := factory.discordEndpoint.HasAddedBot(ctx, project.Discord)
		if err != nil {
			return nil, err
		}

		if !hasAddBot {
			return nil, errors.New("server has not added bot yet")
		}

		code, err := parseInviteDiscordURL(joinDiscord.InviteLink)
		if err != nil {
			return nil, err
		}

		err = factory.discordEndpoint.CheckCode(ctx, project.Discord, code)
		if err != nil {
			return nil, err
		}

		joinDiscord.GuildID = project.Discord
	}

	joinDiscord.retryAfter = ctx.Configs().Quest.Dicord.ReclaimDelay
	joinDiscord.factory = factory
	return &joinDiscord, nil
}

func (p joinDiscordProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *joinDiscordProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	requestUserDiscordID := p.factory.getRequestServiceUserID(ctx, ctx.Configs().Auth.Discord.Name)
	if requestUserDiscordID == "" {
		return Rejected, errorx.New(errorx.Unavailable, "User has not connected to discord")
	}

	isJoined, err := p.factory.discordEndpoint.CheckMember(ctx, p.GuildID, requestUserDiscordID)
	if err != nil {
		ctx.Logger().Debugf("Failed to check member: %v", err)
		return Rejected, nil
	}

	if !isJoined {
		return Rejected, nil
	}

	return Accepted, nil
}

// Invite Discord Processor
type inviteDiscordProcessor struct {
	Number  int    `mapstructure:"number" structs:"number"`
	GuildID string `mapstructure:"guild_id" structs:"guild_id"`

	retryAfter time.Duration
	factory    Factory
}

func newInviteDiscordProcessor(
	ctx xcontext.Context,
	factory Factory,
	quest entity.Quest,
	data map[string]any,
	needParse bool,
) (*inviteDiscordProcessor, error) {
	inviteDiscord := inviteDiscordProcessor{}
	err := mapstructure.Decode(data, &inviteDiscord)
	if err != nil {
		return nil, err
	}

	if needParse {
		if inviteDiscord.Number <= 0 {
			return nil, errors.New("number of invites must be positive")
		}

		project, err := factory.projectRepo.GetByID(ctx, quest.ProjectID.String)
		if err != nil {
			return nil, err
		}

		if project.Discord == "" {
			return nil, errors.New("not yet connected to discord server")
		}

		hasAddBot, err := factory.discordEndpoint.HasAddedBot(ctx, project.Discord)
		if err != nil {
			return nil, err
		}

		if !hasAddBot {
			return nil, errors.New("server has not added bot yet")
		}

		inviteDiscord.GuildID = project.Discord
	}

	inviteDiscord.retryAfter = ctx.Configs().Quest.Dicord.ReclaimDelay
	inviteDiscord.factory = factory
	return &inviteDiscord, nil
}

func (p *inviteDiscordProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *inviteDiscordProcessor) GetActionForClaim(
	ctx xcontext.Context, input string,
) (ActionForClaim, error) {
	requestUserDiscordID := p.factory.getRequestServiceUserID(ctx, ctx.Configs().Auth.Discord.Name)
	if requestUserDiscordID == "" {
		return Rejected, errorx.New(errorx.Unavailable, "User has not connected to discord")
	}

	codeString, err := parseInviteDiscordURL(input)
	if err != nil {
		ctx.Logger().Debugf("Cannot parse invite discord url: %v", err)
		return Rejected, errorx.New(errorx.BadRequest, "Invalid input")
	}

	inviteCode, err := p.factory.discordEndpoint.GetCode(ctx, p.GuildID, codeString)
	if err != nil {
		ctx.Logger().Debugf("Failed to get code: %v", err)
		return Rejected, nil
	}

	if inviteCode.Inviter.ID != requestUserDiscordID {
		return Rejected, nil
	}

	if inviteCode.Uses < p.Number {
		return Rejected, nil
	}

	return Accepted, nil
}

// Join Telegram Processor
type joinTelegramProcessor struct {
	InviteLink string `mapstructure:"invite_link" structs:"invite_link"`

	retryAfter time.Duration
	chatID     string
	factory    Factory
}

func newJoinTelegramProcessor(
	ctx xcontext.Context,
	factory Factory,
	quest entity.Quest,
	data map[string]any,
	needParse bool,
) (*joinTelegramProcessor, error) {
	joinTelegram := joinTelegramProcessor{}

	err := mapstructure.Decode(data, &joinTelegram)
	if err != nil {
		return nil, err
	}

	groupName, err := parseInviteTelegramURL(joinTelegram.InviteLink)
	if err != nil {
		return nil, err
	}

	if needParse {
		if joinTelegram.chatID == "" {
			return nil, errors.New("got an empty chat id")
		}

		if err := factory.projectRoleVerifier.Verify(ctx, quest.ProjectID.String, entity.AdminGroup...); err != nil {
			return nil, err
		}

		requestUserID := factory.getRequestServiceUserID(ctx, ctx.Configs().Auth.Telegram.Name)
		if requestUserID == "" {
			return nil, errors.New("quest creator has not connected to telegram")
		}

		admins, err := factory.telegramEndpoint.GetAdministrators(ctx, joinTelegram.chatID)
		if err != nil {
			return nil, err
		}

		isAdmin := false
		for _, admin := range admins {
			if admin.ID == requestUserID {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			return nil, errors.New("quest creator has not the permission to invite users")
		}
	}

	joinTelegram.chatID = "@" + groupName
	joinTelegram.retryAfter = ctx.Configs().Quest.Telegram.ReclaimDelay
	joinTelegram.factory = factory

	return &joinTelegram, nil
}

func (p joinTelegramProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *joinTelegramProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	requestUserID := p.factory.getRequestServiceUserID(ctx, ctx.Configs().Auth.Telegram.Name)
	if requestUserID == "" {
		return Rejected, errorx.New(errorx.Unavailable, "User has not connected telegram")
	}

	_, err := p.factory.telegramEndpoint.GetMember(ctx, p.chatID, requestUserID)
	if err != nil {
		ctx.Logger().Debugf("Cannot get member: %v", err)
		return Rejected, nil
	}

	return Accepted, nil
}

// Invite Processor
type inviteProcessor struct {
	Number int `mapstructure:"number" structs:"number"`

	retryAfter time.Duration
	projectID  string
	factory    Factory
}

func newInviteProcessor(
	ctx xcontext.Context,
	factory Factory,
	quest entity.Quest,
	data map[string]any,
	needParse bool,
) (*inviteProcessor, error) {
	invite := inviteProcessor{}
	err := mapstructure.Decode(data, &invite)
	if err != nil {
		return nil, err
	}

	if needParse {
		if invite.Number <= 0 {
			return nil, errors.New("number must be positive")
		}
	}

	invite.retryAfter = ctx.Configs().Quest.InviteReclaimDelay
	invite.projectID = quest.ProjectID.String
	invite.factory = factory

	return &invite, nil
}

func (p inviteProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *inviteProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	participant, err := p.factory.participantRepo.Get(ctx, xcontext.GetRequestUserID(ctx), p.projectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get participant: %v", err)
		return Rejected, errorx.Unknown
	}

	if participant.InviteCount < uint64(p.Number) {
		return Rejected, nil
	}

	return Accepted, nil
}
