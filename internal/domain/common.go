package domain

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/structs"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/dateutil"
)

const defaultTimeLayout = time.RFC3339Nano

func convertRewards(entityRewards []entity.Reward) []model.Reward {
	modelRewards := []model.Reward{}
	for _, r := range entityRewards {
		modelRewards = append(modelRewards, model.Reward{Type: string(r.Type), Data: r.Data})
	}
	return modelRewards
}

func convertConditions(entityConditions []entity.Condition) []model.Condition {
	modelConditions := []model.Condition{}
	for _, r := range entityConditions {
		modelConditions = append(modelConditions, model.Condition{Type: string(r.Type), Data: r.Data})
	}
	return modelConditions
}

func convertUser(user *entity.User, serviceUsers []entity.OAuth2) model.User {
	if user == nil {
		return model.User{}
	}

	serviceMap := map[string]string{}
	for _, u := range serviceUsers {
		tag, id, found := strings.Cut(u.ServiceUserID, "_")
		if !found || tag != u.Service {
			continue
		}

		serviceMap[u.Service] = id
	}

	return model.User{
		ID:            user.ID,
		Name:          user.Name,
		WalletAddress: user.WalletAddress.String,
		Role:          string(user.Role),
		ReferralCode:  user.ReferralCode,
		Services:      serviceMap,
		IsNewUser:     user.IsNewUser,
		AvatarURL:     user.ProfilePicture,
	}
}

func convertCategory(category *entity.Category) model.Category {
	if category == nil {
		return model.Category{}
	}

	return model.Category{
		ID:        category.ID,
		Name:      category.Name,
		CreatedBy: category.CreatedBy,
		CreatedAt: category.CreatedAt.Format(defaultTimeLayout),
		UpdatedAt: category.UpdatedAt.Format(defaultTimeLayout),
	}
}

func convertCommunity(community *entity.Community) model.Community {
	if community == nil {
		return model.Community{}
	}

	return model.Community{
		Handle:             community.Handle,
		CreatedAt:          community.CreatedAt.Format(defaultTimeLayout),
		UpdatedAt:          community.UpdatedAt.Format(defaultTimeLayout),
		ReferredBy:         community.ReferredBy.String,
		ReferralStatus:     string(community.ReferralStatus),
		CreatedBy:          community.CreatedBy,
		Introduction:       string(community.Introduction),
		DisplayName:        community.DisplayName,
		Twitter:            community.Twitter,
		Discord:            community.Discord,
		Followers:          community.Followers,
		TrendingScore:      community.TrendingScore,
		LogoURL:            community.LogoPicture,
		WebsiteURL:         community.WebsiteURL,
		DevelopmentStage:   community.DevelopmentStage,
		TeamSize:           community.TeamSize,
		SharedContentTypes: community.SharedContentTypes,
	}
}

func convertBadge(badge *entity.Badge, user model.User, community model.Community) model.Badge {
	if badge == nil {
		return model.Badge{}
	}

	if user.ID == "" {
		user = model.User{ID: badge.UserID}
	}

	return model.Badge{
		User:        user,
		Community:   community,
		Name:        badge.Name,
		Level:       badge.Level,
		WasNotified: badge.WasNotified,
	}
}

func convertQuest(quest *entity.Quest, community model.Community, category model.Category) model.Quest {
	if quest == nil {
		return model.Quest{}
	}

	if category.ID == "" {
		category = model.Category{ID: quest.CategoryID.String}
	}

	return model.Quest{
		ID:             quest.ID,
		Community:      community,
		Type:           string(quest.Type),
		Status:         string(quest.Status),
		Title:          quest.Title,
		Description:    string(quest.Description),
		Category:       category,
		Recurrence:     string(quest.Recurrence),
		ValidationData: quest.ValidationData,
		Points:         quest.Points,
		Rewards:        convertRewards(quest.Rewards),
		ConditionOp:    string(quest.ConditionOp),
		Conditions:     convertConditions(quest.Conditions),
		CreatedAt:      quest.CreatedAt.Format(defaultTimeLayout),
		UpdatedAt:      quest.UpdatedAt.Format(defaultTimeLayout),
	}
}

func convertClaimedQuest(
	claimedQuest *entity.ClaimedQuest, quest model.Quest, user model.User,
) model.ClaimedQuest {
	if claimedQuest == nil {
		return model.ClaimedQuest{}
	}

	if quest.ID == "" {
		quest = model.Quest{ID: claimedQuest.QuestID}
	}

	if user.ID == "" {
		user = model.User{ID: claimedQuest.UserID}
	}

	return model.ClaimedQuest{
		ID:             claimedQuest.ID,
		Quest:          quest,
		User:           user,
		SubmissionData: claimedQuest.SubmissionData,
		Status:         string(claimedQuest.Status),
		ReviewerID:     claimedQuest.ReviewerID,
		ReviewedAt:     claimedQuest.ReviewedAt.Format(defaultTimeLayout),
		Comment:        claimedQuest.Comment,
		CreatedAt:      claimedQuest.CreatedAt.Format(defaultTimeLayout),
		UpdatedAt:      claimedQuest.UpdatedAt.Format(defaultTimeLayout),
	}
}

func convertCollaborator(
	collaborator *entity.Collaborator, community model.Community, user model.User,
) model.Collaborator {
	if collaborator == nil {
		return model.Collaborator{}
	}

	if user.ID == "" {
		user = model.User{ID: collaborator.UserID}
	}

	return model.Collaborator{
		User:      user,
		Community: community,
		Role:      string(collaborator.Role),
		CreatedBy: collaborator.CreatedBy,
	}
}

func convertFollower(follower *entity.Follower, user model.User, community model.Community) model.Follower {
	if follower == nil {
		return model.Follower{}
	}

	if user.ID == "" {
		user = model.User{ID: follower.UserID}
	}

	return model.Follower{
		User:        user,
		Community:   community,
		Points:      follower.Points,
		Quests:      follower.Quests,
		Streaks:     follower.Streaks,
		InviteCode:  follower.InviteCode,
		InvitedBy:   follower.InvitedBy.String,
		InviteCount: follower.InviteCount,
	}
}

func processValidationData(
	ctx context.Context, questFactory questclaim.Factory, includeSecret bool, quest *entity.Quest,
) error {
	processor, err := questFactory.LoadProcessor(ctx, includeSecret, *quest, quest.ValidationData)
	if err != nil {
		return err
	}

	quest.ValidationData = structs.Map(processor)
	return nil
}

func stringToPeriodWithTime(periodString string, current time.Time) (entity.LeaderBoardPeriodType, error) {
	switch periodString {
	case "week":
		return entity.NewLeaderBoardPeriodWeek(current), nil
	case "month":
		return entity.NewLeaderBoardPeriodMonth(current), nil
	}

	return nil, fmt.Errorf("invalid period, expected week or month, but got %s", periodString)
}

func stringToPeriod(periodString string) (entity.LeaderBoardPeriodType, error) {
	switch periodString {
	case "week":
		return entity.NewLeaderBoardPeriodWeek(time.Now()), nil
	case "month":
		return entity.NewLeaderBoardPeriodMonth(time.Now()), nil
	}

	return nil, fmt.Errorf("invalid period, expected week or month, but got %s", periodString)
}

func stringToLastPeriod(periodString string) (entity.LeaderBoardPeriodType, error) {
	switch periodString {
	case "week":
		return entity.NewLeaderBoardPeriodWeek(dateutil.LastWeek(time.Now())), nil
	case "month":
		return entity.NewLeaderBoardPeriodMonth(dateutil.LastMonth(time.Now())), nil
	}

	return nil, fmt.Errorf("invalid period, expected week or month, but got %s", periodString)
}

func checkCommunityHandle(handle string) error {
	if len(handle) < 4 {
		return errors.New("too short")
	}

	if len(handle) > 32 {
		return errors.New("too long")
	}

	ok, err := regexp.MatchString("^[a-z0-9_]*$", handle)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("invalid name")
	}

	return nil
}

func checkCommunityDisplayName(displayName string) error {
	if len(displayName) < 4 {
		return errors.New("too short")
	}

	if len(displayName) > 32 {
		return errors.New("too long")
	}

	return nil
}

func generateCommunityHandle(displayName string) string {
	handle := []rune{}
	for _, c := range displayName {
		if isAsciiLetter(c) {
			handle = append(handle, unicode.ToLower(c))
		} else if c == ' ' {
			handle = append(handle, '_')
		}
	}

	return string(handle)
}

func isAsciiLetter(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') || c == '_'
}
