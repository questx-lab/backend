package blockchain

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	icommon "github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/blockchain/eth"
	"github.com/questx-lab/backend/internal/domain/blockchain/types"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/ethutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
)

type CombinedTransactionKey struct {
	FromWalletNonce string
	ToAddress       string
	Chain           string
	TokenAddress    string
}

type CombinedTransactionValue struct {
	Amount       float64
	PayRewardIDs []string
}

type BlockchainManager struct {
	rootCtx        context.Context
	payRewardRepo  repository.PayRewardRepository
	blockchainRepo repository.BlockChainRepository
	communityRepo  repository.CommunityRepository
	dispatchers    map[string]Dispatcher
	watchers       map[string]Watcher
	ethClients     map[string]eth.EthClient
	redisClient    xredis.Client
}

func NewBlockchainManager(
	ctx context.Context,
	payRewardRepo repository.PayRewardRepository,
	communityRepo repository.CommunityRepository,
	blockchainRepo repository.BlockChainRepository,
	redisClient xredis.Client,
) *BlockchainManager {
	return &BlockchainManager{
		rootCtx:        ctx,
		blockchainRepo: blockchainRepo,
		payRewardRepo:  payRewardRepo,
		communityRepo:  communityRepo,
		dispatchers:    make(map[string]Dispatcher),
		watchers:       make(map[string]Watcher),
		ethClients:     make(map[string]eth.EthClient),
		redisClient:    redisClient,
	}
}

func (m *BlockchainManager) Run(ctx context.Context) {
	for {
		m.reloadChains(ctx)
		m.handlePendingPayrewards(ctx)

		time.Sleep(30 * time.Second)
	}
}

func (m *BlockchainManager) GetTokenInfo(_ context.Context, chain, address string) (types.TokenInfo, error) {
	client, ok := m.ethClients[chain]
	if !ok {
		return types.TokenInfo{}, fmt.Errorf("unsupported chain %s", chain)
	}

	return client.GetTokenInfo(m.rootCtx, address)
}

func (m *BlockchainManager) ERC20BalanceOf(
	_ context.Context, chain, tokenAddress, accountAddress string,
) (*big.Int, error) {
	client, ok := m.ethClients[chain]
	if !ok {
		return nil, fmt.Errorf("unsupported chain %s", chain)
	}

	return client.ERC20BalanceOf(m.rootCtx, tokenAddress, accountAddress)
}

func (m *BlockchainManager) reloadChains(ctx context.Context) {
	allChains, err := m.blockchainRepo.GetAll(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot load all chains: %v", err)
		return
	}

	for _, chain := range allChains {
		if _, ok := m.ethClients[chain.Name]; !ok {
			m.addChain(ctx, &chain)
		}
	}
}

func (m *BlockchainManager) addChain(ctx context.Context, blockchain *entity.Blockchain) {
	xcontext.Logger(ctx).Infof("Begin supporting chain %s", blockchain.Name)
	client := eth.NewEthClients(blockchain, m.blockchainRepo)
	dispatcher := eth.NewEhtDispatcher(client)
	watcher := eth.NewEthWatcher(ctx, blockchain, m.blockchainRepo, client, m.redisClient)

	m.ethClients[blockchain.Name] = client
	m.dispatchers[blockchain.Name] = dispatcher
	m.watchers[blockchain.Name] = watcher

	go client.Start(ctx)
	go watcher.Start(ctx)
}

func (m *BlockchainManager) handlePendingPayrewards(ctx context.Context) {
	counter := icommon.PromCounters[icommon.BlockchainTransactionFailure]

	allPendingPayRewards, err := m.payRewardRepo.GetAllPending(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all pending pay rewards: %v", err)
		return
	}

	walletNonces := map[string]string{}
	tokenMap := map[string]*entity.BlockchainToken{}
	combinedTransactions := map[CombinedTransactionKey]*CombinedTransactionValue{}
	for _, reward := range allPendingPayRewards {
		// Our platform doesn't need a wallet nonce to generate private/public key.
		var walletNonce string
		if reward.FromCommunityID.Valid {
			if _, ok := walletNonces[reward.FromCommunityID.String]; !ok {
				community, err := m.communityRepo.GetByID(ctx, reward.FromCommunityID.String)
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot get community %s of reward %s: %v",
						reward.FromCommunityID.String, reward.ID, err)
					continue
				}

				if community.WalletNonce == "" {
					nonce, err := crypto.GenerateRandomString()
					if err != nil {
						xcontext.Logger(ctx).Errorf("Cannot generate nonce: %v", err)
						continue
					}

					err = m.communityRepo.UpdateByID(ctx, community.ID, entity.Community{
						WalletNonce: nonce,
					})
					if err != nil {
						xcontext.Logger(ctx).Errorf("Cannot update wallet nonce: %v", err)
						continue
					}

					community.WalletNonce = nonce
				}

				walletNonces[community.ID] = community.WalletNonce
			}

			walletNonce = walletNonces[reward.FromCommunityID.String]
		}

		if _, ok := tokenMap[reward.TokenID]; !ok {
			token, err := m.blockchainRepo.GetTokenByID(ctx, reward.TokenID)
			if err != nil {
				xcontext.Logger(ctx).Warnf("Cannot get token %s of reward %s: %v",
					reward.TokenID, reward.ID, err)
				continue
			}

			tokenMap[token.ID] = token
		}

		key := CombinedTransactionKey{
			FromWalletNonce: walletNonce,
			ToAddress:       reward.ToAddress,
			Chain:           tokenMap[reward.TokenID].Chain,
			TokenAddress:    tokenMap[reward.TokenID].Address,
		}

		if _, ok := combinedTransactions[key]; !ok {
			combinedTransactions[key] = &CombinedTransactionValue{Amount: 0, PayRewardIDs: nil}
		}

		combinedTransactions[key].Amount += reward.Amount
		combinedTransactions[key].PayRewardIDs = append(combinedTransactions[key].PayRewardIDs, reward.ID)
	}

	for key, value := range combinedTransactions {
		dispatchedTxReq, err := m.getDispatchedTxRequest(ctx, key, *value)
		if err != nil {
			counter.WithLabelValues("Cannot get dispatched tx request").Inc()
			xcontext.Logger(ctx).Errorf("Cannot get dispatched tx request: %v", err.Error())
			return
		}

		// update to database
		xcontext.Logger(ctx).Infof("Process transaction with hash %s", dispatchedTxReq.TxHash)

		func() {
			ctx = xcontext.WithDBTransaction(ctx)
			defer xcontext.WithRollbackDBTransaction(ctx)

			bcTx := &entity.BlockchainTransaction{
				Base:   entity.Base{ID: uuid.NewString()},
				Status: entity.BlockchainTransactionStatusTypeInProgress,
				Chain:  key.Chain,
				TxHash: dispatchedTxReq.TxHash,
			}

			if err := m.blockchainRepo.CreateTransaction(ctx, bcTx); err != nil {
				counter.WithLabelValues("Unable to create blockchain transaction to database").Inc()
				xcontext.Logger(ctx).Errorf("Unable to create blockchain transaction to database")
				return
			}

			txID := sql.NullString{Valid: true, String: bcTx.ID}
			if err := m.payRewardRepo.UpdateTransactionByIDs(ctx, value.PayRewardIDs, txID); err != nil {
				counter.WithLabelValues("Cannot update payreward blockchain transaction").Inc()
				xcontext.Logger(ctx).Errorf("Cannot update payreward blockchain transaction: %v", err)
				return
			}

			dispatcher, ok := m.dispatchers[key.Chain]
			if !ok {
				counter.WithLabelValues("Dispatcher not exists").Inc()
				xcontext.Logger(ctx).Errorf("Dispatcher not exists")
				return
			}

			result := dispatcher.Dispatch(ctx, dispatchedTxReq)
			if result.Err != types.ErrNil {
				counter.WithLabelValues("Unable to dispatch").Inc()
				xcontext.Logger(ctx).Errorf("Unable to dispatch: %v", result.Err)
				return
			}

			watcher, ok := m.watchers[key.Chain]
			if !ok {
				counter.WithLabelValues("Watcher not exists").Inc()
				xcontext.Logger(ctx).Errorf("Watcher not exists")
				return
			}
			watcher.TrackTx(ctx, dispatchedTxReq.TxHash)

			xcontext.WithCommitDBTransaction(ctx)
		}()
	}
}

func (m *BlockchainManager) getDispatchedTxRequest(
	ctx context.Context,
	transactionKey CombinedTransactionKey,
	transactionValue CombinedTransactionValue,
) (*types.DispatchedTxRequest, error) {
	secret := xcontext.Configs(ctx).Blockchain.SecretKey
	privateKey, err := ethutil.GeneratePrivateKey([]byte(secret), []byte(transactionKey.FromWalletNonce))
	if err != nil {
		return nil, err
	}

	toAddress := common.HexToAddress(transactionKey.ToAddress)
	client, ok := m.ethClients[transactionKey.Chain]
	if !ok {
		return nil, fmt.Errorf("not support chain %s", transactionKey.Chain)
	}

	token, err := m.blockchainRepo.GetToken(ctx, transactionKey.Chain, transactionKey.TokenAddress)
	if err != nil {
		return nil, err
	}

	tx, err := client.GetSignedTransaction(
		ctx,
		token,
		privateKey,
		toAddress,
		transactionValue.Amount,
	)
	if err != nil {
		return nil, err
	}

	b, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &types.DispatchedTxRequest{
		Chain:  transactionKey.Chain,
		Tx:     b,
		TxHash: tx.Hash().Hex(),
		PubKey: ethcrypto.FromECDSAPub(&privateKey.PublicKey),
	}, nil
}
