package model

type GetMyPayRewardRequest struct{}

type GetMyPayRewardResponse struct {
	PayRewards []PayReward `json:"pay_rewards"`
}

type GetClaimableRewardsRequest struct{}

type ClaimableTokenInfo struct {
	TokenID      string  `json:"token_id"`
	TokenSymbol  string  `json:"token_symbol"`
	TokenAddress string  `json:"token_address"`
	Chain        string  `json:"chain"`
	Amount       float64 `json:"amount"`
}

type GetClaimableRewardsResponse struct {
	ReferralCommunities  []Community          `json:"referral_communities"`
	LotteryWinners       []LotteryWinner      `json:"lottery_winners"`
	TotalClaimableTokens []ClaimableTokenInfo `json:"total_claimable_tokens"`
}
