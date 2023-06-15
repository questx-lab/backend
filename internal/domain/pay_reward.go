package domain

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	dispatchers   map[string]interfaze.Dispatcher
	watchers      map[string]interfaze.Watcher
	ethClients    map[string]eth.EthClient
}

func NewPayRewardDomain(
	payRewardRepo repository.PayRewardRepository,
	cfg config.EthConfigs,
	dispatchers map[string]interfaze.Dispatcher,
	watchers map[string]interfaze.Watcher,
	ethClients map[string]eth.EthClient,
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
	privateKey, err := crypto.HexToECDSA(d.cfg.Keys.PrivKey)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	toAddress := common.HexToAddress(p.Address)

	gasPrice, err := d.ethClients[txReq.Chain].SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := d.ethClients[txReq.Chain].GetSignedTransaction(ctx, privateKey, fromAddress, toAddress, big.NewInt(int64(p.Amount)), gasPrice)
	if err != nil {
		return nil, err
	}

	return &types.DispatchedTxRequest{
		Chain:  txReq.Chain,
		Tx:     tx.Hash().Bytes(),
		TxHash: tx.Hash().Hex(),
		PubKey: crypto.FromECDSAPub(publicKeyECDSA),
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
	result := d.dispatchers[tx.Chain].Dispatch(ctx, dispatchedTxReq)
	if result.Err != types.ErrNil {
		xcontext.Logger(ctx).Errorf("Unable to dispatch")
		return
	}

	d.watchers[tx.Chain].TrackTx(ctx, dispatchedTxReq.TxHash)
}
