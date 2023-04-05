package middleware

import (
	"errors"

	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type SessionResponse interface {
	SessionInfo() map[string]any
}

func HandleSaveSession() router.MiddlewareFunc {
	return func(ctx xcontext.Context) error {
		sessionResp, ok := xcontext.GetResponse(ctx).(SessionResponse)
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
