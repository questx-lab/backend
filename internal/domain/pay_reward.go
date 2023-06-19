package domain

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/blockchain/eth"
	interfaze "github.com/questx-lab/backend/pkg/blockchain/interface"
	"github.com/questx-lab/backend/pkg/blockchain/types"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type PayRewardDomain interface {
	GetMyPayRewards(context.Context, *model.GetMyPayRewardRequest) (*model.GetMyPayRewardResponse, error)
	Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time)
}

type payRewardDomain struct {
	payRewardRepo repository.PayRewardRepository
	cfg           config.EthConfigs
	dispatchers   *xsync.MapOf[string, interfaze.Dispatcher]
	watchers      *xsync.MapOf[string, interfaze.Watcher]
	ethClients    *xsync.MapOf[string, eth.EthClient]
}

func NewPayRewardDomain(
	payRewardRepo repository.PayRewardRepository,
	cfg config.EthConfigs,
	dispatchers *xsync.MapOf[string, interfaze.Dispatcher],
	watchers *xsync.MapOf[string, interfaze.Watcher],
	ethClients *xsync.MapOf[string, eth.EthClient],
) *payRewardDomain {
	return &payRewardDomain{
		payRewardRepo: payRewardRepo,
		cfg:           cfg,
		dispatchers:   dispatchers,
		watchers:      watchers,
		ethClients:    ethClients,
	}
}

func (d *payRewardDomain) GetMyPayRewards(
	ctx context.Context, req *model.GetMyPayRewardRequest,
) (*model.GetMyPayRewardResponse, error) {
	txs, err := d.payRewardRepo.GetByUserID(ctx, xcontext.RequestUserID(ctx))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get payReward by user id: %v", err)
		return nil, errorx.Unknown
	}

	clientTxs := []model.PayReward{}
	for _, tx := range txs {
		clientTxs = append(clientTxs, model.PayReward{
			ID:        tx.ID,
			CreatedAt: tx.CreatedAt.Format(time.RFC3339Nano),
			Note:      tx.Note,
			Address:   tx.Address,
			Token:     tx.Token,
			Amount:    tx.Amount,
		})
	}

	return &model.GetMyPayRewardResponse{PayRewards: clientTxs}, nil
}

func (d *payRewardDomain) getDispatchedTxRequest(ctx context.Context, p *entity.PayReward, txReq *model.PayRewardTxRequest) (*types.DispatchedTxRequest, error) {
	cfg := xcontext.Configs(ctx)
	publicKeyBytes, err := hex.DecodeString(cfg.Eth.Keys.PubKey)
	if err != nil {
		return nil, fmt.Errorf("unable to decode public key")
	}

	pubKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse public key")
	}
	fromAddress := crypto.PubkeyToAddress(*pubKey)
	toAddress := common.HexToAddress(p.Address)
	client, ok := d.ethClients.Load(txReq.Chain)
	if !ok {
		return nil, fmt.Errorf("chain %s doesn't have config", txReq.Chain)
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := client.GetSignedTransaction(ctx, fromAddress, toAddress, big.NewInt(int64(p.Amount)), gasPrice)
	if err != nil {
		return nil, err
	}

	return &types.DispatchedTxRequest{
		Chain:  txReq.Chain,
		Tx:     tx.Hash().Bytes(),
		TxHash: tx.Hash().Hex(),
		PubKey: crypto.FromECDSAPub(pubKey),
	}, nil
}

func (d *payRewardDomain) Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time) {
	var tx model.PayRewardTxRequest
	if err := json.Unmarshal(pack.Msg, &tx); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to unmarshal transaction: %v", err.Error())
		return
	}
	payReward, err := d.payRewardRepo.GetByID(ctx, tx.PayRewardID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get pay reward: %v", err.Error())
		return
	}
	dispatchedTxReq, err := d.getDispatchedTxRequest(ctx, payReward, &tx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("cannot get dispatched tx request: %v", err.Error())
		return
	}
	dispatcher, ok := d.dispatchers.Load(tx.Chain)
	if !ok {
		xcontext.Logger(ctx).Errorf("dispatcher not exists")
		return
	}
	result := dispatcher.Dispatch(ctx, dispatchedTxReq)
	if result.Err != types.ErrNil {
		xcontext.Logger(ctx).Errorf("Unable to dispatch")
		return
	}
	watcher, ok := d.watchers.Load(tx.Chain)

	if !ok {
		xcontext.Logger(ctx).Errorf("watcher not exists")
		return
	}

	watcher.TrackTx(ctx, dispatchedTxReq.TxHash)
}
