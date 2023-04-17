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
	ctx.SetRequest(request)
	response := httptest.NewRecorder()
	ctx.SetWriter(response)
	err := domain.Serve(ctx)

	require.NoError(t, err)
}
