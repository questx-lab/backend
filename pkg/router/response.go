package router

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type response struct {
	Code  int64  `json:"code"`
	Error string `json:"error,omitempty"`
	Data  any    `json:"data"`
}

func newResponse(data any) response {
	return response{
		Code: 0,
		Data: data,
	}
}

func newErrorResponse(err error) response {
	errx := errorx.Error{}
	if errors.As(err, &errx) {
		return response{
			Code:  int64(errx.Code),
			Error: errx.Message,
		}
	}

	return response{
		Code:  int64(errorx.Unknown.Code),
		Error: errorx.Unknown.Message,
	}
}

func handleResponse() CloserFunc {
	return func(ctx context.Context) {
		err := func() error {
			if err := xcontext.Error(ctx); err != nil {
				return err
			}

			if resp := xcontext.Response(ctx); resp != nil {
				if err := writeJSON(xcontext.HTTPWriter(ctx), newResponse(resp)); err != nil {
					xcontext.Logger(ctx).Errorf("Cannot write the response %v", err)
					return errorx.New(errorx.BadResponse, "Cannot write the response")
				}
			}

			return nil
		}()

		if err != nil {
			resp := newErrorResponse(err)
			if wsclient := xcontext.WSClient(ctx); wsclient != nil {
				if err := wsWriteJSON(wsclient, resp); err != nil {
					xcontext.Logger(ctx).Errorf("cannot write the response: %s", err.Error())
					wsclient.Conn.Close()
				}
			} else {
				if err := writeJSON(xcontext.HTTPWriter(ctx), resp); err != nil {
					xcontext.Logger(ctx).Errorf("cannot write the response: %s", err.Error())
				}
			}
		}
	}
}

func writeJSON(r http.ResponseWriter, resp any) error {
	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	if _, err := r.Write(b); err != nil {
		return err
	}

	return nil
}

func wsWriteJSON(wsClient *ws.Client, resp any) error {
	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	data := websocket.FormatCloseMessage(websocket.CloseNormalClosure, string(b))
	if err := wsClient.Conn.WriteMessage(websocket.CloseMessage, data); err != nil {
		return err
	}

	return nil
}
