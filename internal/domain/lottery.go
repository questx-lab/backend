package domain

import (
	"context"
	"errors"
	"math"
	"math/big"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/fatih/structs"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/ethutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type LotteryDomain interface {
	CreateLotteryEvent(context.Context, *model.CreateLotteryEventRequest) (*model.CreateLotteryEventResponse, error)
	GetLotteryEvent(context.Context, *model.GetLotteryEventRequest) (*model.GetLotteryEventResponse, error)
	BuyTicket(context.Context, *model.BuyLotteryTicketsRequest) (*model.BuyLotteryTicketsResponse, error)
	Claim(context.Context, *model.ClaimLotteryWinnerRequest) (*model.ClaimLotteryWinnerResponse, error)
}

type lotteryDomain struct {
	lotteryRepo           repository.LotteryRepository
	followerRepo          repository.FollowerRepository
	communityRepo         repository.CommunityRepository
	blockchainRepo        repository.BlockChainRepository
	communityRoleVerifier *common.CommunityRoleVerifier
	questFactory          questclaim.Factory
	blockchainCaller      client.BlockchainCaller
}

func NewLotteryDomain(
	lotteryRepo repository.LotteryRepository,
	followerRepo repository.FollowerRepository,
	communityRepo repository.CommunityRepository,
	blockchainRepo repository.BlockChainRepository,
	communityRoleVerifier *common.CommunityRoleVerifier,
	questFactory questclaim.Factory,
	blockchainCaller client.BlockchainCaller,
) *lotteryDomain {
	return &lotteryDomain{
		lotteryRepo:           lotteryRepo,
		followerRepo:          followerRepo,
		communityRepo:         communityRepo,
		blockchainRepo:        blockchainRepo,
		communityRoleVerifier: communityRoleVerifier,
		questFactory:          questFactory,
		blockchainCaller:      blockchainCaller,
	}
}

func (d *lotteryDomain) CreateLotteryEvent(
	ctx context.Context, req *model.CreateLotteryEventRequest,
) (*model.CreateLotteryEventResponse, error) {
	if req.StartTime.After(req.EndTime) {
		return nil, errorx.New(errorx.BadRequest, "Invalid event time")
	}

	if req.EndTime.Before(time.Now()) {
		return nil, errorx.New(errorx.BadRequest, "End time of event must after now")
	}

	if req.MaxTickets <= 0 {
		return nil, errorx.New(errorx.BadRequest, "The max number of tickets must be a positive number")
	}

	if req.PointPerTicket == 0 {
		return nil, errorx.New(errorx.BadRequest, "Not allow free ticket")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	totalPrizes := 0
	eventPrizes := []*entity.LotteryPrize{}
	totalTokens := map[string]map[string]float64{} // chain - token id - amount
	for i, prize := range req.Prizes {
		if prize.AvailableRewards <= 0 {
			return nil, errorx.New(errorx.BadRequest,
				"The number of available rewards %d must be a positive number", i+1)
		}

		totalPrizes += prize.AvailableRewards
		eventPrize := &entity.LotteryPrize{
			Base:             entity.Base{ID: uuid.NewString()},
			Points:           prize.Points,
			Rewards:          []entity.Reward{},
			AvailableRewards: prize.AvailableRewards,
		}

		for _, r := range prize.Rewards {
			rType, err := enum.ToEnum[entity.RewardType](r.Type)
			if err != nil {
				xcontext.Logger(ctx).Debugf("Invalid reward type: %v", err)
				continue
			}

			reward, err := d.questFactory.NewReward(ctx, community.ID, rType, r.Data)
			if err != nil {
				return nil, err
			}

			eventPrize.Rewards = append(eventPrize.Rewards, entity.Reward{Type: rType, Data: structs.Map(reward)})

			// Calculate token amount and check the balance of community after.
			if rType == entity.CoinReward {
				coinReward, ok := reward.(*questclaim.CoinReward)
				if !ok {
					xcontext.Logger(ctx).Errorf("Cannot cast coin reward")
					return nil, errorx.Unknown
				}

				if _, ok := totalTokens[coinReward.Chain]; !ok {
					totalTokens[coinReward.Chain] = make(map[string]float64)
				}

				totalTokens[coinReward.Chain][coinReward.TokenID] += coinReward.Amount
			}
		}

		if len(eventPrize.Rewards) == 0 && eventPrize.Points == 0 {
			return nil, errorx.New(errorx.BadRequest, "Require at least one reward for prize %d", i)
		}

		eventPrizes = append(eventPrizes, eventPrize)
	}

	if totalPrizes > req.MaxTickets {
		return nil, errorx.New(errorx.BadRequest,
			"Total available rewards must less than or equal to max tickets")
	}

	communityPrivateKey, err := ethutil.GeneratePrivateKey(
		[]byte(xcontext.Configs(ctx).Blockchain.SecretKey), []byte(community.WalletNonce))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot generate private key: %v", err)
		return nil, errorx.Unknown
	}

	communityAddress := ethcrypto.PubkeyToAddress(communityPrivateKey.PublicKey).String()
	for chain, tokenMap := range totalTokens {
		for tokenID, amount := range tokenMap {
			token, err := d.blockchainRepo.GetTokenByID(ctx, tokenID)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get token %s: %v", tokenID, err)
				return nil, errorx.Unknown
			}

			balance, err := d.blockchainCaller.ERC20BalanceOf(ctx, chain, token.Address, communityAddress)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get balance %s-%s of %s: %v",
					chain, tokenID, community.Handle, err)
				return nil, errorx.Unknown
			}

			bigAmount := big.NewInt(int64(amount * math.Pow10(token.Decimals)))
			if bigAmount.Cmp(balance) == 1 {
				balanceFloat, _ := new(big.Float).Quo(
					new(big.Float).SetInt(balance), big.NewFloat(math.Pow10(token.Decimals))).Float64()

				return nil, errorx.New(errorx.Unavailable,
					"Your current balance is %.2f%s, please add another %.2f%s to community wallet at chain %s",
					balanceFloat, token.Symbol, amount-balanceFloat, token.Symbol, token.Chain)
			}
		}
	}

	if err := d.communityRoleVerifier.Verify(ctx, community.ID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	lastEvent, err := d.lotteryRepo.GetLastEventByCommunityID(ctx, community.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		xcontext.Logger(ctx).Errorf("Cannot get the last event: %v", err)
		return nil, errorx.Unknown
	}

	if err == nil {
		if lastEvent.EndTime.After(time.Now()) && lastEvent.UsedTickets < lastEvent.MaxTickets {
			return nil, errorx.New(errorx.Unavailable, "Still have an incompleted lottery event")
		}

		if !req.StartTime.After(lastEvent.StartTime) {
			return nil, errorx.New(errorx.BadRequest,
				"Start time of this event must be after of the previous event")
		}
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	event := &entity.LotteryEvent{
		Base:           entity.Base{ID: uuid.NewString()},
		CommunityID:    community.ID,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		MaxTickets:     req.MaxTickets,
		UsedTickets:    0,
		PointPerTicket: req.PointPerTicket,
	}

	if err := d.lotteryRepo.CreateEvent(ctx, event); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create event: %v", err)
		return nil, errorx.Unknown
	}

	for i := range eventPrizes {
		eventPrize := eventPrizes[i]
		eventPrize.LotteryEventID = event.ID

		if err := d.lotteryRepo.CreatePrize(ctx, eventPrize); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot create lottery prize: %v", err)
			return nil, errorx.Unknown
		}
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.CreateLotteryEventResponse{}, nil
}

func (d *lotteryDomain) GetLotteryEvent(
	ctx context.Context, req *model.GetLotteryEventRequest,
) (*model.GetLotteryEventResponse, error) {
	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	event, err := d.lotteryRepo.GetLastEventByCommunityID(ctx, community.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found any lottery event")
		}

		xcontext.Logger(ctx).Errorf("Cannot get the last event: %v", err)
		return nil, errorx.Unknown
	}

	prizes, err := d.lotteryRepo.GetPrizesByEventID(ctx, event.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get the last event: %v", err)
		return nil, errorx.Unknown
	}

	clientPrizes := []model.LotteryPrize{}
	for _, prize := range prizes {
		clientPrizes = append(clientPrizes, model.ConvertLotteryPrize(&prize))
	}

	return &model.GetLotteryEventResponse{
		Event: model.ConvertLotteryEvent(event, model.ConvertCommunity(community, 0), clientPrizes),
	}, nil
}

func (d *lotteryDomain) BuyTicket(
	ctx context.Context, req *model.BuyLotteryTicketsRequest,
) (*model.BuyLotteryTicketsResponse, error) {
	if req.NumberTickets <= 0 {
		return nil, errorx.New(errorx.BadRequest, "Number of tickets must be a positve number")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	event, err := d.lotteryRepo.GetLastEventByCommunityID(ctx, community.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found any lottery event")
		}

		xcontext.Logger(ctx).Errorf("Cannot get the last event: %v", err)
		return nil, errorx.Unknown
	}

	if !event.EndTime.After(time.Now()) {
		return nil, errorx.New(errorx.NotFound, "The event has ended")
	}

	if event.UsedTickets >= event.MaxTickets {
		return nil, errorx.New(errorx.Unavailable, "Out of tickets")
	}

	userID := xcontext.RequestUserID(ctx)
	results := []model.LotteryWinner{}
	doneTickets := 0
	var stopErr error = errors.New("")
	for doneTickets < req.NumberTickets {
		stopReason, err := func() (string, error) {
			prizes, err := d.lotteryRepo.GetPrizesByEventID(ctx, event.ID)
			if err != nil {
				return "", err
			}

			currentEventInfo, err := d.lotteryRepo.GetEventByID(ctx, event.ID)
			if err != nil {
				return "", err
			}

			ctx = xcontext.WithDBTransaction(ctx)
			defer func() {
				ctx = xcontext.WithRollbackDBTransaction(ctx)
			}()

			if err := d.lotteryRepo.CheckAndUseEventTicket(ctx, event.ID); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return "No more tickets", nil
				}

				return "", err
			}

			if event.PointPerTicket > 0 {
				err = d.followerRepo.DecreasePoint(ctx, userID, community.ID,
					event.PointPerTicket, false)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return "Not enough point", nil
					}

					return "", err
				}
			}

			wonPrize := d.spin(currentEventInfo.UsedTickets, currentEventInfo.MaxTickets, prizes)

			winner := entity.LotteryWinner{}
			if wonPrize != nil {
				if err := d.lotteryRepo.CheckAndWinEventPrize(ctx, wonPrize.ID); err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return "", nil // this prize is out of stock, continue spin.
					}

					return "", err
				}

				winner = entity.LotteryWinner{
					Base:           entity.Base{ID: uuid.NewString()},
					LotteryPrizeID: wonPrize.ID,
					UserID:         xcontext.RequestUserID(ctx),
					IsClaimed:      false,
				}

				if err := d.lotteryRepo.CreateWinner(ctx, &winner); err != nil {
					return "", err
				}

				results = append(results, model.ConvertLotteryWinner(
					&winner, model.ConvertLotteryPrize(wonPrize), model.ConvertShortUser(nil, "")))
			}

			doneTickets++
			ctx = xcontext.WithCommitDBTransaction(ctx)
			return "", nil
		}()

		if err != nil {
			if len(results) == 0 {
				xcontext.Logger(ctx).Errorf("Cannot find the won prize: %v", err)
				return nil, errorx.Unknown
			}

			stopErr = err
			break
		}

		if stopReason != "" {
			stopErr = errors.New(stopReason)
			break
		}
	}

	return &model.BuyLotteryTicketsResponse{Results: results, Error: stopErr.Error()}, nil
}

func (d *lotteryDomain) Claim(
	ctx context.Context, req *model.ClaimLotteryWinnerRequest,
) (*model.ClaimLotteryWinnerResponse, error) {
	if len(req.WinnerIDs) == 0 {
		return nil, errorx.New(errorx.BadRequest, "Require at least one winner id")
	}

	for _, winnerID := range req.WinnerIDs {
		winner, err := d.lotteryRepo.GetWinnerByID(ctx, winnerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found winner record")
			}

			xcontext.Logger(ctx).Errorf("Cannot get winner: %v", err)
			return nil, errorx.Unknown
		}

		if winner.IsClaimed {
			return nil, errorx.New(errorx.Unavailable, "User claimed this reward before")
		}

		userID := xcontext.RequestUserID(ctx)
		if userID != winner.UserID {
			return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
		}

		prize, err := d.lotteryRepo.GetPrizeByID(ctx, winner.LotteryPrizeID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get prize: %v", err)
			return nil, errorx.Unknown
		}

		event, err := d.lotteryRepo.GetEventByID(ctx, prize.LotteryEventID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get event: %v", err)
			return nil, errorx.Unknown
		}

		ctx = xcontext.WithDBTransaction(ctx)
		defer xcontext.WithRollbackDBTransaction(ctx)

		if prize.Points > 0 {
			err = d.followerRepo.IncreasePoint(ctx, userID, event.CommunityID, uint64(prize.Points), false)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot increase point: %v", err)
				return nil, errorx.Unknown
			}
		}

		if err := d.lotteryRepo.ClaimWinnerReward(ctx, winnerID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.Unavailable, "User claimed this reward before")
			}

			xcontext.Logger(ctx).Errorf("Cannot claim winner reward: %v", err)
			return nil, errorx.Unknown
		}

		for _, r := range prize.Rewards {
			reward, err := d.questFactory.LoadReward(ctx, event.CommunityID, r.Type, r.Data)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot load reward: %v", err)
				return nil, errorx.Unknown
			}

			reward.WithLotteryWinner(winner)
			reward.WithWalletAddress(req.WalletAddress)
			if err := reward.Give(ctx); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot give reward: %v", err)
				return nil, errorx.Unknown
			}
		}

		ctx = xcontext.WithCommitDBTransaction(ctx)
	}

	return &model.ClaimLotteryWinnerResponse{}, nil
}

func (d *lotteryDomain) spin(
	usedTickets, maxTickets int, prizes []entity.LotteryPrize,
) *entity.LotteryPrize {
	randomIndex := crypto.RandIntn(maxTickets - usedTickets)
	for _, prize := range prizes {
		remainPrize := prize.AvailableRewards - prize.WonRewards
		if remainPrize == 0 {
			continue
		}

		if randomIndex < remainPrize {
			return &prize
		}

		randomIndex -= remainPrize
	}

	return nil
}
