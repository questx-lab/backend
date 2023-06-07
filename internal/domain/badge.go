package domain

import (
	"context"
	"errors"
	"net/url"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type BadgeDomain interface {
	GetAllBadgeNames(context.Context, *model.GetAllBadgeNamesRequest) (*model.GetAllBadgeNamesResponse, error)
	GetAllBadges(context.Context, *model.GetAllBadgesRequest) (*model.GetAllBadgesResponse, error)
	UpdateBadge(context.Context, *model.UpdateBadgeRequest) (*model.UpdateBadgeResponse, error)
	GetUserBadgeDetails(context.Context, *model.GetUserBadgeDetailsRequest) (*model.GetUserBadgeDetailsResponse, error)
	GetMyBadgeDetails(context.Context, *model.GetMyBadgeDetailsRequest) (*model.GetMyBadgeDetailsResponse, error)
}

type badgeDomain struct {
	badgeRepo       repository.BadgeRepository
	badgeDetailRepo repository.BadgeDetailRepository
	communityRepo   repository.CommunityRepository
	badgeManager    *badge.Manager
}

func NewBadgeDomain(
	badgeRepo repository.BadgeRepository,
	badgeDetailRepo repository.BadgeDetailRepository,
	communityRepo repository.CommunityRepository,
	badgeManager *badge.Manager,
) *badgeDomain {
	return &badgeDomain{
		badgeRepo:       badgeRepo,
		badgeDetailRepo: badgeDetailRepo,
		communityRepo:   communityRepo,
		badgeManager:    badgeManager,
	}
}

func (d *badgeDomain) GetAllBadgeNames(
	ctx context.Context, req *model.GetAllBadgeNamesRequest,
) (*model.GetAllBadgeNamesResponse, error) {
	return &model.GetAllBadgeNamesResponse{Names: d.badgeManager.GetAllBadgeNames()}, nil
}

func (d *badgeDomain) GetAllBadges(
	ctx context.Context, req *model.GetAllBadgesRequest,
) (*model.GetAllBadgesResponse, error) {
	badges, err := d.badgeRepo.GetAll(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all badges: %v", err)
		return nil, errorx.Unknown
	}

	clientBadges := []model.Badge{}
	for _, b := range badges {
		clientBadges = append(clientBadges, convertBadge(&b))
	}

	return &model.GetAllBadgesResponse{Badges: clientBadges}, nil
}

func (d *badgeDomain) UpdateBadge(
	ctx context.Context, req *model.UpdateBadgeRequest,
) (*model.UpdateBadgeResponse, error) {
	if !slices.Contains(d.badgeManager.GetAllBadgeNames(), req.Name) {
		return nil, errorx.New(errorx.Unavailable, "Badge %s is unavailable", req.Name)
	}

	if req.Level <= 0 {
		return nil, errorx.New(errorx.BadRequest, "Require a positive level")
	}

	if req.Value <= 0 {
		return nil, errorx.New(errorx.BadRequest, "Require a positive value")
	}

	_, err := url.ParseRequestURI(req.IconURL)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid icon url: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid icon url")
	}

	err = d.badgeRepo.Create(ctx, &entity.Badge{
		Base:        entity.Base{ID: uuid.NewString()},
		Name:        req.Name,
		Level:       req.Level,
		Value:       req.Value,
		Description: req.Description,
		IconURL:     req.IconURL,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create or update badges: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateBadgeResponse{}, nil
}

func (d *badgeDomain) GetUserBadgeDetails(
	ctx context.Context, req *model.GetUserBadgeDetailsRequest,
) (*model.GetUserBadgeDetailsResponse, error) {
	var community *entity.Community
	if req.CommunityHandle != "" {
		var err error
		community, err = d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found community")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}
	}

	communityID := ""
	if community != nil {
		communityID = community.ID
	}

	badgeDetails, err := d.badgeDetailRepo.GetAll(ctx, req.UserID, communityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get badge details: %v", err)
		return nil, errorx.Unknown
	}

	badges, err := d.badgeRepo.GetAll(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get badges: %v", err)
		return nil, errorx.Unknown
	}

	badgeMap := map[string]entity.Badge{}
	for _, b := range badges {
		badgeMap[b.ID] = b
	}

	clientBadgeDetails := []model.BadgeDetail{}
	for _, detail := range badgeDetails {
		badge, ok := badgeMap[detail.BadgeID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found badge %s", detail.BadgeID)
			return nil, errorx.Unknown
		}

		clientBadgeDetails = append(
			clientBadgeDetails,
			convertBadgeDetail(
				&detail,
				convertUser(nil, nil),
				convertCommunity(community, 0),
				convertBadge(&badge),
			),
		)
	}

	return &model.GetUserBadgeDetailsResponse{BadgeDetails: clientBadgeDetails}, nil
}

func (d *badgeDomain) GetMyBadgeDetails(
	ctx context.Context, req *model.GetMyBadgeDetailsRequest,
) (*model.GetMyBadgeDetailsResponse, error) {
	var community *entity.Community
	if req.CommunityHandle != "" {
		var err error
		community, err = d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found community")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}
	}

	communityID := ""
	if community != nil {
		communityID = community.ID
	}

	requestUserID := xcontext.RequestUserID(ctx)
	badgeDetails, err := d.badgeDetailRepo.GetAll(ctx, requestUserID, communityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get badges: %v", err)
		return nil, errorx.Unknown
	}

	badges, err := d.badgeRepo.GetAll(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get badges: %v", err)
		return nil, errorx.Unknown
	}

	badgeMap := map[string]entity.Badge{}
	for _, b := range badges {
		badgeMap[b.ID] = b
	}

	clientBadgeDetails := []model.BadgeDetail{}
	needUpdate := false
	for _, detail := range badgeDetails {
		badge, ok := badgeMap[detail.BadgeID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found badge %s", detail.BadgeID)
			return nil, errorx.Unknown
		}

		clientBadgeDetails = append(
			clientBadgeDetails,
			convertBadgeDetail(
				&detail,
				convertUser(nil, nil),
				convertCommunity(community, 0),
				convertBadge(&badge),
			),
		)

		if !detail.WasNotified {
			needUpdate = true
		}
	}

	if needUpdate {
		if err := d.badgeDetailRepo.UpdateNotification(ctx, requestUserID, communityID); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot update notification of badge: %v", err)
			return nil, errorx.Unknown
		}
	}

	return &model.GetMyBadgeDetailsResponse{BadgeDetails: clientBadgeDetails}, nil
}
