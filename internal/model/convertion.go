package model

import (
	"database/sql"
	"strings"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/api/discord"
)

const DefaultTimeLayout string = time.RFC3339Nano
const DefaultDateLayout string = "2006-01-02"
const DefaultMonthLayout string = "01-2006"

func ConvertRewards(entityRewards []entity.Reward) []Reward {
	modelRewards := []Reward{}
	for _, r := range entityRewards {
		modelRewards = append(modelRewards, Reward{Type: string(r.Type), Data: r.Data})
	}
	return modelRewards
}

func ConvertConditions(entityConditions []entity.Condition) []Condition {
	modelConditions := []Condition{}
	for _, r := range entityConditions {
		modelConditions = append(modelConditions, Condition{Type: string(r.Type), Data: r.Data})
	}
	return modelConditions
}

func ConvertUser(
	user *entity.User,
	serviceUsers []entity.OAuth2,
	includeSensitive bool,
	status string,
) User {
	if user == nil {
		return User{}
	}

	serviceMap := map[string]string{}
	for _, u := range serviceUsers {
		var id string
		if u.Service == "twitter" {
			id = u.ServiceUsername
		} else {
			var tag string
			var found bool
			tag, id, found = strings.Cut(u.ServiceUserID, "_")
			if !found || tag != u.Service {
				continue
			}
		}

		serviceMap[u.Service] = id
	}

	if !includeSensitive {
		user.Role = ""
		user.WalletAddress = sql.NullString{Valid: false, String: ""}
		user.IsNewUser = false
	}

	return User{
		ShortUser: ShortUser{
			ID:        user.ID,
			Name:      user.Name,
			AvatarURL: user.ProfilePicture,
			Status:    status,
		},
		WalletAddress: user.WalletAddress.String,
		Role:          string(user.Role),
		ReferralCode:  user.ReferralCode,
		Services:      serviceMap,
		IsNewUser:     user.IsNewUser,
	}
}

func ConvertShortUser(user *entity.User, status string) ShortUser {
	if user == nil {
		return ShortUser{}
	}

	return ShortUser{
		ID:        user.ID,
		Name:      user.Name,
		AvatarURL: user.ProfilePicture,
		Status:    status,
	}
}

func ConvertCategory(category *entity.Category) Category {
	if category == nil {
		return Category{}
	}

	return Category{
		ID:        category.ID,
		Name:      category.Name,
		Position:  category.Position,
		CreatedBy: category.CreatedBy,
		CreatedAt: category.CreatedAt.Format(DefaultTimeLayout),
		UpdatedAt: category.UpdatedAt.Format(DefaultTimeLayout),
	}
}

func ConvertRole(role *entity.Role) Role {
	if role == nil {
		return Role{}
	}

	return Role{
		ID:         role.ID,
		Name:       role.Name,
		Permission: role.Permissions,
		Color:      role.Color,
	}
}

func ConvertCommunity(community *entity.Community, totalQuests int) Community {
	if community == nil {
		return Community{}
	}

	return Community{
		Handle:            community.Handle,
		CreatedAt:         community.CreatedAt.Format(DefaultTimeLayout),
		UpdatedAt:         community.UpdatedAt.Format(DefaultTimeLayout),
		ReferredBy:        community.ReferredBy.String,
		ReferralStatus:    string(community.ReferralStatus),
		CreatedBy:         community.CreatedBy,
		Introduction:      string(community.Introduction),
		DisplayName:       community.DisplayName,
		Twitter:           community.Twitter,
		Discord:           community.Discord,
		DiscordInviteLink: community.DiscordInviteLink,
		Followers:         community.Followers,
		TrendingScore:     community.TrendingScore,
		LogoURL:           community.LogoPicture,
		WebsiteURL:        community.WebsiteURL,
		NumberOfQuests:    totalQuests,
		Status:            string(community.Status),
		// Do not leak owner email. Only superadmin can see the owner email when
		// get list of communities.
	}
}

func ConvertCommunityStats(records []entity.CommunityStats) []CommunityStats {
	result := []CommunityStats{}
	for _, r := range records {
		result = append(result, CommunityStats{
			Date:          r.Date.Format(DefaultDateLayout),
			FollowerCount: r.FollowerCount,
		})
	}

	return result
}

func ConvertBadge(badge *entity.Badge) Badge {
	if badge == nil {
		return Badge{}
	}

	return Badge{
		ID:          badge.ID,
		Name:        badge.Name,
		Level:       badge.Level,
		Description: badge.Description,
		IconURL:     badge.IconURL,
	}
}

func ConvertBadgeDetail(
	badgeDetail *entity.BadgeDetail,
	user ShortUser,
	community Community,
	badge Badge,
) BadgeDetail {
	if badgeDetail == nil {
		return BadgeDetail{}
	}

	if user.ID == "" {
		user = ShortUser{ID: badgeDetail.UserID}
	}

	if badge.ID == "" {
		badge = Badge{ID: badgeDetail.BadgeID}
	}

	return BadgeDetail{
		User:        user,
		Community:   community,
		Badge:       badge,
		WasNotified: badgeDetail.WasNotified,
		CreatedAt:   badgeDetail.CreatedAt.Format(DefaultTimeLayout),
	}
}

func ConvertQuest(quest *entity.Quest, community Community, category Category) Quest {
	if quest == nil {
		return Quest{}
	}

	if category.ID == "" {
		category = Category{ID: quest.CategoryID.String}
	}

	return Quest{
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
		Rewards:        ConvertRewards(quest.Rewards),
		ConditionOp:    string(quest.ConditionOp),
		Conditions:     ConvertConditions(quest.Conditions),
		CreatedAt:      quest.CreatedAt.Format(DefaultTimeLayout),
		UpdatedAt:      quest.UpdatedAt.Format(DefaultTimeLayout),
		IsHighlight:    quest.IsHighlight,
		Position:       quest.Position,
	}
}

func ConvertClaimedQuest(
	claimedQuest *entity.ClaimedQuest, quest Quest, user ShortUser,
) ClaimedQuest {
	if claimedQuest == nil {
		return ClaimedQuest{}
	}

	if quest.ID == "" {
		quest = Quest{ID: claimedQuest.QuestID}
	}

	if user.ID == "" {
		user = ShortUser{ID: claimedQuest.UserID}
	}

	reviewedAt := ""
	if claimedQuest.ReviewedAt.Valid {
		reviewedAt = claimedQuest.ReviewedAt.Time.Format(DefaultTimeLayout)
	}

	return ClaimedQuest{
		ID:             claimedQuest.ID,
		Quest:          quest,
		User:           user,
		SubmissionData: claimedQuest.SubmissionData,
		Status:         string(claimedQuest.Status),
		ReviewerID:     claimedQuest.ReviewerID,
		ReviewedAt:     reviewedAt,
		Comment:        claimedQuest.Comment,
		CreatedAt:      claimedQuest.CreatedAt.Format(DefaultTimeLayout),
		UpdatedAt:      claimedQuest.UpdatedAt.Format(DefaultTimeLayout),
	}
}

func ConvertFollower(
	follower *entity.Follower, roles []Role, user ShortUser, community Community,
) Follower {
	if follower == nil {
		return Follower{}
	}

	if user.ID == "" {
		user = ShortUser{ID: follower.UserID}
	}

	return Follower{
		User:        user,
		Community:   community,
		Roles:       roles,
		Points:      follower.Points,
		Quests:      follower.Quests,
		InviteCode:  follower.InviteCode,
		InvitedBy:   follower.InvitedBy.String,
		InviteCount: follower.InviteCount,
		ChatLevel:   follower.ChatLevel,
	}
}

func ConvertFollowerStreak(streaks []entity.FollowerStreak) []FollowerStreak {
	result := []FollowerStreak{}
	for _, s := range streaks {
		result = append(result, FollowerStreak{
			StartTime: s.StartTime.Format(DefaultDateLayout),
			Streaks:   s.Streaks,
		})
	}

	return result
}

func ConvertDiscordRole(role discord.Role) DiscordRole {
	return DiscordRole{
		ID:       role.ID,
		Name:     role.Name,
		Position: role.Position,
	}
}

func ConvertBlockchainConnection(c *entity.BlockchainConnection) BlockchainConnection {
	if c == nil {
		return BlockchainConnection{}
	}

	return BlockchainConnection{
		Type: string(c.Type),
		URL:  c.URL,
	}
}

func ConvertBlockchain(
	b *entity.Blockchain,
	connections []BlockchainConnection,
	tokens []BlockchainToken,
) Blockchain {
	if b == nil {
		return Blockchain{}
	}

	return Blockchain{
		Name:                 b.Name,
		ID:                   b.ID,
		DisplayName:          b.DisplayName,
		UseExternalRPC:       b.UseExternalRPC,
		UseEip1559:           b.UseEip1559,
		BlockTime:            b.BlockTime,
		AdjustTime:           b.AdjustTime,
		ThresholdUpdateBlock: b.ThresholdUpdateBlock,
		CurrencySymbol:       b.CurrencySymbol,
		ExplorerURL:          b.ExplorerURL,
		Connections:          connections,
		Tokens:               tokens,
	}
}

func ConvertBlockchainTransaction(tx *entity.BlockchainTransaction) BlockchainTransaction {
	if tx == nil {
		return BlockchainTransaction{}
	}

	return BlockchainTransaction{
		TxHash:    tx.TxHash,
		Chain:     tx.Chain,
		Status:    string(tx.Status),
		CreatedAt: tx.CreatedAt.Format(DefaultTimeLayout),
		UpdatedAt: tx.UpdatedAt.Format(DefaultTimeLayout),
	}
}

func ConvertBlockchainToken(token *entity.BlockchainToken) BlockchainToken {
	if token == nil {
		return BlockchainToken{}
	}

	return BlockchainToken{
		ID:      token.ID,
		Name:    token.Name,
		Symbol:  token.Symbol,
		Chain:   token.Chain,
		Address: token.Address,
	}
}

func ConvertPayReward(
	pw *entity.PayReward,
	token BlockchainToken,
	toUser ShortUser,
	referralCommunityHandle string,
	fromCommunityHandle string,
	tx BlockchainTransaction,
) PayReward {
	if pw == nil {
		return PayReward{}
	}

	if toUser.ID == "" {
		toUser = ShortUser{ID: pw.ToUserID}
	}

	return PayReward{
		ID:                      pw.ID,
		Token:                   token,
		ToUser:                  toUser,
		ClaimedQuestID:          pw.ClaimedQuestID.String,
		ReferralCommunityHandle: referralCommunityHandle,
		FromCommunityHandle:     fromCommunityHandle,
		ToAddress:               pw.ToAddress,
		Amount:                  pw.Amount,
		CreatedAt:               pw.CreatedAt.Format(DefaultTimeLayout),
		UpdatedAt:               pw.UpdatedAt.Format(DefaultTimeLayout),
		Transaction:             tx,
	}
}

func ConvertChatMessage(msg *entity.ChatMessage, author ShortUser, reactions []ChatReactionState) ChatMessage {
	if msg == nil {
		return ChatMessage{}
	}

	if author.ID == "" {
		author.ID = msg.AuthorID
	}

	return ChatMessage{
		ID:          msg.ID,
		ChannelID:   msg.ChannelID,
		Author:      author,
		Content:     msg.Content,
		ReplyTo:     msg.ReplyTo,
		Attachments: msg.Attachments,
		Reactions:   reactions,
	}
}

func ConvertChatChannel(channel *entity.ChatChannel, communityHandle string) ChatChannel {
	if channel == nil {
		return ChatChannel{}
	}

	return ChatChannel{
		ID:              channel.ID,
		UpdatedAt:       channel.UpdatedAt.Format(DefaultTimeLayout),
		CommunityHandle: communityHandle,
		Name:            channel.Name,
		LastMessageID:   channel.LastMessageID,
		Description:     channel.Description,
	}
}

func ConvertChatMember(member *entity.ChatMember, channel ChatChannel) ChatMember {
	if member == nil {
		return ChatMember{}
	}

	if channel.ID == 0 {
		channel.ID = member.ChannelID
	}

	return ChatMember{
		UserID:            member.UserID,
		Channel:           channel,
		LastReadMessageID: member.LastReadMessageID,
	}
}

func ConvertLotteryEvent(
	event *entity.LotteryEvent, community Community, prizes []LotteryPrize,
) LotteryEvent {
	if event == nil {
		return LotteryEvent{}
	}

	return LotteryEvent{
		ID:             event.ID,
		Community:      community,
		StartTime:      event.StartTime.Format(DefaultTimeLayout),
		EndTime:        event.EndTime.Format(DefaultTimeLayout),
		MaxTickets:     event.MaxTickets,
		UsedTickets:    event.UsedTickets,
		PointPerTicket: int(event.PointPerTicket),
		Prizes:         prizes,
	}
}

func ConvertLotteryPrize(prize *entity.LotteryPrize) LotteryPrize {
	if prize == nil {
		return LotteryPrize{}
	}

	return LotteryPrize{
		ID:               prize.ID,
		EventID:          prize.LotteryEventID,
		Points:           prize.Points,
		Rewards:          ConvertRewards(prize.Rewards),
		AvailableRewards: prize.AvailableRewards,
	}
}

func ConvertLotteryWinner(
	winner *entity.LotteryWinner, prize LotteryPrize, user ShortUser,
) LotteryWinner {
	if winner == nil {
		return LotteryWinner{}
	}

	if prize.ID == "" {
		prize.ID = winner.LotteryPrizeID
	}

	if user.ID == "" {
		user.ID = winner.UserID
	}

	return LotteryWinner{
		ID:        winner.ID,
		CreatedAt: winner.CreatedAt.Format(DefaultTimeLayout),
		Prize:     prize,
		User:      user,
	}
}
