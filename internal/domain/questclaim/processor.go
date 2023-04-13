package questclaim

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

// VisitLink Processor
type visitLinkProcessor struct {
	Link string `mapstructure:"link" json:"link,omitempty"`
}

func newVisitLinkProcessor(ctx xcontext.Context, data map[string]any) (*visitLinkProcessor, error) {
	visitLink := visitLinkProcessor{}
	err := mapstructure.Decode(data, &visitLink)
	if err != nil {
		return nil, err
	}

	if visitLink.Link == "" {
		return nil, errors.New("Not found link in validation data")
	}

	_, err = url.ParseRequestURI(visitLink.Link)
	if err != nil {
		return nil, err
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
	AutoValidate bool   `mapstructure:"auto_validate" json:"auto_validate,omitempty"`
	Answer       string `mapstructure:"answer" json:"answer,omitempty"`
}

func newTextProcessor(ctx xcontext.Context, data map[string]any) (*textProcessor, error) {
	text := textProcessor{}
	err := mapstructure.Decode(data, &text)
	if err != nil {
		return nil, err
	}

	return &text, nil
}

func (v *textProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
	if !v.AutoValidate {
		return NeedManualReview, nil
	}

	if v.Answer != input {
		return Rejected, nil
	}

	return Accepted, nil
}

// Twitter Follow Processor
type twitterFollowProcessor struct {
	TwitterHandle string `mapstructure:"twitter_handle" json:"twitter_handle,omitempty"`

	user     TwitterUser
	endpoint twitter.IEndpoint
}

func newTwitterFollowProcessor(
	ctx xcontext.Context, endpoint twitter.IEndpoint, data map[string]any,
) (*twitterFollowProcessor, error) {
	twitterFollow := twitterFollowProcessor{}
	err := mapstructure.Decode(data, &twitterFollow)
	if err != nil {
		return nil, err
	}

	user, err := parseTwitterUserURL(twitterFollow.TwitterHandle)
	if err != nil {
		return nil, err
	}

	twitterFollow.user = user
	twitterFollow.endpoint = endpoint
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

	b, err := p.endpoint.CheckFollowing(ctx, p.user.UserScreenName)
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
	Like    bool `mapstructure:"like" json:"like,omitempty"`
	Retweet bool `mapstructure:"retweet" json:"retweet,omitempty"`
	Reply   bool `mapstructure:"reply" json:"reply,omitempty"`

	TweetURL     string `mapstructure:"tweet_url" json:"tweet_url,omitempty"`
	DefaultReply string `mapstructure:"default_reply" json:"default_reply,omitempty"`

	originTweet Tweet
	endpoint    twitter.IEndpoint
}

func newTwitterReactionProcessor(
	ctx xcontext.Context, endpoint twitter.IEndpoint, data map[string]any,
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

	remoteTweet, err := endpoint.GetTweet(ctx, tweet.TweetID)
	if err != nil {
		return nil, err
	}

	if remoteTweet.AuthorScreenName != tweet.UserScreenName {
		return nil, errors.New("invalid user")
	}

	twitterReaction.originTweet = tweet
	twitterReaction.endpoint = endpoint
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

	isLikeAccepted := true
	if p.Like {
		isLikeAccepted = false

		tweets, err := p.endpoint.GetLikedTweet(ctx)
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

		retweets, err := p.endpoint.GetRetweet(ctx, p.originTweet.TweetID)
		if err != nil {
			ctx.Logger().Errorf("Cannot get retweet: %v", err)
			return Rejected, errorx.Unknown
		}

		for _, retweet := range retweets {
			if retweet.AuthorScreenName == p.endpoint.OnBehalf() {
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

		if replyTweet.UserScreenName == p.endpoint.OnBehalf() {
			_, err := p.endpoint.GetTweet(ctx, replyTweet.TweetID)
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
	IncludedWords []string `mapstructure:"included_words" json:"included_words,omitempty"`
	DefaultTweet  string   `mapstructure:"default_tweet" json:"default_tweet,omitempty"`

	endpoint twitter.IEndpoint
}

func newTwitterTweetProcessor(
	ctx xcontext.Context, endpoint twitter.IEndpoint, data map[string]any,
) (*twitterTweetProcessor, error) {
	twitterTweet := twitterTweetProcessor{}
	err := mapstructure.Decode(data, &twitterTweet)
	if err != nil {
		return nil, err
	}

	twitterTweet.endpoint = endpoint
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

	if tw.UserScreenName != p.endpoint.OnBehalf() {
		return Rejected, nil
	}

	resp, err := p.endpoint.GetTweet(ctx, tw.TweetID)
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
	SpaceURL string `mapstructure:"space_url" json:"space_url,omitempty"`

	endpoint twitter.IEndpoint
}

func newTwitterJoinSpaceProcessor(
	ctx xcontext.Context, endpoint twitter.IEndpoint, data map[string]any,
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

	twitterJoinSpace.endpoint = endpoint
	return &twitterJoinSpace, nil
}

func (p *twitterJoinSpaceProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
	return NeedManualReview, nil
}

// Join Discord Processor
type joinDiscordProcessor struct {
	InviteLink string `mapstructure:"invite_link" json:"invite_link,omitempty"`

	guildID  string
	endpoint discord.IEndpoint
}

func newJoinDiscordProcessor(
	ctx xcontext.Context, endpoint discord.IEndpoint, data map[string]any,
) (*joinDiscordProcessor, error) {
	joinDiscord := joinDiscordProcessor{}
	err := mapstructure.Decode(data, &joinDiscord)
	if err != nil {
		return nil, err
	}

	code, err := getDiscordInviteCode(joinDiscord.InviteLink)
	if err != nil {
		return nil, err
	}

	guild, err := endpoint.GetGuildFromCode(ctx, code)
	if err != nil {
		return nil, err
	}

	hasAddBot, err := endpoint.HasAddedBot(ctx, guild.ID)
	if err != nil {
		return nil, err
	}

	if !hasAddBot {
		return nil, errors.New("server has not added bot yet")
	}

	joinDiscord.guildID = guild.ID
	joinDiscord.endpoint = endpoint
	return &joinDiscord, nil
}

func (p *joinDiscordProcessor) GetActionForClaim(
	ctx xcontext.Context, lastClaimed *entity.ClaimedQuest, input string,
) (ActionForClaim, error) {
	isJoined, err := p.endpoint.CheckMember(ctx, p.guildID)
	if err != nil {
		ctx.Logger().Debugf("Failed to check member: %v", err)
		return Rejected, nil
	}

	if !isJoined {
		return Rejected, nil
	}

	return Accepted, nil
}
