package blockchain

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/common"
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
		m.handlePendingPayRewards(ctx)

		time.Sleep(30 * time.Second)
	}
}

func (m *BlockchainManager) ERC20TokenInfo(
	_ context.Context, chain, address string,
) (types.TokenInfo, error) {
	client, ok := m.ethClients[chain]
	if !ok {
		return types.TokenInfo{}, fmt.Errorf("unsupported chain %s", chain)
	}

	return client.ERC20TokenInfo(m.rootCtx, address)
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

func (m *BlockchainManager) ERC1155BalanceOf(
	_ context.Context, chain, address string, tokenID int64,
) (*big.Int, error) {
	client, ok := m.ethClients[chain]
	if !ok {
		return nil, fmt.Errorf("unsupported chain %s", chain)
	}

	return client.ERC1155BalanceOf(m.rootCtx, address, tokenID)
}

func (m *BlockchainManager) ERC1155TokenURI(
	_ context.Context, chain string, tokenID int64,
) (string, error) {
	client, ok := m.ethClients[chain]
	if !ok {
		return "", fmt.Errorf("unsupported chain %s", chain)
	}

	return client.ERC1155TokenURI(m.rootCtx, tokenID)
}

func (m *BlockchainManager) MintNFT(
	_ context.Context, communityID, chain string, nftID int64, amount int,
) (string, error) {
	ctx := m.rootCtx

	client, ok := m.ethClients[chain]
	if !ok {
		return "", fmt.Errorf("not support chain %s", chain)
	}

	community, err := m.communityRepo.GetByID(ctx, communityID)
	if err != nil {
		return "", err
	}

	communityAddress, err := ethutil.GeneratePublicKey(
		[]byte(xcontext.Configs(ctx).Blockchain.SecretKey), []byte(community.WalletNonce))
	if err != nil {
		return "", err
	}

	tx, err := client.GetSignedMintNftTx(ctx, communityAddress, nftID, amount)
	if err != nil {
		return "", err
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	bcTx := &entity.BlockchainTransaction{
		Base:   entity.Base{ID: uuid.NewString()},
		Status: entity.BlockchainTransactionStatusTypeInProgress,
		Chain:  chain,
		TxHash: tx.Hash().Hex(),
	}

	if err := m.blockchainRepo.CreateTransaction(ctx, bcTx); err != nil {
		return "", err
	}

	// Get the dispatcher and dispatch this transaction.
	dispatcher, ok := m.dispatchers[chain]
	if !ok {
		return "", fmt.Errorf("dispatcher %s not exists", chain)
	}

	// Get the watcher and track this transaction status.
	watcher, ok := m.watchers[chain]
	if !ok {
		return "", fmt.Errorf("watcher %s not exists", chain)
	}

	result := dispatcher.Dispatch(ctx, &types.DispatchedTxRequest{Chain: chain, Tx: tx})
	if result.Err != types.ErrNil {
		return "", fmt.Errorf("unable to dispatch: %v", result.Err)
	}

	watcher.TrackTx(ctx, tx.Hash().Hex())
	xcontext.WithCommitDBTransaction(ctx)

	return bcTx.ID, nil
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
	client := eth.NewEthClients(blockchain, m.blockchainRepo, m.redisClient)
	dispatcher := eth.NewEhtDispatcher(client)
	watcher := eth.NewEthWatcher(ctx, blockchain, m.blockchainRepo, client, m.redisClient)

	m.ethClients[blockchain.Name] = client
	m.dispatchers[blockchain.Name] = dispatcher
	m.watchers[blockchain.Name] = watcher

	go client.Start(ctx)
	go watcher.Start(ctx)
}

func (m *BlockchainManager) handlePendingPayRewards(ctx context.Context) {
	combinedTransactions := m.combineTransferTokenTransaction(ctx)
	if combinedTransactions == nil {
		return
	}

	m.dispatchCombinedTransactions(ctx, combinedTransactions)
}

func (m *BlockchainManager) combineTransferTokenTransaction(
	ctx context.Context,
) map[CombinedTransactionKey]*CombinedTransactionValue {
	allPendingPayRewards, err := m.payRewardRepo.GetAllPending(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all pending pay rewards: %v", err)
		return nil
	}

	// Combine all pay rewards in the same chain, token, sender, recipient for
	// one transaction.
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

				// Back-compatible for communities with no nonce value.
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
				xcontext.Logger(ctx).Errorf("Cannot get token %s of reward %s: %v",
					reward.TokenID, reward.ID, err)
				continue
			}

			tokenMap[token.ID] = token
		}

		// The transaction key is combination of from, to, chain, token.
		key := CombinedTransactionKey{
			FromWalletNonce: walletNonce,
			ToAddress:       reward.ToAddress,
			Chain:           tokenMap[reward.TokenID].Chain,
			TokenAddress:    tokenMap[reward.TokenID].Address,
		}

		if _, ok := combinedTransactions[key]; !ok {
			combinedTransactions[key] = &CombinedTransactionValue{Amount: 0, PayRewardIDs: nil}
		}

		// The transaction value is the amount of token.
		combinedTransactions[key].Amount += reward.Amount

		// We also need to know this transaction is for which pay rewards to
		// update status of them.
		combinedTransactions[key].PayRewardIDs = append(combinedTransactions[key].PayRewardIDs, reward.ID)
	}

	return combinedTransactions
}

func (m *BlockchainManager) dispatchCombinedTransactions(
	ctx context.Context, combinedTransactions map[CombinedTransactionKey]*CombinedTransactionValue,
) {
	counter := common.PromCounters[common.BlockchainTransactionFailure]

	// Loop for all combined transaction, we will dispatch and track them.
	for key, value := range combinedTransactions {
		dispatchedTxReq, err := m.getDispatchedTransferTokenTxRequest(
			ctx, key.FromWalletNonce, key.ToAddress, key.Chain, key.TokenAddress, value.Amount)
		if err != nil {
			counter.WithLabelValues("Cannot get dispatched tx request").Inc()
			xcontext.Logger(ctx).Errorf("Cannot get dispatched tx request: %v", err.Error())
			return
		}

		xcontext.Logger(ctx).Infof("Process transaction with hash %s", dispatchedTxReq.Tx.Hash().Hex())

		func() {
			ctx = xcontext.WithDBTransaction(ctx)
			defer xcontext.WithRollbackDBTransaction(ctx)

			// Create blockchain transactions in database to track their status.
			bcTx := &entity.BlockchainTransaction{
				Base:   entity.Base{ID: uuid.NewString()},
				Status: entity.BlockchainTransactionStatusTypeInProgress,
				Chain:  key.Chain,
				TxHash: dispatchedTxReq.Tx.Hash().Hex(),
			}

			if err := m.blockchainRepo.CreateTransaction(ctx, bcTx); err != nil {
				counter.WithLabelValues("Unable to create blockchain transaction to database").Inc()
				xcontext.Logger(ctx).Errorf("Unable to create blockchain transaction to database: %v", err)
				return
			}

			// Update pay rewards status for which this transaction executes.
			txID := sql.NullString{Valid: true, String: bcTx.ID}
			if err := m.payRewardRepo.UpdateTransactionByIDs(ctx, value.PayRewardIDs, txID); err != nil {
				counter.WithLabelValues("Cannot update payreward blockchain transaction").Inc()
				xcontext.Logger(ctx).Errorf("Cannot update payreward blockchain transaction: %v", err)
				return
			}

			// Get the dispatcher and dispatch this transaction.
			dispatcher, ok := m.dispatchers[key.Chain]
			if !ok {
				counter.WithLabelValues("Dispatcher not exists").Inc()
				xcontext.Logger(ctx).Errorf("Dispatcher not exists")
				return
			}

			// Get the watcher and track this transaction status.
			watcher, ok := m.watchers[key.Chain]
			if !ok {
				counter.WithLabelValues("Watcher not exists").Inc()
				xcontext.Logger(ctx).Errorf("Watcher not exists")
				return
			}

			result := dispatcher.Dispatch(ctx, dispatchedTxReq)
			if result.Err != types.ErrNil {
				counter.WithLabelValues("Unable to dispatch").Inc()
				xcontext.Logger(ctx).Errorf("Unable to dispatch: %v", result.Err)
				return
			}

			watcher.TrackTx(ctx, dispatchedTxReq.Tx.Hash().Hex())
			xcontext.WithCommitDBTransaction(ctx)
		}()
	}
}

func (m *BlockchainManager) getDispatchedTransferTokenTxRequest(
	ctx context.Context,
	fromNonce string,
	toAddressHex string,
	chain string,
	tokenAddress string,
	amount float64,
) (*types.DispatchedTxRequest, error) {
	toAddress := ethcommon.HexToAddress(toAddressHex)
	client, ok := m.ethClients[chain]
	if !ok {
		return nil, fmt.Errorf("not support chain %s", chain)
	}

	token, err := m.blockchainRepo.GetToken(ctx, chain, tokenAddress)
	if err != nil {
		return nil, err
	}

	tx, err := client.GetSignedTransferTokenTx(ctx, token, fromNonce, toAddress, amount)
	if err != nil {
		return nil, err
	}

	return &types.DispatchedTxRequest{Chain: chain, Tx: tx}, nil
}
