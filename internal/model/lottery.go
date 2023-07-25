package model

import "time"

type CreateLotteryEventRequest struct {
	CommunityHandle string    `json:"community_handle"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	MaxTickets      int       `json:"max_tickets"`
	PointPerTicket  uint64    `json:"point_per_ticket"`
	Prizes          []struct {
		Points           int      `json:"points"`
		Rewards          []Reward `json:"rewards"`
		AvailableRewards int      `json:"available_rewards"`
	} `json:"prizes"`
}

type CreateLotteryEventResponse struct{}

type GetLotteryEventRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetLotteryEventResponse struct {
	Event LotteryEvent `json:"event"`
}

type BuyLotteryTicketsRequest struct {
	CommunityHandle string `json:"community_handle"`
	NumberTickets   int    `json:"number_tickets"`
}

type BuyLotteryTicketsResponse struct {
	Results []LotteryWinner `json:"results"`
	Error   string          `json:"error"`
}

type ClaimLotteryWinnerRequest struct {
	WinnerIDs     []string `json:"winner_ids"`
	WalletAddress string   `json:"wallet_address"`
}

type ClaimLotteryWinnerResponse struct{}
