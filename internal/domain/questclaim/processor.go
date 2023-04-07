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

// Twitter Processor
type twitterFollow struct {
	AccountURL string `mapstructure:"account_url," json:"account_url,omitempty"`
}

type twitterLike struct {
	TweetURL string `mapstructure:"tweet_url" json:"tweet_url,omitempty"`
}

type twitterReply struct {
	TweetURL     string `mapstructure:"tweet_url" json:"tweet_url,omitempty"`
	DefaultReply string `mapstructure:"default_reply" json:"default_reply,omitempty"`
}

type twitterRetweet struct {
	TweetURL string `mapstructure:"tweet_url" json:"tweet_url,omitempty"`
}

type twitterTweet struct {
	IncluedWords []string `mapstructure:"inclued_words" json:"inclued_words,omitempty"`
	DefaultTweet string   `mapstructure:"default_tweet" json:"default_tweet,omitempty"`
}

type twitterJoinSpace struct {
	SpaceURL string `mapstructure:"space_url" json:"space_url,omitempty"`
}

type twitterProcessor struct {
	Follow    *twitterFollow    `mapstructure:"follow" json:"follow,omitempty"`
	Like      *twitterLike      `mapstructure:"like" json:"like,omitempty"`
	Reply     *twitterReply     `mapstructure:"reply" json:"reply,omitempty"`
	Retweet   *twitterRetweet   `mapstructure:"retweet" json:"retweet,omitempty"`
	Tweet     *twitterTweet     `mapstructure:"tweet" json:"tweet,omitempty"`
	JoinSpace *twitterJoinSpace `mapstructure:"join_space," json:"join_space,omitempty"`
}

func newTwitterProcessor(ctx xcontext.Context, data map[string]any) (*twitterProcessor, error) {
	twitter := twitterProcessor{}
	err := mapstructure.Decode(data, &twitter)
	if err != nil {
		return nil, err
	}

	// TODO: Also need to check if these following URL existed or not.

	if twitter.Follow != nil {
		_, err = url.ParseRequestURI(twitter.Follow.AccountURL)
		if err != nil {
			return nil, err
		}
	}

	if twitter.Like != nil {
		_, err = url.ParseRequestURI(twitter.Like.TweetURL)
		if err != nil {
			return nil, err
		}
	}

	if twitter.Reply != nil {
		_, err = url.ParseRequestURI(twitter.Reply.TweetURL)
		if err != nil {
			return nil, err
		}
	}

	if twitter.Retweet != nil {
		_, err = url.ParseRequestURI(twitter.Retweet.TweetURL)
		if err != nil {
			return nil, err
		}
	}

	if twitter.JoinSpace != nil {
		_, err = url.ParseRequestURI(twitter.JoinSpace.SpaceURL)
		if err != nil {
			return nil, err
		}
	}

	return &twitter, nil
}

func (p *twitterProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	return NeedManualReview, nil
}
