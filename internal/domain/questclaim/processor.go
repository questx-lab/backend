package questclaim

import (
	"errors"
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

func (p *urlProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
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

func (v *visitLinkProcessor) GetActionForClaim(
	xcontext.Context, *entity.ClaimedQuest, string,
) (ActionForClaim, error) {
	return Accepted, nil
}

// Text Processor
// TODO: Add retry_after when the claimed quest is rejected by auto validate.
type textProcessor struct {
	AutoValidate bool   `mapstructure:"auto_validate" structs:"auto_validate"`
	Answer       string `mapstructure:"answer" structs:"answer"`
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

func (p *textProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
	if !p.AutoValidate {
		return NeedManualReview, nil
	}

	if p.Answer != input {
		return Rejected, nil
	}

	return Accepted, nil
}

// Quiz Processor
type quizProcessor struct {
	Question string   `mapstructure:"question" structs:"question"`
	Options  []string `mapstructure:"options" structs:"options"`
	Answer   string   `mapstructure:"answer" structs:"answer"`
}

func newQuizProcessor(ctx xcontext.Context, data map[string]any, needParse bool) (*quizProcessor, error) {
	quiz := quizProcessor{}
	err := mapstructure.Decode(data, &quiz)
	if err != nil {
		return nil, err
	}

	if needParse {
		if len(quiz.Options) < 2 {
			return nil, errors.New("provide at least two options")
		}

		ok := false
		for _, option := range quiz.Options {
			if quiz.Answer == option {
				ok = true
				break
			}
		}

		if !ok {
			return nil, errors.New("not found the answer in options")
		}
	}

	return &quiz, nil
}

func (p *quizProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
	if input == p.Answer {
		return Accepted, nil
	}

	return Rejected, nil
}

// Image Processor
type imageProcessor struct{}

func newImageProcessor(xcontext.Context, map[string]any) (*imageProcessor, error) {
	return &imageProcessor{}, nil
}

func (p *imageProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
	// TODO: Input is a link of image, need to validate the image.
	return NeedManualReview, nil
}

// Empty Processor
type emptyProcessor struct{}

func newEmptyProcessor(xcontext.Context, map[string]any) (*emptyProcessor, error) {
	return &emptyProcessor{}, nil
}

func (p *emptyProcessor) GetActionForClaim(xcontext.Context, *entity.ClaimedQuest, string) (ActionForClaim, error) {
	return Accepted, nil
}

// Twitter Follow Processor
type twitterFollowProcessor struct {
	TwitterHandle string `mapstructure:"twitter_handle" structs:"twitter_handle"`

	target  twitterUser
	factory Factory
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

	twitterFollow.target = target
	twitterFollow.factory = factory
	return &twitterFollow, nil
}

func (p *twitterFollowProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
	// Check time for reclaiming.
	if lastClaimed != nil {
		if elapsed := time.Since(lastClaimed.CreatedAt); elapsed <= ctx.Configs().Quest.Twitter.ReclaimDelay {
			waitFor := ctx.Configs().Quest.Twitter.ReclaimDelay - elapsed
			return Rejected, errorx.New(errorx.TooManyRequests,
				"You need to wait for %s to reclaim this quest", waitFor)
		}
	}

	userScreenName := p.factory.getRequestUserServiceID(ctx, ctx.Configs().Auth.Twitter.Name)
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

	twitterReaction.originTweet = tweet
	twitterReaction.factory = factory
	return &twitterReaction, nil
}

func (p *twitterReactionProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
	// Check time for reclaiming.
	if lastClaimed != nil {
		if elapsed := time.Since(lastClaimed.CreatedAt); elapsed <= ctx.Configs().Quest.Twitter.ReclaimDelay {
			waitFor := ctx.Configs().Quest.Twitter.ReclaimDelay - elapsed
			return Rejected, errorx.New(errorx.TooManyRequests,
				"You need to wait for %s to reclaim this quest", waitFor)
		}
	}

	userScreenName := p.factory.getRequestUserServiceID(ctx, ctx.Configs().Auth.Twitter.Name)
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

	factory Factory
}

func newTwitterTweetProcessor(
	ctx xcontext.Context, factory Factory, data map[string]any,
) (*twitterTweetProcessor, error) {
	twitterTweet := twitterTweetProcessor{}
	err := mapstructure.Decode(data, &twitterTweet)
	if err != nil {
		return nil, err
	}

	twitterTweet.factory = factory
	return &twitterTweet, nil
}

func (p *twitterTweetProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
	// Check time for reclaiming.
	if lastClaimed != nil {
		if elapsed := time.Since(lastClaimed.CreatedAt); elapsed <= ctx.Configs().Quest.Twitter.ReclaimDelay {
			waitFor := ctx.Configs().Quest.Twitter.ReclaimDelay - elapsed
			return Rejected, errorx.New(errorx.TooManyRequests,
				"You need to wait for %s to reclaim this quest", waitFor)
		}
	}

	tw, err := parseTweetURL(input)
	if err != nil {
		ctx.Logger().Debugf("Cannot parse tweet url: %v", err)
		return Rejected, errorx.New(errorx.BadRequest, "Invalid tweet url")
	}

	userScreenName := p.factory.getRequestUserServiceID(ctx, ctx.Configs().Auth.Twitter.Name)
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

func (p *twitterJoinSpaceProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
	return NeedManualReview, nil
}

// Join Discord Processor
type joinDiscordProcessor struct {
	Code    string `mapstructure:"code" structs:"code"`
	GuildID string `mapstructure:"guild_id" structs:"guild_id"`

	factory Factory
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
		project, err := factory.projectRepo.GetByID(ctx, quest.ProjectID)
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

		err = factory.discordEndpoint.CheckCode(ctx, project.Discord, joinDiscord.Code)
		if err != nil {
			return nil, err
		}

		joinDiscord.GuildID = project.Discord
	}

	joinDiscord.factory = factory

	return &joinDiscord, nil
}

func (p *joinDiscordProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
	requestUserDiscordID := p.factory.getRequestUserServiceID(ctx, ctx.Configs().Auth.Discord.Name)
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

// Join Telegram Processor
type joinTelegramProcessor struct {
	InviteLink string `mapstructure:"invite_link" structs:"invite_link"`

	chatID  string
	factory Factory
}

func newJoinTelegramProcessor(
	ctx xcontext.Context,
	factory Factory,
	quest entity.Quest,
	data map[string]any,
	needParse bool,
) (*joinTelegramProcessor, error) {
	joinTelegram := joinTelegramProcessor{factory: factory}
	err := mapstructure.Decode(data, &joinTelegram)
	if err != nil {
		return nil, err
	}

	groupName, err := parseInviteTelegramURL(joinTelegram.InviteLink)
	if err != nil {
		return nil, err
	}
	joinTelegram.chatID = "@" + groupName

	if needParse {
		if joinTelegram.chatID == "" {
			return nil, errors.New("got an empty chat id")
		}

		if err := factory.projectRoleVerifier.Verify(ctx, quest.ProjectID, entity.AdminGroup...); err != nil {
			return nil, err
		}

		requestUserID := factory.getRequestUserServiceID(ctx, ctx.Configs().Auth.Telegram.Name)
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

	return &joinTelegram, nil
}

func (p *joinTelegramProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
	requestUserID := p.factory.getRequestUserServiceID(ctx, ctx.Configs().Auth.Telegram.Name)
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

	projectID string
	factory   Factory
}

func newInviteProcessor(
	ctx xcontext.Context,
	factory Factory,
	quest entity.Quest,
	data map[string]any,
	needParse bool,
) (*inviteProcessor, error) {
	invite := inviteProcessor{factory: factory, projectID: quest.ProjectID}
	err := mapstructure.Decode(data, &invite)
	if err != nil {
		return nil, err
	}

	if needParse {
		if invite.Number <= 0 {
			return nil, errors.New("number must be positive")
		}
	}

	return &invite, nil
}

func (p *inviteProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
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
