package domain

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/structs"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
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

func convertUser(
	user *entity.User,
	serviceUsers []entity.OAuth2,
	includeSensitive bool,
) model.User {
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

	if !includeSensitive {
		user.Role = ""
		user.WalletAddress = sql.NullString{Valid: false, String: ""}
		user.IsNewUser = false
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

func convertCommunity(community *entity.Community, totalQuests int) model.Community {
	if community == nil {
		return model.Community{}
	}

	return model.Community{
		Handle:         community.Handle,
		CreatedAt:      community.CreatedAt.Format(defaultTimeLayout),
		UpdatedAt:      community.UpdatedAt.Format(defaultTimeLayout),
		ReferredBy:     community.ReferredBy.String,
		ReferralStatus: string(community.ReferralStatus),
		CreatedBy:      community.CreatedBy,
		Introduction:   string(community.Introduction),
		DisplayName:    community.DisplayName,
		Twitter:        community.Twitter,
		Discord:        community.Discord,
		Followers:      community.Followers,
		TrendingScore:  community.TrendingScore,
		LogoURL:        community.LogoPicture,
		WebsiteURL:     community.WebsiteURL,
		NumberOfQuests: totalQuests,
		Status:         string(community.Status),
		// Do not leak owner email. Only superadmin can see the owner email when
		// get pending communities.
	}
}

func convertBadge(badge *entity.Badge) model.Badge {
	if badge == nil {
		return model.Badge{}
	}

	return model.Badge{
		ID:          badge.ID,
		Name:        badge.Name,
		Level:       badge.Level,
		Description: badge.Description,
		IconURL:     badge.IconURL,
	}
}

func convertBadgeDetail(
	badgeDetail *entity.BadgeDetail,
	user model.User,
	community model.Community,
	badge model.Badge,
) model.BadgeDetail {
	if badgeDetail == nil {
		return model.BadgeDetail{}
	}

	if user.ID == "" {
		user = model.User{ID: badgeDetail.UserID}
	}

	if badge.ID == "" {
		badge = model.Badge{ID: badgeDetail.BadgeID}
	}

	return model.BadgeDetail{
		User:        user,
		Community:   community,
		Badge:       badge,
		WasNotified: badgeDetail.WasNotified,
		CreatedAt:   badgeDetail.CreatedAt.Format(defaultTimeLayout),
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
		IsHighlight:    quest.IsHighlight,
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

	reviewedAt := ""
	if claimedQuest.ReviewedAt.Valid {
		reviewedAt = claimedQuest.ReviewedAt.Time.Format(defaultTimeLayout)
	}

	return model.ClaimedQuest{
		ID:             claimedQuest.ID,
		Quest:          quest,
		User:           user,
		SubmissionData: claimedQuest.SubmissionData,
		Status:         string(claimedQuest.Status),
		ReviewerID:     claimedQuest.ReviewerID,
		ReviewedAt:     reviewedAt,
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

func convertGameMap(gameMap *entity.GameMap) model.GameMap {
	if gameMap == nil {
		return model.GameMap{}
	}

	return model.GameMap{
		ID:        gameMap.ID,
		Name:      gameMap.Name,
		ConfigURL: gameMap.ConfigURL,
	}
}

func convertGameCharacter(character *entity.GameCharacter) model.GameCharacter {
	if character == nil {
		return model.GameCharacter{}
	}

	return model.GameCharacter{
		ID:           character.ID,
		Name:         character.Name,
		Level:        character.Level,
		ConfigURL:    character.ConfigURL,
		ImageURL:     character.ImageURL,
		ThumbnailURL: character.ThumbnailURL,
		Points:       character.Points,
		CreatedAt:    character.CreatedAt.Format(defaultTimeLayout),
		UpdatedAt:    character.UpdatedAt.Format(defaultTimeLayout),
	}
}

func convertGameCommunityCharacter(
	communityCharacter *entity.GameCommunityCharacter,
	character model.GameCharacter,
) model.GameCommunityCharacter {
	if communityCharacter == nil {
		return model.GameCommunityCharacter{}
	}

	return model.GameCommunityCharacter{
		CommunityID:   communityCharacter.CommunityID,
		Points:        communityCharacter.Points,
		GameCharacter: character,
		CreatedAt:     communityCharacter.CreatedAt.Format(defaultTimeLayout),
		UpdatedAt:     communityCharacter.UpdatedAt.Format(defaultTimeLayout),
	}
}

func convertGameUserCharacter(
	userCharacter *entity.GameUserCharacter,
	character model.GameCharacter,
	isEquipped bool,
) model.GameUserCharacter {
	if userCharacter == nil {
		return model.GameUserCharacter{}
	}

	return model.GameUserCharacter{
		UserID:        userCharacter.UserID,
		CommunityID:   userCharacter.CommunityID,
		IsEquipped:    isEquipped,
		GameCharacter: character,
		UpdatedAt:     userCharacter.UpdatedAt.Format(defaultTimeLayout),
		CreatedAt:     userCharacter.CreatedAt.Format(defaultTimeLayout),
	}
}

func convertGameRoom(gameRoom *entity.GameRoom, gameMap model.GameMap) model.GameRoom {
	if gameRoom == nil {
		return model.GameRoom{}
	}

	if gameMap.ID == "" {
		gameMap = model.GameMap{ID: gameRoom.MapID}
	}

	return model.GameRoom{
		ID:   gameRoom.ID,
		Name: gameRoom.Name,
		Map:  gameMap,
	}
}

func convertDiscordRole(role discord.Role) model.DiscordRole {
	return model.DiscordRole{
		ID:       role.ID,
		Name:     role.Name,
		Position: role.Position,
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

func checkCommunityHandle(ctx context.Context, handle string) error {
	if len(handle) < 4 {
		return errorx.New(errorx.BadRequest, "Handle too short (at least 4 characters)")
	}

	if len(handle) > 32 {
		return errorx.New(errorx.BadRequest, "Handle too long (at most 32 characters)")
	}

	ok, err := regexp.MatchString("^[a-z0-9_]*$", handle)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot execute regex pattern: %v", err)
		return errorx.Unknown
	}

	if !ok {
		return errorx.New(errorx.BadRequest, "Name contains invalid characters")
	}

	return nil
}

func checkCommunityDisplayName(displayName string) error {
	if len(displayName) < 4 {
		return errorx.New(errorx.BadRequest, "Display name too short (at least 4 characters)")
	}

	if len(displayName) > 32 {
		return errorx.New(errorx.BadRequest, "Display name too long (at most 32 characters)")
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
