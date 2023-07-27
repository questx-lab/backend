package domain

import (
	"database/sql"
	"strings"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/api/discord"
)

const defaultTimeLayout string = time.RFC3339Nano

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

func convertRole(role *entity.Role) model.Role {
	if role == nil {
		return model.Role{}
	}

	return model.Role{
		ID:        role.ID,
		Name:      role.Name,
		Permision: role.Permissions,
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
		Position:       quest.Position,
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

func convertFollower(
	follower *entity.Follower, roles []model.Role, user model.User, community model.Community,
) model.Follower {
	if follower == nil {
		return model.Follower{}
	}

	if user.ID == "" {
		user = model.User{ID: follower.UserID}
	}

	return model.Follower{
		User:        user,
		Community:   community,
		Roles:       roles,
		Points:      follower.Points,
		Quests:      follower.Quests,
		Streaks:     follower.Streaks,
		InviteCode:  follower.InviteCode,
		InvitedBy:   follower.InvitedBy.String,
		InviteCount: follower.InviteCount,
		ChatLevel:   follower.ChatLevel,
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

func convertBlockchainConnection(c *entity.BlockchainConnection) model.BlockchainConnection {
	if c == nil {
		return model.BlockchainConnection{}
	}

	return model.BlockchainConnection{
		Type: string(c.Type),
		URL:  c.URL,
	}
}

func convertBlockchain(b *entity.Blockchain, connections []model.BlockchainConnection) model.Blockchain {
	if b == nil {
		return model.Blockchain{}
	}

	return model.Blockchain{
		Name:                  b.Name,
		ID:                    b.ID,
		UseExternalRPC:        b.UseExternalRPC,
		UseEip1559:            b.UseEip1559,
		BlockTime:             b.BlockTime,
		AdjustTime:            b.AdjustTime,
		ThresholdUpdateBlock:  b.ThresholdUpdateBlock,
		BlockchainConnections: connections,
	}
}

func convertBlockchainTransaction(tx *entity.BlockchainTransaction) model.BlockchainTransaction {
	if tx == nil {
		return model.BlockchainTransaction{}
	}

	return model.BlockchainTransaction{
		TxHash:    tx.TxHash,
		Chain:     tx.Chain,
		Status:    string(tx.Status),
		CreatedAt: tx.CreatedAt.Format(defaultTimeLayout),
		UpdatedAt: tx.UpdatedAt.Format(defaultTimeLayout),
	}
}

func convertBlockchainToken(token *entity.BlockchainToken) model.BlockchainToken {
	if token == nil {
		return model.BlockchainToken{}
	}

	return model.BlockchainToken{
		ID:      token.ID,
		Name:    token.Name,
		Symbol:  token.Symbol,
		Chain:   token.Chain,
		Address: token.Address,
	}
}

func convertPayReward(
	pw *entity.PayReward,
	token model.BlockchainToken,
	toUser model.User,
	referralCommunityHandle string,
	fromCommunityHandle string,
	tx model.BlockchainTransaction,
) model.PayReward {
	if pw == nil {
		return model.PayReward{}
	}

	if toUser.ID == "" {
		toUser = model.User{ID: pw.ToUserID}
	}

	return model.PayReward{
		ID:                      pw.ID,
		Token:                   token,
		ToUser:                  toUser,
		ClaimedQuestID:          pw.ClaimedQuestID.String,
		ReferralCommunityHandle: referralCommunityHandle,
		FromCommunityHandle:     fromCommunityHandle,
		ToAddress:               pw.ToAddress,
		Amount:                  pw.Amount,
		CreatedAt:               pw.CreatedAt.Format(defaultTimeLayout),
		UpdatedAt:               pw.UpdatedAt.Format(defaultTimeLayout),
		Transaction:             tx,
	}
}

func convertChatMessage(msg *entity.ChatMessage, author model.User, reactions []model.ChatReactionState) model.ChatMessage {
	if msg == nil {
		return model.ChatMessage{}
	}

	if author.ID == "" {
		author.ID = msg.AuthorID
	}

	return model.ChatMessage{
		ID:          msg.ID,
		ChannelID:   msg.ChannelID,
		Author:      author,
		Content:     msg.Content,
		ReplyTo:     msg.ReplyTo,
		Attachments: msg.Attachments,
		Reactions:   reactions,
	}
}

func convertChatChannel(channel *entity.ChatChannel, communityHandle string) model.ChatChannel {
	if channel == nil {
		return model.ChatChannel{}
	}

	return model.ChatChannel{
		ID:              channel.ID,
		UpdatedAt:       channel.UpdatedAt.Format(defaultTimeLayout),
		CommunityHandle: communityHandle,
		Name:            channel.Name,
		LastMessageID:   channel.LastMessageID,
	}
}

func convertLotteryEvent(
	event *entity.LotteryEvent, community model.Community, prizes []model.LotteryPrize,
) model.LotteryEvent {
	if event == nil {
		return model.LotteryEvent{}
	}

	return model.LotteryEvent{
		ID:             event.ID,
		Community:      community,
		StartTime:      event.StartTime.Format(defaultTimeLayout),
		EndTime:        event.EndTime.Format(defaultTimeLayout),
		MaxTickets:     event.MaxTickets,
		UsedTickets:    event.UsedTickets,
		PointPerTicket: int(event.PointPerTicket),
		Prizes:         prizes,
	}
}

func convertLotteryPrize(prize *entity.LotteryPrize) model.LotteryPrize {
	if prize == nil {
		return model.LotteryPrize{}
	}

	return model.LotteryPrize{
		ID:               prize.ID,
		EventID:          prize.LotteryEventID,
		Points:           prize.Points,
		Rewards:          convertRewards(prize.Rewards),
		AvailableRewards: prize.AvailableRewards,
	}
}

func convertLotteryWinner(
	winner *entity.LotteryWinner, prize model.LotteryPrize, user model.User,
) model.LotteryWinner {
	if winner == nil {
		return model.LotteryWinner{}
	}

	if prize.ID == "" {
		prize.ID = winner.LotteryPrizeID
	}

	if user.ID == "" {
		user.ID = winner.UserID
	}

	return model.LotteryWinner{
		ID:        winner.ID,
		CreatedAt: winner.CreatedAt.Format(defaultTimeLayout),
		Prize:     prize,
		User:      user,
	}
}
