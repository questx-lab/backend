package domain

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/mocks"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/stretchr/testify/require"
	"github.com/test-go/testify/mock"
)

func Test_userDomain_GetReferralInfo(t *testing.T) {
	ctx := testutil.NewMockContext()
	testutil.CreateFixtureDb(ctx)

	domain := NewUserDomain(
		repository.NewUserRepository(),
		repository.NewOAuth2Repository(),
		repository.NewParticipantRepository(),
		repository.NewBadgeRepository(),
		badge.NewManager(
			repository.NewBadgeRepository(),
			&testutil.MockBadge{
				NameValue:     badge.SharpScoutBadgeName,
				IsGlobalValue: false,
				ScanFunc: func(ctx xcontext.Context, userID, projectID string) (int, error) {
					return 0, nil
				},
			},
		),
		nil,
	)

	inviteResp, err := domain.GetInvite(ctx, &model.GetInviteRequest{
		InviteCode: testutil.Participant1.InviteCode,
	})
	require.NoError(t, err)
	require.Equal(t, inviteResp.Project.ID, testutil.Project1.ID)
	require.Equal(t, inviteResp.Project.Name, testutil.Project1.Name)
}

func Test_userDomain_FollowProject_and_GetBadges(t *testing.T) {
	ctx := testutil.NewMockContext()
	testutil.CreateFixtureDb(ctx)

	userRepo := repository.NewUserRepository()
	oauth2Repo := repository.NewOAuth2Repository()
	pariticipantRepo := repository.NewParticipantRepository()
	badgeRepo := repository.NewBadgeRepository()

	newUser := &entity.User{Base: entity.Base{ID: uuid.NewString()}}
	require.NoError(t, userRepo.Create(ctx, newUser))

	domain := NewUserDomain(
		userRepo,
		oauth2Repo,
		pariticipantRepo,
		badgeRepo,
		badge.NewManager(
			badgeRepo,
			&testutil.MockBadge{
				NameValue:     badge.SharpScoutBadgeName,
				IsGlobalValue: false,
				ScanFunc: func(ctx xcontext.Context, userID, projectID string) (int, error) {
					return 1, nil
				},
			},
		),
		nil,
	)

	ctx = testutil.NewMockContextWithUserID(ctx, newUser.ID)
	_, err := domain.FollowProject(ctx, &model.FollowProjectRequest{
		ProjectID: testutil.Participant1.ProjectID,
		InvitedBy: testutil.Participant1.UserID,
	})
	require.NoError(t, err)

	// Get badges and check their level, name. Ensure that they haven't been
	// notified to client yet.
	badges, err := domain.GetBadges(ctx, &model.GetBadgesRequest{
		UserID:    testutil.Participant1.UserID,
		ProjectID: testutil.Participant1.ProjectID,
	})
	require.NoError(t, err)
	require.Len(t, badges.Badges, 1)
	require.Equal(t, badge.SharpScoutBadgeName, badges.Badges[0].Name)
	require.Equal(t, 1, badges.Badges[0].Level)
	require.Equal(t, false, badges.Badges[0].WasNotified)

	// Get badges again and ensure they was notified to client.
	badges, err = domain.GetBadges(ctx, &model.GetBadgesRequest{
		UserID:    testutil.Participant1.UserID,
		ProjectID: testutil.Participant1.ProjectID,
	})
	require.NoError(t, err)
	require.Len(t, badges.Badges, 1)
	require.Equal(t, true, badges.Badges[0].WasNotified)
}

func Test_userDomain_UploadAvatar(t *testing.T) {
	path := "out.png"
	generateRandomImage(path)
	defer deleteImage(path)
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()
	fw, err := writer.CreateFormFile("avatar", file.Name())
	require.NoError(t, err)

	_, err = io.Copy(fw, file)
	require.NoError(t, err)
	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/testAvatar", body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	ctx := testutil.NewMockContextWith(request)
	ctx = testutil.NewMockContextWithUserID(ctx, testutil.User1.ID)
	testutil.CreateFixtureDb(ctx)

	userRepo := repository.NewUserRepository()
	mockedStorage := &mocks.Storage{}
	domain := NewUserDomain(userRepo, nil, nil, nil, nil, mockedStorage)

	mockedStorage.On("BulkUpload", mock.Anything, mock.Anything).Return([]*storage.UploadResponse{
		{Url: "28x28.png"},
		{Url: "56x56.png"},
		{Url: "128x128.png"},
	}, nil)
	_, err = domain.UploadAvatar(ctx, &model.UploadAvatarRequest{})
	require.NoError(t, err)

	user, err := userRepo.GetByID(ctx, testutil.User1.ID)
	require.NoError(t, err)
	require.Len(t, user.ProfilePictures, 3)
	require.Equal(t, user.ProfilePictures[common.AvatarSizes[0].String()],
		map[string]any{"filename": "", "url": "28x28.png"})
	require.Equal(t, user.ProfilePictures[common.AvatarSizes[1].String()],
		map[string]any{"filename": "", "url": "56x56.png"})
	require.Equal(t, user.ProfilePictures[common.AvatarSizes[2].String()],
		map[string]any{"filename": "", "url": "128x128.png"})
}

func generateRandomImage(path string) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	img.Set(2, 3, color.RGBA{255, 0, 0, 255})
	f, _ := os.Create(path)
	defer f.Close()
	_ = png.Encode(f, img)
}

func deleteImage(path string) {
	_ = os.Remove(path)
}
