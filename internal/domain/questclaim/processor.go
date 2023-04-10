package questclaim

import (
	"errors"
	"net/url"

	"github.com/mitchellh/mapstructure"
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

func (v *visitLinkProcessor) GetActionForClaim(xcontext.Context, string) (ActionForClaim, error) {
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

func (v *textProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
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
	AccountURL string `mapstructure:"account_url" json:"account_url,omitempty"`
}

func newTwitterFollowProcessor(ctx xcontext.Context, data map[string]any) (*twitterFollowProcessor, error) {
	twitterFollow := twitterFollowProcessor{}
	err := mapstructure.Decode(data, &twitterFollow)
	if err != nil {
		return nil, err
	}

	_, err = url.ParseRequestURI(twitterFollow.AccountURL)
	if err != nil {
		return nil, err
	}

	return &twitterFollow, nil
}

func (p *twitterFollowProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	return NeedManualReview, nil
}

// Twitter Reaction Processsor
type twitterReactionProcessor struct {
	Like         bool   `mapstructure:"like" json:"like,omitempty"`
	Retweet      bool   `mapstructure:"retweet" json:"retweet,omitempty"`
	Reply        bool   `mapstructure:"reply" json:"reply,omitempty"`
	TweetURL     string `mapstructure:"tweet_url" json:"tweet_url,omitempty"`
	DefaultReply string `mapstructure:"default_reply" json:"default_reply,omitempty"`
}

func newTwitterReactionProcessor(ctx xcontext.Context, data map[string]any) (*twitterReactionProcessor, error) {
	twitterReaction := twitterReactionProcessor{}
	err := mapstructure.Decode(data, &twitterReaction)
	if err != nil {
		return nil, err
	}

	_, err = url.ParseRequestURI(twitterReaction.TweetURL)
	if err != nil {
		return nil, err
	}

	return &twitterReaction, nil
}

func (p *twitterReactionProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	return NeedManualReview, nil
}

// Twitter Tweet Processor
type twitterTweetProcessor struct {
	IncludedWords []string `mapstructure:"included_words" json:"included_words,omitempty"`
	DefaultTweet  string   `mapstructure:"default_tweet" json:"default_tweet,omitempty"`
}

func newTwitterTweetProcessor(ctx xcontext.Context, data map[string]any) (*twitterTweetProcessor, error) {
	twitterTweet := twitterTweetProcessor{}
	err := mapstructure.Decode(data, &twitterTweet)
	if err != nil {
		return nil, err
	}

	return &twitterTweet, nil
}

func (p *twitterTweetProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	return NeedManualReview, nil
}

// Twitter Join Space Processsor
type twitterJoinSpaceProcessor struct {
	SpaceURL string `mapstructure:"space_url" json:"space_url,omitempty"`
}

func newTwitterJoinSpaceProcessor(ctx xcontext.Context, data map[string]any) (*twitterJoinSpaceProcessor, error) {
	twitterJoinSpace := twitterJoinSpaceProcessor{}
	err := mapstructure.Decode(data, &twitterJoinSpace)
	if err != nil {
		return nil, err
	}

	_, err = url.ParseRequestURI(twitterJoinSpace.SpaceURL)
	if err != nil {
		return nil, err
	}

	return &twitterJoinSpace, nil
}

func (p *twitterJoinSpaceProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	return NeedManualReview, nil
}
