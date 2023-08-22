package blockchain

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/domain/blockchain/eth"
	"github.com/questx-lab/backend/internal/domain/blockchain/types"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/ethutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
)

type ERC20TransactionKey struct {
	FromWalletNonce string
	ToAddress       string
	Chain           string
	TokenAddress    string
}

type ERC20TransactionValue struct {
	Amount       float64
	PayRewardIDs []string
}

type ERC1155TransactionKey struct {
	FromWalletNonce string
	Chain           string
}

type ERC1155TransactionValue struct {
	ToAddress   string
	TokenID     int64
	Amount      int
	PayRewardID string
}

type BlockchainManager struct {
	rootCtx        context.Context
	payRewardRepo  repository.PayRewardRepository
	blockchainRepo repository.BlockChainRepository
	communityRepo  repository.CommunityRepository
	nftRepo        repository.NftRepository
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
	nftRepo repository.NftRepository,
	redisClient xredis.Client,
) *BlockchainManager {
	return &BlockchainManager{
		rootCtx:        ctx,
		blockchainRepo: blockchainRepo,
		payRewardRepo:  payRewardRepo,
		communityRepo:  communityRepo,
		nftRepo:        nftRepo,
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

func (m *BlockchainManager) MintNFT(
	_ context.Context, communityID, chain string, nftID int64, amount int, ipfs string,
) error {
	ctx := m.rootCtx

	client, ok := m.ethClients[chain]
	if !ok {
		return fmt.Errorf("not support chain %s", chain)
	}

	community, err := m.communityRepo.GetByID(ctx, communityID)
	if err != nil {
		return err
	}

	communityAddress, err := ethutil.GeneratePublicKey(
		[]byte(xcontext.Configs(ctx).Blockchain.SecretKey), []byte(community.WalletNonce))
	if err != nil {
		return err
	}

	tx, err := client.GetSignedMintNftTx(ctx, communityAddress, nftID, amount, ipfs)
	if err != nil {
		return err
	}

	// Get the dispatcher and dispatch this transaction.
	dispatcher, ok := m.dispatchers[chain]
	if !ok {
		return fmt.Errorf("dispatcher %s not exists", chain)
	}

	// Get the watcher and track this transaction status.
	watcher, ok := m.watchers[chain]
	if !ok {
		return fmt.Errorf("watcher %s not exists", chain)
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
		return err
	}

	nftMintHistory := &entity.NonFungibleTokenMintHistory{
		NonFungibleTokenID: nftID,
		TransactionID:      bcTx.ID,
		Amount:             amount,
	}

	if err := m.nftRepo.CreateHistory(ctx, nftMintHistory); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to create nft mint history: %v", err)
		return errorx.Unknown
	}

	result := dispatcher.Dispatch(ctx, &types.DispatchedTxRequest{Chain: chain, Tx: tx})
	if result.Err != types.ErrNil {
		return fmt.Errorf("unable to dispatch: %v", result.Err)
	}

	watcher.TrackMintTx(ctx, tx.Hash().Hex(), nftID, amount)
	xcontext.WithCommitDBTransaction(ctx)

	return nil
}

func (m *BlockchainManager) DeployNFT(_ context.Context, chain string) (string, error) {
	client, ok := m.ethClients[chain]
	if !ok {
		return "", fmt.Errorf("not support chain %s", chain)
	}

	return client.DeployXquestNFT(m.rootCtx)
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
	watcher := eth.NewEthWatcher(ctx, blockchain, m.blockchainRepo, m.nftRepo, client, m.redisClient)

	m.ethClients[blockchain.Name] = client
	m.dispatchers[blockchain.Name] = dispatcher
	m.watchers[blockchain.Name] = watcher

	go client.Start(ctx)
	go watcher.Start(ctx)
}

func (m *BlockchainManager) handlePendingPayRewards(ctx context.Context) {
	erc20Transactions, erc1155Transactions := m.combineTransactions(ctx)
	if erc20Transactions == nil {
		return
	}

	m.dispatchERC20Transactions(ctx, erc20Transactions)
	m.dispatchERC1155Transactions(ctx, erc1155Transactions)
}

func (m *BlockchainManager) combineTransactions(
	ctx context.Context,
) (map[ERC20TransactionKey]*ERC20TransactionValue, map[ERC1155TransactionKey][]ERC1155TransactionValue) {
	allPendingPayRewards, err := m.payRewardRepo.GetAllPending(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all pending pay rewards: %v", err)
		return nil, nil
	}

	// Combine all pay rewards in the same chain, token, sender, recipient for
	// one transaction.
	walletNonces := map[string]string{}
	tokenMap := map[string]*entity.BlockchainToken{}
	nftMap := map[int64]*entity.NonFungibleToken{}
	erc20Transactions := map[ERC20TransactionKey]*ERC20TransactionValue{}
	erc1155Transactions := map[ERC1155TransactionKey][]ERC1155TransactionValue{}
	for _, reward := range allPendingPayRewards {
		// Our platform doesn't need a wallet nonce to generate private/public key.
		var walletNonce string

		if reward.TokenID.Valid == reward.NonFungibleTokenID.Valid {
			xcontext.Logger(ctx).Errorf(
				"An invalid pay reward record, containing both token and nft: %s", reward.ID)
			continue
		}

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

		if reward.TokenID.Valid {
			if _, ok := tokenMap[reward.TokenID.String]; !ok {
				token, err := m.blockchainRepo.GetTokenByID(ctx, reward.TokenID.String)
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot get token %s of reward %s: %v",
						reward.TokenID.String, reward.ID, err)
					continue
				}

				tokenMap[token.ID] = token
			}

			// The transaction key is combination of from, to, chain, token.
			key := ERC20TransactionKey{
				FromWalletNonce: walletNonce,
				ToAddress:       reward.ToAddress,
				Chain:           tokenMap[reward.TokenID.String].Chain,
				TokenAddress:    tokenMap[reward.TokenID.String].Address,
			}

			if _, ok := erc20Transactions[key]; !ok {
				erc20Transactions[key] = &ERC20TransactionValue{Amount: 0, PayRewardIDs: nil}
			}

			// The transaction value is the amount of token.
			erc20Transactions[key].Amount += reward.Amount

			// We also need to know this transaction is for which pay rewards to
			// update status of them.
			erc20Transactions[key].PayRewardIDs = append(erc20Transactions[key].PayRewardIDs, reward.ID)
		} else {
			if _, ok := nftMap[reward.NonFungibleTokenID.Int64]; !ok {
				nft, err := m.nftRepo.GetByID(ctx, reward.NonFungibleTokenID.Int64)
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot get nft %s of reward %s: %v",
						reward.NonFungibleTokenID.Int64, reward.ID, err)
					continue
				}

				nftMap[nft.ID] = nft
			}

			// The transaction key is combination of from address and chain.
			key := ERC1155TransactionKey{
				FromWalletNonce: walletNonce,
				Chain:           nftMap[reward.NonFungibleTokenID.Int64].Chain,
			}

			// We also need to know this transaction is for which pay rewards to
			// update status of them.
			erc1155Transactions[key] = append(erc1155Transactions[key], ERC1155TransactionValue{
				ToAddress:   reward.ToAddress,
				TokenID:     reward.NonFungibleTokenID.Int64,
				Amount:      int(reward.Amount),
				PayRewardID: reward.ID,
			})
		}
	}

	return erc20Transactions, erc1155Transactions
}

func (m *BlockchainManager) dispatchERC20Transactions(
	ctx context.Context, erc20Transactions map[ERC20TransactionKey]*ERC20TransactionValue,
) {
	// Loop for all combined transaction, we will dispatch and track them.
	for key, value := range erc20Transactions {
		dispatchedTxReq, err := m.getDispatchedTransferTokenTxRequest(
			ctx, key.FromWalletNonce, key.ToAddress, key.Chain, key.TokenAddress, value.Amount)
		if err != nil {
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
				xcontext.Logger(ctx).Errorf("Unable to create blockchain transaction to database: %v", err)
				return
			}

			// Update pay rewards status for which this transaction executes.
			txID := sql.NullString{Valid: true, String: bcTx.ID}
			if err := m.payRewardRepo.UpdateTransactionByIDs(ctx, value.PayRewardIDs, txID); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot update payreward blockchain transaction: %v", err)
				return
			}

			// Get the dispatcher and dispatch this transaction.
			dispatcher, ok := m.dispatchers[key.Chain]
			if !ok {
				xcontext.Logger(ctx).Errorf("Dispatcher not exists")
				return
			}

			// Get the watcher and track this transaction status.
			watcher, ok := m.watchers[key.Chain]
			if !ok {
				xcontext.Logger(ctx).Errorf("Watcher not exists")
				return
			}

			result := dispatcher.Dispatch(ctx, dispatchedTxReq)
			if result.Err != types.ErrNil {
				xcontext.Logger(ctx).Errorf("Unable to dispatch: %v", result.Err)
				return
			}

			watcher.TrackTx(ctx, dispatchedTxReq.Tx.Hash().Hex())
			xcontext.WithCommitDBTransaction(ctx)
		}()
	}
}

func (m *BlockchainManager) dispatchERC1155Transactions(
	ctx context.Context, erc1155Transactions map[ERC1155TransactionKey][]ERC1155TransactionValue,
) {
	// Loop for all combined transaction, we will dispatch and track them.
	for key, value := range erc1155Transactions {
		dispatchedTxReq, err := m.getDispatchedERC1155TxRequest(
			ctx, key.FromWalletNonce, key.Chain, value)
		if err != nil {
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
				xcontext.Logger(ctx).Errorf("Unable to create blockchain transaction to database: %v", err)
				return
			}

			// Update pay rewards status for which this transaction executes.
			payRewardIDs := []string{}
			for _, v := range value {
				payRewardIDs = append(payRewardIDs, v.PayRewardID)
			}

			txID := sql.NullString{Valid: true, String: bcTx.ID}
			if err := m.payRewardRepo.UpdateTransactionByIDs(ctx, payRewardIDs, txID); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot update payreward blockchain transaction: %v", err)
				return
			}

			// Get the dispatcher and dispatch this transaction.
			dispatcher, ok := m.dispatchers[key.Chain]
			if !ok {
				xcontext.Logger(ctx).Errorf("Dispatcher not exists")
				return
			}

			// Get the watcher and track this transaction status.
			watcher, ok := m.watchers[key.Chain]
			if !ok {
				xcontext.Logger(ctx).Errorf("Watcher not exists")
				return
			}

			result := dispatcher.Dispatch(ctx, dispatchedTxReq)
			if result.Err != types.ErrNil {
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

func (m *BlockchainManager) getDispatchedERC1155TxRequest(
	ctx context.Context,
	fromNonce string,
	chain string,
	values []ERC1155TransactionValue,
) (*types.DispatchedTxRequest, error) {
	client, ok := m.ethClients[chain]
	if !ok {
		return nil, fmt.Errorf("not support chain %s", chain)
	}

	recipients := []ethcommon.Address{}
	nftIDs := []int64{}
	amounts := []int{}
	for _, v := range values {
		recipients = append(recipients, ethcommon.HexToAddress(v.ToAddress))
		nftIDs = append(nftIDs, v.TokenID)
		amounts = append(amounts, v.Amount)
	}

	tx, err := client.GetSignedTransferNFTsTx(ctx, fromNonce, recipients, nftIDs, amounts)
	if err != nil {
		return nil, err
	}

	return &types.DispatchedTxRequest{Chain: chain, Tx: tx}, nil
}
