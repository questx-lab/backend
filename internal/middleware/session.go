package middleware

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type SessionResponse interface {
	SessionInfo() map[string]any
}

func HandleSaveSession() router.MiddlewareFunc {
	return func(ctx context.Context) (context.Context, error) {
		sessionResp, ok := xcontext.Response(ctx).(SessionResponse)
		if !ok {
			return nil, nil
		}

		sessionInfo := sessionResp.SessionInfo()
		if sessionInfo == nil {
			return nil, errors.New("no session info")
		}

		req := xcontext.HTTPRequest(ctx)
		cfg := xcontext.Configs(ctx)
		session, err := xcontext.SessionStore(ctx).Get(req, cfg.Session.Name)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot decode the current existing session: %v", err)

			session, err = xcontext.SessionStore(ctx).New(req, cfg.Session.Name)
			if err != nil {
				return nil, err
			}
		}

		for k, v := range sessionInfo {
			session.Values[k] = v
		}

		if err := session.Save(req, xcontext.HTTPWriter(ctx)); err != nil {
			return nil, err
		}

		return nil, nil
	}
}
