package middleware

import (
	"errors"

	"github.com/questx-lab/backend/pkg/router"
)

type SessionResponse interface {
	SessionInfo() map[string]any
}

func HandleSession() router.MiddlewareFunc {
	return func(ctx router.Context) error {
		sessionResp, ok := ctx.GetResponse().(SessionResponse)
		if !ok {
			return nil
		}

		sessionInfo := sessionResp.SessionInfo()
		if sessionInfo == nil {
			return errors.New("no session info")
		}

		session, err := ctx.SessionStore().Get(ctx.Request(), ctx.Configs().Session.Name)
		if err != nil {
			return err
		}

		for k, v := range sessionInfo {
			session.Values[k] = v
		}

		if err := session.Save(ctx.Request(), ctx.Writer()); err != nil {
			return err
		}

		return nil
	}
}
