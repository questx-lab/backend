package model

type GetMyPayRewardRequest struct{}

type GetMyPayRewardResponse struct {
	PayRewards []PayReward
}

type PayRewardTxRequest struct {
	PayRewardID string `json:"pay_reward_id"`
	Chain       string `json:"chain"`
}
