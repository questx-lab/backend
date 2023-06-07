package model

type GetMyPayRewardRequest struct{}

type GetMyPayRewardResponse struct {
	PayRewards []PayReward `pay_rewards`
}
