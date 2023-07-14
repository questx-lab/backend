package model

type GetMyPayRewardRequest struct{}

type GetMyPayRewardResponse struct {
	PayRewards []PayReward `json:"pay_rewards"`
}
