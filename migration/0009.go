package migration

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func migrate0009(ctx context.Context) error {
	var quests []entity.Quest
	err := xcontext.DB(ctx).Where("type=? AND community_id IS NOT NULL", entity.QuestTwitterFollow).Find(&quests).Error
	if err != nil {
		return err
	}

	twitterEndpoint := twitter.New(xcontext.Configs(ctx).Quest.Twitter)
	for _, q := range quests {
		if screenName := q.ValidationData["twitter_screen_name"]; screenName == "" || screenName == nil {
			handle, ok := q.ValidationData["twitter_handle"].(string)
			if !ok {
				return fmt.Errorf("invalid twitter handle of %s", q.ID)
			}

			user, err := extractUserFromHandle(ctx, twitterEndpoint, handle)
			if err != nil {
				xcontext.Logger(ctx).Warnf("Cannot extract user from handle %s: %v", handle, err)
				xcontext.Logger(ctx).Warnf("Fall back to https://twitter.com/elonmusk", handle, err)
				user, err = extractUserFromHandle(ctx, twitterEndpoint, "https://twitter.com/elonmusk")
				if err != nil {
					return err
				}

				q.ValidationData["twitter_handle"] = "https://twitter.com/elonmusk"
			}

			q.ValidationData["twitter_screen_name"] = user.ScreenName
			q.ValidationData["twitter_photo_url"] = user.PhotoURL
			q.ValidationData["twitter_name"] = user.Name

			if err = xcontext.DB(ctx).Updates(&q).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func extractUserFromHandle(ctx context.Context, endpoint twitter.IEndpoint, handle string) (twitter.User, error) {
	u, err := url.ParseRequestURI(handle)
	if err != nil {
		return twitter.User{}, err
	}

	if u.Scheme != "https" {
		return twitter.User{}, fmt.Errorf("invalid scheme of %s", q.ID)
	}

	if u.Host != "twitter.com" {
		return twitter.User{}, fmt.Errorf("invalid domain of %s", q.ID)
	}

	path := strings.TrimLeft(u.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 1 {
		return twitter.User{}, fmt.Errorf("invalid path of %s", q.ID)
	}

	user, err := endpoint.GetUser(ctx, parts[0])
	if err != nil {
		return twitter.User{}, err
	}

	return user, nil
}
