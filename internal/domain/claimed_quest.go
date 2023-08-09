package domain

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/dateutil"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
	"gorm.io/gorm"
)

type ClaimedQuestDomain interface {
	Claim(context.Context, *model.ClaimQuestRequest) (*model.ClaimQuestResponse, error)
	ClaimReferral(context.Context, *model.ClaimReferralRequest) (*model.ClaimReferralResponse, error)
	Get(context.Context, *model.GetClaimedQuestRequest) (*model.GetClaimedQuestResponse, error)
	GetList(context.Context, *model.GetListClaimedQuestRequest) (*model.GetListClaimedQuestResponse, error)
	Review(context.Context, *model.ReviewRequest) (*model.ReviewResponse, error)
	ReviewAll(context.Context, *model.ReviewAllRequest) (*model.ReviewAllResponse, error)
	GivePoint(context.Context, *model.GivePointRequest) (*model.GivePointResponse, error)
}

type claimedQuestDomain struct {
	claimedQuestRepo         repository.ClaimedQuestRepository
	questRepo                repository.QuestRepository
	followerRepo             repository.FollowerRepository
	followerRoleRepo         repository.FollowerRoleRepository
	communityRepo            repository.CommunityRepository
	categoryRepo             repository.CategoryRepository
	roleVerifier             *common.CommunityRoleVerifier
	userRepo                 repository.UserRepository
	questFactory             questclaim.Factory
	badgeManager             *badge.Manager
	leaderboard              statistic.Leaderboard
	notificationEngineCaller client.NotificationEngineCaller
	redisClient              xredis.Client
}

func NewClaimedQuestDomain(
	claimedQuestRepo repository.ClaimedQuestRepository,
	questRepo repository.QuestRepository,
	followerRepo repository.FollowerRepository,
	followerRoleRepo repository.FollowerRoleRepository,
	userRepo repository.UserRepository,
	communityRepo repository.CommunityRepository,
	categoryRepo repository.CategoryRepository,
	badgeManager *badge.Manager,
	leaderboard statistic.Leaderboard,
	roleVerifier *common.CommunityRoleVerifier,
	notificationEngineCaller client.NotificationEngineCaller,
	questFactory questclaim.Factory,
	redisClient xredis.Client,
) *claimedQuestDomain {
	return &claimedQuestDomain{
		claimedQuestRepo:         claimedQuestRepo,
		questRepo:                questRepo,
		followerRepo:             followerRepo,
		followerRoleRepo:         followerRoleRepo,
		userRepo:                 userRepo,
		communityRepo:            communityRepo,
		roleVerifier:             roleVerifier,
		categoryRepo:             categoryRepo,
		questFactory:             questFactory,
		badgeManager:             badgeManager,
		leaderboard:              leaderboard,
		notificationEngineCaller: notificationEngineCaller,
		redisClient:              redisClient,
	}
}

func (d *claimedQuestDomain) Claim(
	ctx context.Context, req *model.ClaimQuestRequest,
) (*model.ClaimQuestResponse, error) {
	quest, err := d.questRepo.GetByID(ctx, req.QuestID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	if quest.Status != entity.QuestActive {
		return nil, errorx.New(errorx.Unavailable, "Only allow to claim active quests")
	}

	if !quest.CommunityID.Valid {
		return nil, errorx.New(errorx.Unavailable, "Unable to claim a template")
	}

	// Check if user follows the community.
	requestUserID := xcontext.RequestUserID(ctx)
	_, err = d.followerRepo.Get(ctx, requestUserID, quest.CommunityID.String)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			xcontext.Logger(ctx).Errorf("Cannot get the follower: %v", err)
			return nil, errorx.Unknown
		}

		err := followCommunity(
			ctx,
			d.userRepo,
			d.communityRepo,
			d.followerRepo,
			d.followerRoleRepo,
			d.notificationEngineCaller,
			d.redisClient,
			requestUserID, quest.CommunityID.String, "",
		)
		if err != nil {
			return nil, err
		}
	}

	// Check the condition and recurrence.
	reason, err := d.questFactory.IsClaimable(ctx, *quest)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot determine claimable: %v", err)
		return nil, errorx.Unknown
	}

	if reason != nil {
		return nil, errorx.New(errorx.Unavailable, reason.Message)
	}

	// Auto review the action/submission data of user with validation data.
	// After this step, we can determine if the quest user claimed is accepted,
	// rejected, or need a manual review.
	processor, err := d.questFactory.LoadProcessor(ctx, true, *quest, quest.ValidationData)
	if err != nil {
		return nil, err
	}

	actionForClaim, err := processor.GetActionForClaim(ctx, req.SubmissionData)
	if err != nil {
		return nil, err
	}

	var status entity.ClaimedQuestStatus
	switch {
	case actionForClaim.Is(questclaim.Accepted):
		status = entity.AutoAccepted
	case actionForClaim.Is(questclaim.Rejected):
		status = entity.AutoRejected
	case actionForClaim.Is(questclaim.NeedManualReview):
		status = entity.Pending
	}

	// Store the ClaimedQuest into database.
	claimedQuest := &entity.ClaimedQuest{
		Base:           entity.Base{ID: uuid.NewString()},
		QuestID:        req.QuestID,
		UserID:         xcontext.RequestUserID(ctx),
		Status:         status,
		SubmissionData: req.SubmissionData,
		ReviewedAt:     sql.NullTime{Valid: false},
		WalletAddress:  req.WalletAddress,
	}

	if status != entity.Pending {
		claimedQuest.ReviewedAt = sql.NullTime{Valid: true, Time: time.Now()}
	}

	// GiveReward can write something to database.
	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	// If the claimed quest is auto accepted or pending (even rejected after
	// reviewing), the streak will be stacked.
	if status != entity.AutoRejected {
		// Get the last claimed quest (accepted or pending of any quest) to
		// calculate streak.
		lastClaimedAnyQuest, err := d.claimedQuestRepo.GetLast(ctx, repository.GetLastClaimedQuestFilter{
			UserID: xcontext.RequestUserID(ctx), CommunityID: quest.CommunityID.String,
			Status: []entity.ClaimedQuestStatus{entity.Pending, entity.Accepted, entity.AutoAccepted},
		})
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			xcontext.Logger(ctx).Errorf("Cannot get claimed quest: %v", err)
			return nil, errorx.Unknown
		}

		isStreak := false
		if err == nil && dateutil.IsYesterday(lastClaimedAnyQuest.CreatedAt, claimedQuest.CreatedAt) {
			isStreak = true
		}

		err = d.followerRepo.UpdateStreak(ctx, requestUserID, quest.CommunityID.String, isStreak)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot update streak of follower: %v", err)
			return nil, errorx.Unknown
		}

		err = d.badgeManager.
			WithBadges(badge.RainBowBadgeName).
			ScanAndGive(ctx, requestUserID, quest.CommunityID.String)
		if err != nil {
			return nil, err
		}
	}

	err = d.claimedQuestRepo.Create(ctx, claimedQuest)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot claim quest: %v", err)
		return nil, errorx.Unknown
	}

	// Give reward to user if the claimed quest is accepted.
	if status == entity.AutoAccepted {
		if err := d.giveReward(ctx, *quest, *claimedQuest); err != nil {
			return nil, err
		}
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.ClaimQuestResponse{
		ID:      claimedQuest.ID,
		Status:  string(status),
		Message: actionForClaim.Message(),
	}, nil
}

func (d *claimedQuestDomain) ClaimReferral(
	ctx context.Context, req *model.ClaimReferralRequest,
) (*model.ClaimReferralResponse, error) {
	requestUserID := xcontext.RequestUserID(ctx)
	communities, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
		ReferredBy:     requestUserID,
		ReferralStatus: []entity.ReferralStatusType{entity.ReferralClaimable},
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get claimable referral communities: %v", err)
		return nil, errorx.Unknown
	}

	if len(communities) == 0 {
		return nil, errorx.New(errorx.Unavailable, "Not found any claimable referral community")
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	for _, community := range communities {
		referralReward, err := d.questFactory.LoadReferralReward(ctx)
		if err != nil {
			return nil, err
		}

		referralReward.WithReferralCommunity(&community)
		referralReward.WithWalletAddress(req.WalletAddress)
		if err := referralReward.Give(ctx); err != nil {
			return nil, err
		}

		err = d.communityRepo.UpdateReferralStatusByID(ctx, community.ID, entity.ReferralClaimed)
		if err != nil {
			xcontext.Logger(ctx).Errorf(
				"Cannot update referral status of community %s: %v", community.ID, err)
			return nil, errorx.Unknown
		}
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.ClaimReferralResponse{}, nil
}

func (d *claimedQuestDomain) Get(
	ctx context.Context, req *model.GetClaimedQuestRequest,
) (*model.GetClaimedQuestResponse, error) {
	if req.ID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty id")
	}

	claimedQuest, err := d.claimedQuestRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found claimed quest")
		}

		xcontext.Logger(ctx).Errorf("Cannot get claimed quest: %v", err)
		return nil, errorx.Unknown
	}

	quest, err := d.questRepo.GetByIDIncludeSoftDeleted(ctx, claimedQuest.QuestID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get quest: %v", err)
		return nil, errorx.Unknown
	}

	if claimedQuest.UserID != xcontext.RequestUserID(ctx) {
		if err = d.roleVerifier.Verify(ctx, quest.CommunityID.String); err != nil {
			xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
			return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
		}
	}

	user, err := d.userRepo.GetByID(ctx, claimedQuest.UserID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get users: %v", err)
		return nil, errorx.Unknown
	}

	var community *entity.Community
	if quest.CommunityID.Valid {
		var err error
		community, err = d.communityRepo.GetByID(ctx, quest.CommunityID.String)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found community")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}
	}

	var category *entity.Category
	if quest.CategoryID.Valid {
		category, err = d.categoryRepo.GetByID(ctx, quest.CategoryID.String)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get category: %v", err)
			return nil, errorx.Unknown
		}
	}

	resp := model.GetClaimedQuestResponse(model.ConvertClaimedQuest(
		claimedQuest,
		model.ConvertQuest(quest, model.ConvertCommunity(community, 0), model.ConvertCategory(category)),
		model.ConvertShortUser(user, ""),
	))
	return &resp, nil
}

func (d *claimedQuestDomain) GetList(
	ctx context.Context, req *model.GetListClaimedQuestRequest,
) (*model.GetListClaimedQuestResponse, error) {
	if req.CommunityHandle == "" && req.UserID == "" {
		return nil, errorx.New(errorx.BadRequest, "Require community_handle or user_id")
	}

	var communityID string
	if req.CommunityHandle != "" {
		community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found community")
			}

			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}

		communityID = community.ID
	}

	err := func() error {
		// User can get his claimed quest history without community reviewer
		// permission.
		if req.UserID != "" && xcontext.RequestUserID(ctx) == req.UserID {
			return nil
		}

		if communityID != "" {
			return d.roleVerifier.Verify(ctx, communityID)
		}

		return errors.New("not allow getting claimed quest of another user")
	}()
	if err != nil {
		xcontext.Logger(ctx).Errorf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	apiCfg := xcontext.Configs(ctx).ApiServer
	if req.Limit == 0 {
		req.Limit = apiCfg.DefaultLimit
	}

	if req.Limit < 0 {
		return nil, errorx.New(errorx.BadRequest, "Limit must be positive")
	}

	if req.Limit > apiCfg.MaxLimit {
		return nil, errorx.New(errorx.BadRequest, "Exceed the maximum of limit (%d)", apiCfg.MaxLimit)
	}

	var statusFilter []entity.ClaimedQuestStatus
	if req.Status != "" {
		status := strings.Split(req.Status, ",")
		for _, s := range status {
			statusEnum, err := enum.ToEnum[entity.ClaimedQuestStatus](s)
			if err != nil {
				xcontext.Logger(ctx).Debugf("Invalid claimed quest status: %v", err)
				return nil, errorx.New(errorx.BadRequest, "Invalid status filter")
			}

			statusFilter = append(statusFilter, statusEnum)
		}
	}

	var recurrenceFilter []entity.RecurrenceType
	if req.Recurrence != "" {
		recurrences := strings.Split(req.Recurrence, ",")
		for _, recurrence := range recurrences {
			recurrenceEnum, err := enum.ToEnum[entity.RecurrenceType](recurrence)
			if err != nil {
				xcontext.Logger(ctx).Debugf("Invalid recurrence: %v", err)
				return nil, errorx.New(errorx.BadRequest, "Invalid recurrence filter")
			}

			recurrenceFilter = append(recurrenceFilter, recurrenceEnum)
		}
	}

	var userIDFilter []string
	if len(req.UserID) > 0 {
		userIDFilter = strings.Split(req.UserID, ",")
	}

	var questIDFilter []string
	if len(req.QuestID) > 0 {
		questIDFilter = strings.Split(req.QuestID, ",")
	}

	claimedQuests, err := d.claimedQuestRepo.GetList(
		ctx,
		&repository.ClaimedQuestFilter{
			CommunityID: communityID,
			Status:      statusFilter,
			Recurrences: recurrenceFilter,
			QuestIDs:    questIDFilter,
			UserIDs:     userIDFilter,
			Offset:      req.Offset,
			Limit:       req.Limit,
			ReverseTime: req.ReverseTime,
		},
	)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get list claimed quest: %v", err)
		return nil, errorx.Unknown
	}

	questMap := map[string]*entity.Quest{}
	userMap := map[string]*entity.User{}
	for _, cq := range claimedQuests {
		questMap[cq.QuestID] = nil
		userMap[cq.UserID] = nil
	}

	quests, err := d.questRepo.GetByIDsIncludeSoftDeleted(ctx, common.MapKeys(questMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get quests: %v", err)
		return nil, errorx.Unknown
	}

	communityMap := map[string]*entity.Community{}
	for i := range quests {
		questMap[quests[i].ID] = &quests[i]

		if !quests[i].CommunityID.Valid {
			xcontext.Logger(ctx).Errorf("Got a claimed quest of a template %s", quests[i].ID)
			return nil, errorx.Unknown
		}

		communityMap[quests[i].CommunityID.String] = nil
	}

	communities, err := d.communityRepo.GetByIDs(ctx, common.MapKeys(communityMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get communities: %v", err)
		return nil, errorx.Unknown
	}

	for i := range communities {
		communityMap[communities[i].ID] = &communities[i]
	}

	users, err := d.userRepo.GetByIDs(ctx, common.MapKeys(userMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get users: %v", err)
		return nil, errorx.Unknown
	}

	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	categories, err := d.categoryRepo.GetList(ctx, communityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get category: %v", err)
		return nil, errorx.Unknown
	}

	categoryMap := map[string]*entity.Category{}
	for i := range categories {
		categoryMap[categories[i].ID] = &categories[i]
	}

	clientClaimedQuests := []model.ClaimedQuest{}
	for _, cq := range claimedQuests {
		quest, ok := questMap[cq.QuestID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found quest %s in claimed quest %s", cq.QuestID, cq.ID)
			continue
		}

		community, ok := communityMap[quest.CommunityID.String]
		if !ok {
			xcontext.Logger(ctx).Errorf(
				"Not found community %s in claimed quest %s", quest.CommunityID.String, cq.ID)
			return nil, errorx.Unknown
		}

		user, ok := userMap[cq.UserID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Not found user %s in claimed quest %s", cq.UserID, cq.ID)
			return nil, errorx.Unknown
		}

		var category *entity.Category
		if quest.CategoryID.Valid {
			var ok bool
			category, ok = categoryMap[quest.CategoryID.String]
			if !ok {
				xcontext.Logger(ctx).Errorf("Invalid category id %s", quest.CategoryID.String)
				return nil, errorx.Unknown
			}
		}

		clientClaimedQuests = append(
			clientClaimedQuests,
			model.ConvertClaimedQuest(
				&cq,
				model.ConvertQuest(
					quest,
					model.ConvertCommunity(community, 0),
					model.ConvertCategory(category),
				),
				model.ConvertShortUser(user, ""),
			),
		)
	}

	return &model.GetListClaimedQuestResponse{ClaimedQuests: clientClaimedQuests}, nil
}

func (d *claimedQuestDomain) Review(
	ctx context.Context, req *model.ReviewRequest,
) (*model.ReviewResponse, error) {
	if len(req.IDs) == 0 {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty id")
	}

	firstClaimedQuest, err := d.claimedQuestRepo.GetByID(ctx, req.IDs[0])
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot get the first claimed quest: %v", err)
		return nil, errorx.Unknown
	}

	firstQuest, err := d.questRepo.GetByIDIncludeSoftDeleted(ctx, firstClaimedQuest.QuestID)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot get the first quest: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, firstQuest.CommunityID.String); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	claimedQuests, err := d.claimedQuestRepo.GetByIDs(ctx, req.IDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get claimed quest by ids: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.review(ctx, claimedQuests, req.Action, req.Comment); err != nil {
		return nil, err
	}

	return &model.ReviewResponse{}, nil
}

func (d *claimedQuestDomain) ReviewAll(
	ctx context.Context, req *model.ReviewAllRequest,
) (*model.ReviewAllResponse, error) {
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow an empty community id")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, community.ID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	var recurrenceFilter []entity.RecurrenceType
	for _, recurrence := range req.Recurrences {
		recurrenceEnum, err := enum.ToEnum[entity.RecurrenceType](recurrence)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Invalid recurrence: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid recurrence filter")
		}

		recurrenceFilter = append(recurrenceFilter, recurrenceEnum)
	}

	var filterStatuses []entity.ClaimedQuestStatus
	for _, s := range req.Statuses {
		status, err := enum.ToEnum[entity.ClaimedQuestStatus](s)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Invalid filter status: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid filter status")
		}

		filterStatuses = append(filterStatuses, status)
	}

	if len(filterStatuses) == 0 {
		// Support back compatible.
		filterStatuses = []entity.ClaimedQuestStatus{entity.Pending}
	}

	claimedQuests, err := d.claimedQuestRepo.GetList(
		ctx,
		&repository.ClaimedQuestFilter{
			CommunityID: community.ID,
			QuestIDs:    req.QuestIDs,
			UserIDs:     req.UserIDs,
			Status:      filterStatuses,
			Recurrences: recurrenceFilter,
			Offset:      0,
			Limit:       -1,
		},
	)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get claimed quest: %v", err)
		return nil, errorx.Unknown
	}

	excludeMap := map[string]any{}
	for _, id := range req.Excludes {
		excludeMap[id] = nil
	}

	finalClaimedQuests := []entity.ClaimedQuest{}
	for _, cq := range claimedQuests {
		if _, ok := excludeMap[cq.ID]; !ok {
			finalClaimedQuests = append(finalClaimedQuests, cq)
		}
	}

	if err := d.review(ctx, finalClaimedQuests, req.Action, req.Comment); err != nil {
		return nil, err
	}

	return &model.ReviewAllResponse{Quantity: len(finalClaimedQuests)}, nil
}

func (d *claimedQuestDomain) review(
	ctx context.Context,
	claimedQuests []entity.ClaimedQuest,
	action string,
	comment string,
) error {
	if len(claimedQuests) == 0 {
		return errorx.New(errorx.Unavailable, "No claimed quest will be reviewed")
	}

	reviewAction, err := enum.ToEnum[entity.ClaimedQuestStatus](action)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid review action: %v", err)
		return errorx.New(errorx.BadRequest, "Invalid action")
	}

	questSet := map[string]any{}
	claimedQuestSet := map[string]any{}
	for _, cq := range claimedQuests {
		switch reviewAction {
		case entity.Pending: // Unapprove
			if cq.Status != entity.Accepted && cq.Status != entity.Rejected {
				return errorx.New(errorx.BadRequest, "Claimed quest must be accepted or rejected")
			}
		case entity.Accepted, entity.Rejected:
			if cq.Status != entity.Pending {
				return errorx.New(errorx.BadRequest, "Claimed quest must be pending")
			}
		default:
			return errorx.New(errorx.BadRequest, "Review action must be accepted, rejected, or pending")
		}

		claimedQuestSet[cq.ID] = nil
		questSet[cq.QuestID] = nil
	}

	quests, err := d.questRepo.GetByIDsIncludeSoftDeleted(ctx, common.MapKeys(questSet))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get quest: %v", err)
		return errorx.Unknown
	}

	questInverse := map[string]entity.Quest{}
	for _, q := range quests {
		if q.CommunityID != quests[0].CommunityID {
			return errorx.New(errorx.BadRequest, "You can only review claimed quests of one community")
		}

		questInverse[q.ID] = q
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	requestUserID := xcontext.RequestUserID(ctx)
	err = d.claimedQuestRepo.UpdateReviewByIDs(
		ctx, common.MapKeys(claimedQuestSet),
		&entity.ClaimedQuest{
			Status:     reviewAction,
			Comment:    comment,
			ReviewerID: requestUserID,
			ReviewedAt: sql.NullTime{Valid: true, Time: time.Now()},
		},
	)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to update status: %v", err)
		return errorx.New(errorx.Internal, "Unable to update status for claim quest")
	}

	switch reviewAction {
	case entity.Accepted:
		for _, claimedQuest := range claimedQuests {
			quest, ok := questInverse[claimedQuest.QuestID]
			if !ok {
				xcontext.Logger(ctx).Errorf(
					"Not found quest %s of claimed quest %s", claimedQuest.QuestID, claimedQuest.ID)
				return errorx.Unknown
			}

			if err := d.giveReward(ctx, quest, claimedQuest); err != nil {
				return err
			}
		}
	case entity.Pending: // Unapprove
		for _, claimedQuest := range claimedQuests {
			if claimedQuest.Status != entity.Accepted {
				continue
			}

			quest, ok := questInverse[claimedQuest.QuestID]
			if !ok {
				xcontext.Logger(ctx).Errorf(
					"Not found quest %s of claimed quest %s", claimedQuest.QuestID, claimedQuest.ID)
				return errorx.Unknown
			}

			if err := d.revertQuest(ctx, quest, claimedQuest); err != nil {
				return err
			}
		}
	}

	xcontext.WithCommitDBTransaction(ctx)
	return nil
}

func (d *claimedQuestDomain) GivePoint(
	ctx context.Context, req *model.GivePointRequest,
) (*model.GivePointResponse, error) {
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community handle")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, community.ID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied when give reward: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Only community owner can give reward directly")
	}

	if _, err := d.followerRepo.Get(ctx, req.UserID, community.ID); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			xcontext.Logger(ctx).Errorf("Cannot get the follower: %v", err)
			return nil, errorx.Unknown
		}

		return nil, errorx.New(errorx.Unavailable, "User must follow the community before")
	}

	err = d.followerRepo.IncreasePoint(ctx, req.UserID, community.ID, req.Points, false)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot increase points: %v", err)
		return nil, errorx.Unknown
	}

	err = d.leaderboard.ChangePointLeaderboard(ctx, int64(req.Points), time.Now(), req.UserID, community.ID)
	if err != nil {
		return nil, err
	}

	return &model.GivePointResponse{}, nil
}

func (d *claimedQuestDomain) giveReward(
	ctx context.Context,
	quest entity.Quest,
	claimedQuest entity.ClaimedQuest,
) error {
	for _, data := range quest.Rewards {
		reward, err := d.questFactory.LoadReward(ctx, quest.CommunityID.String, data.Type, data.Data)
		if err != nil {
			xcontext.Logger(ctx).Warnf("Invalid reward data: %v", err)
			continue
		}

		reward.WithClaimedQuest(&claimedQuest)
		if err := reward.Give(ctx); err != nil {
			return err
		}
	}

	err := d.followerRepo.IncreasePoint(
		ctx, claimedQuest.UserID, quest.CommunityID.String, quest.Points, true)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot increase points for user: %v", err)
		return errorx.Unknown
	}

	follower, err := d.followerRepo.Get(ctx, xcontext.RequestUserID(ctx), quest.CommunityID.String)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get follower info of user: %v", err)
		return errorx.Unknown
	}

	if follower.Quests == 1 && follower.InvitedBy.Valid {
		err := d.followerRepo.IncreaseInviteCount(ctx, follower.InvitedBy.String, quest.CommunityID.String)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorx.New(errorx.NotFound, "Invalid invite user id")
			}

			xcontext.Logger(ctx).Errorf("Cannot increase invite: %v", err)
			return errorx.Unknown
		}

		err = d.badgeManager.
			WithBadges(badge.SharpScoutBadgeName).
			ScanAndGive(ctx, follower.InvitedBy.String, quest.CommunityID.String)
		if err != nil {
			return err
		}
	}

	err = d.badgeManager.
		WithBadges(badge.QuestWarriorBadgeName).
		ScanAndGive(ctx, claimedQuest.UserID, quest.CommunityID.String)
	if err != nil {
		return err
	}

	reviewedAt := claimedQuest.ReviewedAt.Time
	userID := claimedQuest.UserID
	communityID := quest.CommunityID.String

	err = d.leaderboard.ChangeQuestLeaderboard(ctx, 1, reviewedAt, userID, communityID)
	if err != nil {
		return err
	}

	err = d.leaderboard.ChangePointLeaderboard(ctx, int64(quest.Points), reviewedAt, userID, communityID)
	if err != nil {
		return err
	}

	return nil
}

func (d *claimedQuestDomain) revertQuest(
	ctx context.Context,
	quest entity.Quest,
	claimedQuest entity.ClaimedQuest,
) error {
	err := d.followerRepo.DecreasePoint(
		ctx, claimedQuest.UserID, quest.CommunityID.String, quest.Points, true)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to complete quest for user: %v", err)
		return errorx.Unknown
	}

	reviewedAt := claimedQuest.ReviewedAt.Time
	userID := claimedQuest.UserID
	communityID := quest.CommunityID.String

	err = d.leaderboard.ChangeQuestLeaderboard(ctx, -1, reviewedAt, userID, communityID)
	if err != nil {
		return err
	}

	err = d.leaderboard.ChangePointLeaderboard(ctx, -int64(quest.Points), reviewedAt, userID, communityID)
	if err != nil {
		return err
	}

	return nil
}
