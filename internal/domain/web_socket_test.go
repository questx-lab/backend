package domain

import (
	"net/http/httptest"
	"testing"

	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"

	"github.com/stretchr/testify/require"
)

func Test_WebSocket(t *testing.T) {
	verifier := middleware.NewAuthVerifier().WithAccessToken()
	roomRepo := repository.NewRoomRepository()

	domain := NewWsDomain(roomRepo, verifier)
	go domain.Run()
	ctx := testutil.NewMockContext()
	ctx = testutil.NewMockContextWithUserID(ctx, testutil.User1.ID)
	request := httptest.NewRequest("GET", "/testWebsocket", nil)
	request.Header.Add("Connection", "Upgrade")
	request.Header.Add("Upgrade", "websocket")
	request.Header.Add("Sec-WebSocket-Version", "13")
	request.Header.Add("Sec-WebSocket-Key", "x3JJHMbDL1EzLkh9GBhXDw==")
	// request.Header.Add("Sec-WebSocket-Version", "13")

	ctx.SetRequest(request)
	response := testutil.NewRecorder(nil)

	ctx.SetWriter(response)
	err := domain.Serve(ctx)

	require.NoError(t, err)
}
