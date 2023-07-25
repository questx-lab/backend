package questclaim

import (
	"context"
	"encoding/json"
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

func newURLProcessor(context.Context, map[string]any) (*urlProcessor, error) {
	return &urlProcessor{}, nil
}

func (urlProcessor) RetryAfter() time.Duration {
	return 0
}

func (p *urlProcessor) GetActionForClaim(ctx context.Context, submissionData string) (ActionForClaim, error) {
	_, err := url.ParseRequestURI(submissionData)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid submission data: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid submission data")
	}

	return NeedManualReview, nil
}

// VisitLink Processor
type visitLinkProcessor struct {
	Link string `mapstructure:"link" structs:"link"`
}

func newVisitLinkProcessor(ctx context.Context, data map[string]any, needParse bool) (*visitLinkProcessor, error) {
	visitLink := visitLinkProcessor{}
	err := mapstructure.Decode(data, &visitLink)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	if needParse {
		if visitLink.Link == "" {
			return nil, errorx.New(errorx.NotFound, "Not found link")
		}

		_, err = url.ParseRequestURI(visitLink.Link)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Invalid link: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid link")
		}
	}

	return &visitLink, nil
}

func (visitLinkProcessor) RetryAfter() time.Duration {
	return 0
}

func (v *visitLinkProcessor) GetActionForClaim(context.Context, string) (ActionForClaim, error) {
	return Accepted, nil
}

// Text Processor
type textProcessor struct {
	AutoValidate       bool          `mapstructure:"auto_validate" structs:"auto_validate"`
	Answer             string        `mapstructure:"answer" structs:"answer"`
	RetryAfterDuration time.Duration `mapstructure:"retry_after" structs:"retry_after"`
}

func newTextProcessor(ctx context.Context, data map[string]any, needParse, includeAnswer bool) (*textProcessor, error) {
	text := textProcessor{}
	err := mapstructure.Decode(data, &text)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	if needParse {
		if text.AutoValidate && text.Answer == "" {
			return nil, errorx.New(errorx.NotFound, "Not found answer in auto validate mode")
		}
	}

	if !includeAnswer {
		text.Answer = ""
	}

	return &text, nil
}

func (p textProcessor) RetryAfter() time.Duration {
	return p.RetryAfterDuration
}

func (p *textProcessor) GetActionForClaim(ctx context.Context, submissionData string) (ActionForClaim, error) {
	if !p.AutoValidate {
		return NeedManualReview, nil
	}

	if p.Answer != submissionData {
		return Rejected.WithMessage("Wrong answer"), nil
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
	Quizzes            []quiz        `mapstructure:"quizzes" structs:"quizzes"`
	RetryAfterDuration time.Duration `mapstructure:"retry_after" structs:"retry_after"`
}

func newQuizProcessor(ctx context.Context, data map[string]any, needParse bool) (*quizProcessor, error) {
	quiz := quizProcessor{}
	err := mapstructure.Decode(data, &quiz)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	cfg := xcontext.Configs(ctx)
	if needParse {
		if len(quiz.Quizzes) > cfg.Quest.QuizMaxQuestions {
			return nil, errorx.New(errorx.BadRequest, "Too many questions")
		}

		for _, q := range quiz.Quizzes {
			if len(q.Options) < 2 {
				return nil, errorx.New(errorx.BadRequest, "Provide at least two options")
			}

			if len(q.Options) > cfg.Quest.QuizMaxOptions {
				return nil, errorx.New(errorx.BadRequest, "Too many options")
			}

			if len(q.Answers) == 0 || len(q.Answers) > len(q.Options) {
				return nil, errorx.New(errorx.BadRequest, "Too many answers")
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
					return nil, errorx.New(errorx.NotFound, "Not found the answer in options")
				}
			}

		}
	}

	return &quiz, nil
}

func (p quizProcessor) RetryAfter() time.Duration {
	return p.RetryAfterDuration
}

func (p *quizProcessor) GetActionForClaim(ctx context.Context, submissionData string) (ActionForClaim, error) {
	answers := quizAnswers{}
	err := json.Unmarshal([]byte(submissionData), &answers)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot unmarshal submission data: %v", err)
		return nil, errorx.Unknown
	}

	if len(answers.Answers) != len(p.Quizzes) {
		return nil, errorx.New(errorx.BadRequest, "Invalid number of answers")
	}

	for i, answer := range answers.Answers {
		ok := false
		for _, correctAnswer := range p.Quizzes[i].Answers {
			if answer == correctAnswer {
				ok = true
			}
		}

		if !ok {
			return Rejected.WithMessage("Wrong answer at quiz %d", i+1), nil
		}
	}

	return Accepted, nil
}

// Image Processor
type imageProcessor struct{}

func newImageProcessor(context.Context, map[string]any) (*imageProcessor, error) {
	return &imageProcessor{}, nil
}

func (imageProcessor) RetryAfter() time.Duration {
	return 0
}

func (p *imageProcessor) GetActionForClaim(ctx context.Context, submissionData string) (ActionForClaim, error) {
	// TODO: Submission data is a link of image, need to validate the image.
	return NeedManualReview, nil
}

// Empty Processor
type emptyProcessor struct{}

func newEmptyProcessor(context.Context, map[string]any) (*emptyProcessor, error) {
	return &emptyProcessor{}, nil
}

func (emptyProcessor) RetryAfter() time.Duration {
	return 0
}

func (p *emptyProcessor) GetActionForClaim(context.Context, string) (ActionForClaim, error) {
	return Accepted, nil
}

// Twitter Follow Processor
type twitterFollowProcessor struct {
	TwitterHandle     string `mapstructure:"twitter_handle" structs:"twitter_handle"`
	TwitterName       string `mapstructure:"twitter_name" structs:"twitter_name"`
	TwitterScreenName string `mapstructure:"twitter_screen_name" structs:"twitter_screen_name"`
	TwitterPhotoURL   string `mapstructure:"twitter_photo_url" structs:"twitter_photo_url"`

	retryAfter time.Duration
	target     twitterUser
	factory    Factory
}

func newTwitterFollowProcessor(
	ctx context.Context, factory Factory, data map[string]any, needParse bool,
) (*twitterFollowProcessor, error) {
	twitterFollow := twitterFollowProcessor{}
	err := mapstructure.Decode(data, &twitterFollow)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	target, err := parseTwitterUserURL(twitterFollow.TwitterHandle)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid twitter handle url: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid twitter handle url")
	}

	if needParse {
		user, err := factory.twitterEndpoint.GetUser(ctx, target.UserScreenName)
		if err != nil {
			xcontext.Logger(ctx).Warnf("Cannot get twitter user: %v", err)
			return nil, errorx.New(errorx.Unavailable, "Cannot verify twitter user")
		}

		twitterFollow.TwitterName = user.Name
		twitterFollow.TwitterScreenName = user.Handle
		twitterFollow.TwitterPhotoURL = user.PhotoURL
	}

	twitterFollow.retryAfter = xcontext.Configs(ctx).Quest.Twitter.ReclaimDelay
	twitterFollow.target = target
	twitterFollow.factory = factory
	return &twitterFollow, nil
}

func (p twitterFollowProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *twitterFollowProcessor) GetActionForClaim(ctx context.Context, submissionData string) (ActionForClaim, error) {
	userScreenName := p.factory.getRequestServiceUserID(ctx, xcontext.Configs(ctx).Auth.Twitter.Name)
	if userScreenName == "" {
		return nil, errorx.New(errorx.Unavailable, "User has not connected to twitter")
	}

	b, err := p.factory.twitterEndpoint.CheckFollowing(ctx, userScreenName, p.target.UserScreenName)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot check following: %v", err)
		return nil, errorx.New(errorx.Unavailable, "Invalid twitter response")
	}

	if !b {
		return Rejected.WithMessage("User has not follow the target"), nil
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
	ctx context.Context, factory Factory, data map[string]any, needParse bool,
) (*twitterReactionProcessor, error) {
	twitterReaction := twitterReactionProcessor{}
	err := mapstructure.Decode(data, &twitterReaction)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	tweet, err := parseTweetURL(twitterReaction.TweetURL)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Invalid tweet url: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid tweet url")
	}

	if needParse {
		remoteTweet, err := factory.twitterEndpoint.GetTweet(ctx, tweet.UserScreenName, tweet.TweetID)
		if err != nil {
			xcontext.Logger(ctx).Warnf("Cannot get tweet: %v", err)
			return nil, errorx.New(errorx.Unavailable, "Cannot verify tweet")
		}

		if remoteTweet.Author != tweet.UserScreenName {
			return nil, errorx.New(errorx.Unavailable, "Invalid tweet url")
		}
	}

	twitterReaction.retryAfter = xcontext.Configs(ctx).Quest.Twitter.ReclaimDelay
	twitterReaction.originTweet = tweet
	twitterReaction.factory = factory
	return &twitterReaction, nil
}

func (p twitterReactionProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *twitterReactionProcessor) GetActionForClaim(ctx context.Context, submissionData string) (ActionForClaim, error) {
	userScreenName := p.factory.getRequestServiceUserID(ctx, xcontext.Configs(ctx).Auth.Twitter.Name)
	if userScreenName == "" {
		return nil, errorx.New(errorx.Unavailable, "User has not connected to twitter")
	}

	isLikeAccepted := true
	if p.Like {
		isLikeAccepted = false

		ok, err := p.factory.twitterEndpoint.CheckLiked(
			ctx, userScreenName, p.originTweet.UserScreenName, p.originTweet.TweetID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get liked tweet: %v", err)
			return nil, errorx.Unknown
		}

		if ok {
			isLikeAccepted = true
		}
	}

	var reply *twitter.Tweet
	var retweet *twitter.Tweet
	if p.Reply || p.Retweet {
		var err error
		reply, retweet, err = p.factory.twitterEndpoint.GetReplyAndRetweet(
			ctx, userScreenName, p.originTweet.UserScreenName, p.originTweet.TweetID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get liked tweet: %v", err)
			return nil, errorx.Unknown
		}
	}

	isRetweetAccepted := true
	if p.Retweet {
		if retweet == nil {
			isRetweetAccepted = false
		}
	}

	isReplyAccepted := true
	if p.Reply {
		if reply == nil {
			isReplyAccepted = false
		}
	}

	if !isLikeAccepted {
		return Rejected.WithMessage("User has not liked the tweet"), nil
	}

	if !isRetweetAccepted {
		return Rejected.WithMessage("User has not retweet the tweet"), nil
	}

	if !isReplyAccepted {
		return Rejected.WithMessage("User has not reply the tweet"), nil
	}

	return Accepted, nil
}

// Twitter Tweet Processor
type twitterTweetProcessor struct {
	IncludedWords []string `mapstructure:"included_words" structs:"included_words"`
	DefaultTweet  string   `mapstructure:"default_tweet" structs:"default_tweet"`

	retryAfter time.Duration
	factory    Factory
}

func newTwitterTweetProcessor(
	ctx context.Context, factory Factory, data map[string]any,
) (*twitterTweetProcessor, error) {
	twitterTweet := twitterTweetProcessor{}
	err := mapstructure.Decode(data, &twitterTweet)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	twitterTweet.retryAfter = xcontext.Configs(ctx).Quest.Twitter.ReclaimDelay
	twitterTweet.factory = factory
	return &twitterTweet, nil
}

func (p twitterTweetProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *twitterTweetProcessor) GetActionForClaim(ctx context.Context, submissionData string) (ActionForClaim, error) {
	tw, err := parseTweetURL(submissionData)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot parse tweet url: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid tweet url")
	}

	userScreenName := p.factory.getRequestServiceUserID(ctx, xcontext.Configs(ctx).Auth.Twitter.Name)
	if userScreenName == "" {
		return nil, errorx.New(errorx.Unavailable, "User has not connected to twitter")
	}

	if tw.UserScreenName != userScreenName {
		return Rejected.WithMessage("The tweet URL is not yours"), nil
	}

	resp, err := p.factory.twitterEndpoint.GetTweet(ctx, tw.UserScreenName, tw.TweetID)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot get tweet: %v", err)
		return Rejected.WithMessage("Not found tweet"), nil
	}

	if resp.Author != tw.UserScreenName {
		return Rejected.WithMessage("The tweet is not yours"), nil
	}

	for _, word := range p.IncludedWords {
		if !strings.Contains(resp.Text, word) {
			return Rejected.WithMessage("The tweet doesn't include \"%s\"", word), nil
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
	ctx context.Context, factory Factory, data map[string]any,
) (*twitterJoinSpaceProcessor, error) {
	twitterJoinSpace := twitterJoinSpaceProcessor{}
	err := mapstructure.Decode(data, &twitterJoinSpace)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	_, err = url.ParseRequestURI(twitterJoinSpace.SpaceURL)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid space url: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid space url")
	}

	twitterJoinSpace.factory = factory
	return &twitterJoinSpace, nil
}

func (p twitterJoinSpaceProcessor) RetryAfter() time.Duration {
	return 0
}

func (p *twitterJoinSpaceProcessor) GetActionForClaim(ctx context.Context, submissionData string) (ActionForClaim, error) {
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
	ctx context.Context,
	factory Factory,
	quest entity.Quest,
	data map[string]any,
	needParse bool,
) (*joinDiscordProcessor, error) {
	joinDiscord := joinDiscordProcessor{}
	err := mapstructure.Decode(data, &joinDiscord)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	if needParse {
		community, err := factory.communityRepo.GetByID(ctx, quest.CommunityID.String)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}

		if community.Discord == "" {
			return nil, errorx.New(errorx.Unavailable, "Community hasn't connected to discord server")
		}

		hasAddBot, err := factory.discordEndpoint.HasAddedBot(ctx, community.Discord)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot call api hasAddedBot: %v", err)
			return nil, errorx.Unknown
		}

		if !hasAddBot {
			return nil, errorx.New(errorx.Unavailable, "Community hasn't added bot to discord server")
		}

		code, err := parseInviteDiscordURL(joinDiscord.InviteLink)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Cannot parse invite link: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid invite link")
		}

		err = factory.discordEndpoint.CheckCode(ctx, community.Discord, code)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Cannot check code: %v", err)
			return nil, errorx.New(errorx.Unavailable,
				"Invite link doesn't belongs to server, or expired, or overused")
		}

		joinDiscord.GuildID = community.Discord
	}

	joinDiscord.retryAfter = xcontext.Configs(ctx).Quest.Dicord.ReclaimDelay
	joinDiscord.factory = factory
	return &joinDiscord, nil
}

func (p joinDiscordProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *joinDiscordProcessor) GetActionForClaim(ctx context.Context, submissionData string) (ActionForClaim, error) {
	userDiscordID := p.factory.getRequestServiceUserID(ctx, xcontext.Configs(ctx).Auth.Discord.Name)
	if userDiscordID == "" {
		return nil, errorx.New(errorx.Unavailable, "User has not connected to discord")
	}

	member, err := p.factory.discordEndpoint.GetMember(ctx, p.GuildID, userDiscordID)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Failed to check member: %v", err)
		return Rejected.WithMessage("Unable to get your information in discord server"), nil
	}

	if member.ID == "" {
		return Rejected.WithMessage("User has not joined in the discord server"), nil
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
	ctx context.Context,
	factory Factory,
	quest entity.Quest,
	data map[string]any,
	needParse bool,
) (*inviteDiscordProcessor, error) {
	inviteDiscord := inviteDiscordProcessor{}
	err := mapstructure.Decode(data, &inviteDiscord)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	if needParse {
		if inviteDiscord.Number <= 0 {
			return nil, errorx.New(errorx.BadRequest, "Invalid number of invites")
		}

		community, err := factory.communityRepo.GetByID(ctx, quest.CommunityID.String)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}

		if community.Discord == "" {
			return nil, errorx.New(errorx.Unavailable, "Community hasn't connected to discord server")
		}

		hasAddBot, err := factory.discordEndpoint.HasAddedBot(ctx, community.Discord)
		if err != nil {
			xcontext.Logger(ctx).Warnf("Cannot call hasAddedBot api: %v", err)
			return nil, errorx.Unknown
		}

		if !hasAddBot {
			return nil, errorx.New(errorx.Unavailable, "Community hasn't added bot to discord server")
		}

		inviteDiscord.GuildID = community.Discord
	}

	inviteDiscord.retryAfter = xcontext.Configs(ctx).Quest.Dicord.ReclaimDelay
	inviteDiscord.factory = factory
	return &inviteDiscord, nil
}

func (p *inviteDiscordProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *inviteDiscordProcessor) GetActionForClaim(
	ctx context.Context, submissionData string,
) (ActionForClaim, error) {
	userDiscordID := p.factory.getRequestServiceUserID(ctx, xcontext.Configs(ctx).Auth.Discord.Name)
	if userDiscordID == "" {
		return nil, errorx.New(errorx.Unavailable, "User has not connected to discord")
	}

	codeString, err := parseInviteDiscordURL(submissionData)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot parse invite discord url: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid submission data")
	}

	inviteCode, err := p.factory.discordEndpoint.GetCode(ctx, p.GuildID, codeString)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Failed to get code: %v", err)
		return Rejected.WithMessage("Unable to find yours code"), nil
	}

	if inviteCode.Inviter.ID != userDiscordID {
		return Rejected.WithMessage("This is not yours code"), nil
	}

	if inviteCode.Uses < p.Number {
		return Rejected.WithMessage(
			"Not enough number of invites (got %d, but expected %d)", inviteCode.Uses, p.Number), nil
	}

	return Accepted, nil
}

// Join Telegram Processor
type joinTelegramProcessor struct {
	GroupLink string `mapstructure:"group_link" structs:"group_link"`

	retryAfter time.Duration
	chatID     string
	factory    Factory
}

func newJoinTelegramProcessor(
	ctx context.Context,
	factory Factory,
	quest entity.Quest,
	data map[string]any,
	needParse bool,
) (*joinTelegramProcessor, error) {
	joinTelegram := joinTelegramProcessor{}

	err := mapstructure.Decode(data, &joinTelegram)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	groupName, err := parseInviteTelegramURL(joinTelegram.GroupLink)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot parse invite telegram link: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid link")
	}

	if groupName == "" {
		return nil, errorx.New(errorx.BadRequest, "Invalid link with an empty group name")
	}

	chatID := "@" + groupName
	if needParse {
		chat, err := factory.telegramEndpoint.GetChat(ctx, chatID)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Cannot get telegram chat: %v", err)
			return nil, errorx.New(errorx.Unavailable,
				"Please use the group link (not invite link) or check the link again")
		}

		if !chat.IsPublic {
			return nil, errorx.New(errorx.Unavailable, "The telegram group must be public")
		}
	}

	joinTelegram.chatID = chatID
	joinTelegram.retryAfter = xcontext.Configs(ctx).Quest.Telegram.ReclaimDelay
	joinTelegram.factory = factory

	return &joinTelegram, nil
}

func (p joinTelegramProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *joinTelegramProcessor) GetActionForClaim(ctx context.Context, submissionData string) (ActionForClaim, error) {
	telegramUserID := p.factory.getRequestServiceUserID(ctx, xcontext.Configs(ctx).Auth.Telegram.Name)
	if telegramUserID == "" {
		return nil, errorx.New(errorx.Unavailable, "User has not connected telegram")
	}

	_, err := p.factory.telegramEndpoint.GetMember(ctx, p.chatID, telegramUserID)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot get member: %v", err)
		return Rejected.WithMessage("User has not joined in telegram group"), nil
	}

	return Accepted, nil
}

// Invite Processor
type inviteProcessor struct {
	Number int `mapstructure:"number" structs:"number"`

	retryAfter  time.Duration
	communityID string
	factory     Factory
}

func newInviteProcessor(
	ctx context.Context,
	factory Factory,
	quest entity.Quest,
	data map[string]any,
	needParse bool,
) (*inviteProcessor, error) {
	invite := inviteProcessor{}
	err := mapstructure.Decode(data, &invite)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	if needParse {
		if invite.Number <= 0 {
			return nil, errorx.New(errorx.BadRequest, "Number of invites must be positive")
		}
	}

	invite.retryAfter = xcontext.Configs(ctx).Quest.InviteReclaimDelay
	invite.communityID = quest.CommunityID.String
	invite.factory = factory

	return &invite, nil
}

func (p inviteProcessor) RetryAfter() time.Duration {
	return p.retryAfter
}

func (p *inviteProcessor) GetActionForClaim(ctx context.Context, submissionData string) (ActionForClaim, error) {
	follower, err := p.factory.followerRepo.Get(ctx, xcontext.RequestUserID(ctx), p.communityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get follower: %v", err)
		return nil, errorx.Unknown
	}

	if follower.InviteCount < uint64(p.Number) {
		return Rejected.WithMessage(
			"Not enough number of invites (got %d, but expected %d)", follower.InviteCount, p.Number), nil
	}

	return Accepted, nil
}
