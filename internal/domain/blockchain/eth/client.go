package eth

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/questx-lab/backend/contract/erc20"
	"github.com/questx-lab/backend/contract/xquestnft"
	"github.com/questx-lab/backend/internal/domain/blockchain/types"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/ethutil"
	"github.com/questx-lab/backend/pkg/numberutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/net/html"

	"github.com/ethereum/go-ethereum/common"
)

const (
	RpcTimeOut      = time.Second * 5
	MaxShuffleTimes = 20
)

// A wrapper around eth.client so that we can mock in watcher tests.
type EthClient interface {
	Start(ctx context.Context)

	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*ethtypes.Block, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SendTransaction(ctx context.Context, tx *ethtypes.Transaction) error
	BalanceAt(ctx context.Context, from common.Address, block *big.Int) (*big.Int, error)
	GetSignedTransferTokenTx(ctx context.Context, token *entity.BlockchainToken, senderNonce string, recipient common.Address, amount float64) (*ethtypes.Transaction, error)
	GetSignedMintNftTx(ctx context.Context, mintTo common.Address, nftID int64, amount int) (*ethtypes.Transaction, error)
	GetTokenInfo(ctx context.Context, address string) (types.TokenInfo, error)
	ERC20BalanceOf(ctx context.Context, tokenAddress, accountAddress string) (*big.Int, error)
}

// Default implementation of ETH client. Since eth RPC often unstable, this client maintains a list
// of different RPC to connect to and uses the ones that is stable to dispatch a transaction.
type defaultEthClient struct {
	chain           string
	chainID         *big.Int
	useExternalRpcs bool

	clients   []*ethclient.Client
	healthies []bool
	rpcs      []string

	mutex sync.RWMutex

	blockchainRepo repository.BlockChainRepository
}

func NewEthClients(
	blockchain *entity.Blockchain,
	blockchainRepo repository.BlockChainRepository,
) EthClient {
	c := &defaultEthClient{
		chain:           blockchain.Name,
		chainID:         big.NewInt(blockchain.ID),
		useExternalRpcs: blockchain.UseExternalRPC,
		mutex:           sync.RWMutex{},
		blockchainRepo:  blockchainRepo,
	}

	return c
}

func (c *defaultEthClient) Start(ctx context.Context) {
	go c.loopCheck(ctx)
}

// loopCheck
func (c *defaultEthClient) loopCheck(ctx context.Context) {
	for {
		time.Sleep(xcontext.Configs(ctx).Blockchain.RefreshConnectionFrequency)
		c.updateRpcs(ctx)
	}
}

func (c *defaultEthClient) updateRpcs(ctx context.Context) {
	rpcs := []string{}
	connections, err := c.blockchainRepo.GetConnectionsByChain(ctx, c.chain)
	if err != nil || len(connections) == 0 {
		xcontext.Logger(ctx).Errorf("Cannot get any connections of chain %s: %v", c.chain, err)
	} else {
		for _, conn := range connections {
			if conn.Type == entity.BlockchainConnectionRPC {
				rpcs = append(rpcs, "https://"+conn.URL)
			}
		}
	}

	if c.useExternalRpcs {
		// Get external rpcs.
		externals, err := c.GetExtraRpcs(ctx)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Failed to get external rpc info: %v", err)
		} else {
			rpcs = append(rpcs, externals...)
		}
	}

	c.mutex.RLock()
	oldClients := c.clients
	c.mutex.RUnlock()

	rpcs, clients, healthies := c.getRpcsHealthiness(ctx, rpcs)

	// Close all the old clients
	c.mutex.Lock()
	for _, client := range oldClients {
		client.Close()
	}

	c.rpcs, c.clients, c.healthies = rpcs, clients, healthies
	c.mutex.Unlock()
}

func (c *defaultEthClient) getRpcsHealthiness(ctx context.Context, allRpcs []string) ([]string, []*ethclient.Client, []bool) {
	clients := make([]*ethclient.Client, 0)
	rpcs := make([]string, 0)
	healthies := make([]bool, 0)

	type healthyNode struct {
		client *ethclient.Client
		rpc    string
		height int64
	}

	nodes := make([]*healthyNode, 0)
	for _, rpc := range allRpcs {
		client, err := ethclient.Dial(rpc)
		if err == nil {
			var cancel func()
			ctx, cancel = context.WithTimeout(ctx, RpcTimeOut)
			block, err := client.BlockByNumber(ctx, nil)
			cancel()

			if err == nil && block.Number() != nil {
				nodes = append(nodes, &healthyNode{
					client: client,
					rpc:    rpc,
					height: block.Number().Int64(),
				})
			}

			client.Close()
		}
	}

	if len(nodes) == 0 {
		return rpcs, clients, healthies
	}

	// Sorts all nodes by height
	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].height > nodes[j].height
	})

	// Only select some nodes within a certain height from the median
	height := nodes[len(nodes)/2].height
	for _, node := range nodes {
		if numberutil.AbsInt64(node.height-height) < 5 {
			rpcs = append(rpcs, node.rpc)
			clients = append(clients, node.client)
			healthies = append(healthies, true)
		}
	}

	// Log all healthy rpcs
	xcontext.Logger(ctx).Errorf("Healthy rpcs for chain %s: %s", c.chain, rpcs)

	return rpcs, clients, healthies
}

func (c *defaultEthClient) processData(text string) []string {
	tokenizer := html.NewTokenizer(strings.NewReader(text))
	var data string
	for {
		tokenType := tokenizer.Next()
		stop := false
		switch tokenType {
		case html.ErrorToken:
			stop = true

		case html.TextToken:
			text := tokenizer.Token().Data
			var js json.RawMessage
			if json.Unmarshal([]byte(text), &js) == nil {
				data = text
			}
		}

		if stop {
			break
		}
	}

	// Process the data
	type result struct {
		Props struct {
			PageProps struct {
				Chain struct {
					Name string `json:"name"`
					RPC  []struct {
						Url string `json:"url"`
					} `json:"rpc"`
				} `json:"chain"`
			} `json:"pageProps"`
		} `json:"props"`
	}

	r := &result{}
	err := json.Unmarshal([]byte(data), r)
	if err != nil {
		panic(err)
	}

	ret := make([]string, 0)
	for _, rpc := range r.Props.PageProps.Chain.RPC {
		ret = append(ret, rpc.Url)
	}

	return ret
}

func (c *defaultEthClient) GetExtraRpcs(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("https://chainlist.org/chain/%d", c.chainID)
	xcontext.Logger(ctx).Infof("Getting extra rpcs status from remote link %s for chain %s",
		url, c.chain)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get chain list data, status code = %d", res.StatusCode)
	}

	bz, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	ret := c.processData(string(bz))

	return ret, nil
}

func (c *defaultEthClient) shuffle() ([]*ethclient.Client, []bool, []string) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	n := len(c.clients)
	if n == 0 {
		return nil, nil, nil
	}

	clients := make([]*ethclient.Client, n)
	healthy := make([]bool, n)
	rpcs := make([]string, n)

	copy(clients, c.clients)
	copy(healthy, c.healthies)
	copy(rpcs, c.rpcs)

	for i := 0; i < MaxShuffleTimes; i++ {
		x := rand.Intn(n)
		y := rand.Intn(n)

		tmpClient := clients[x]
		clients[x] = clients[y]
		clients[y] = tmpClient

		tmpHealth := healthy[x]
		healthy[x] = healthy[y]
		healthy[y] = tmpHealth

		tmpRpc := rpcs[x]
		rpcs[x] = rpcs[y]
		rpcs[y] = tmpRpc
	}

	return clients, healthy, rpcs
}

func (c *defaultEthClient) getHealthyClient(ctx context.Context) (*ethclient.Client, string) {
	c.mutex.RLock()
	if c.clients == nil {
		c.mutex.RUnlock()
		c.updateRpcs(ctx)
	} else {
		c.mutex.RUnlock()
	}

	// Shuffle rpcs so that we will use different healthy rpc
	clients, healthies, rpcs := c.shuffle()
	for i, healthy := range healthies {
		if healthy {
			return clients[i], rpcs[i]
		}
	}

	return nil, ""
}

func (c *defaultEthClient) execute(ctx context.Context, f func(client *ethclient.Client, rpc string) (any, error)) (any, error) {
	client, rpc := c.getHealthyClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("no healthy RPC for chain %s", c.chain)
	}

	ret, err := f(client, rpc)
	if err == nil {
		return ret, nil
	}

	return ret, err
}

func (c *defaultEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	num, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		return client.BlockNumber(ctx)
	})

	if err != nil {
		return 0, err
	}

	return num.(uint64), nil
}

func (c *defaultEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*ethtypes.Block, error) {
	block, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		return client.BlockByNumber(ctx, number)
	})

	if err != nil {
		return nil, err
	}

	return block.(*ethtypes.Block), nil
}

func (c *defaultEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error) {
	receipt, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		return client.TransactionReceipt(ctx, txHash)
	})

	return receipt.(*ethtypes.Receipt), err
}

func (c *defaultEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	gas, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		return client.SuggestGasPrice(ctx)
	})

	if err != nil {
		return nil, err
	}

	return gas.(*big.Int), nil
}

func (c *defaultEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	nonce, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		return client.PendingNonceAt(ctx, account)
	})

	if err != nil {
		return 0, err
	}

	return nonce.(uint64), nil
}

func (c *defaultEthClient) SendTransaction(ctx context.Context, tx *ethtypes.Transaction) error {
	_, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		err := client.SendTransaction(ctx, tx)
		return 0, err
	})

	return err
}

func (c *defaultEthClient) BalanceAt(ctx context.Context, from common.Address, block *big.Int) (*big.Int, error) {
	balance, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		balance, err := client.BalanceAt(ctx, from, block)
		if err == nil && balance != nil && balance.Cmp(big.NewInt(0)) == 0 {
			xcontext.Logger(ctx).Errorf("Balance is 0 for using URL %s", rpc)
		}

		return balance, err
	})

	if err != nil {
		return nil, err
	}

	return balance.(*big.Int), err
}

func (c *defaultEthClient) GetSignedTransferTokenTx(
	ctx context.Context,
	token *entity.BlockchainToken,
	senderNonce string,
	recipient common.Address,
	amount float64,
) (*ethtypes.Transaction, error) {
	signedTx, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		tokenInstance, err := erc20.NewErc20(common.HexToAddress(token.Address), client)
		if err != nil {
			return nil, err
		}

		secret := xcontext.Configs(ctx).Blockchain.SecretKey
		senderPrivateKey, err := ethutil.GeneratePrivateKey([]byte(secret), []byte(senderNonce))
		if err != nil {
			return nil, err
		}

		signedTx, err := tokenInstance.Transfer(
			c.TransactionOpts(ctx, senderPrivateKey, common.Big0),
			recipient,
			big.NewInt(int64(amount*math.Pow10(token.Decimals))),
		)
		if err != nil {
			return nil, err
		}

		return signedTx, nil
	})
	if err != nil {
		return nil, err
	}

	return signedTx.(*ethtypes.Transaction), nil
}

func (c *defaultEthClient) GetSignedMintNftTx(
	ctx context.Context,
	mintTo common.Address,
	nftID int64,
	amount int,
) (*ethtypes.Transaction, error) {
	signedTx, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		blockchain, err := c.blockchainRepo.Get(ctx, c.chain)
		if err != nil {
			return nil, err
		}

		tokenInstance, err := xquestnft.NewXquestnft(common.HexToAddress(blockchain.XQuestNFTAddress), client)
		if err != nil {
			return nil, err
		}

		secret := xcontext.Configs(ctx).Blockchain.SecretKey
		platformPrivateKey, err := ethutil.GeneratePrivateKey([]byte(secret), []byte{})
		if err != nil {
			return nil, err
		}

		signedTx, err := tokenInstance.Mint(
			c.TransactionOpts(ctx, platformPrivateKey, common.Big0),
			mintTo,
			big.NewInt(nftID),
			big.NewInt(int64(amount)),
			nil,
		)
		if err != nil {
			return nil, err
		}

		return signedTx, nil
	})
	if err != nil {
		return nil, err
	}

	return signedTx.(*ethtypes.Transaction), nil
}

func (c *defaultEthClient) TransactionOpts(
	ctx context.Context, fromPrivateKey *ecdsa.PrivateKey, value *big.Int,
) *bind.TransactOpts {
	return &bind.TransactOpts{
		From: crypto.PubkeyToAddress(fromPrivateKey.PublicKey),
		Signer: func(a common.Address, t *ethtypes.Transaction) (*ethtypes.Transaction, error) {
			signedTx, err := ethtypes.SignTx(t, ethtypes.NewEIP155Signer(c.chainID), fromPrivateKey)
			if err != nil {
				return nil, err
			}
			return signedTx, nil
		},
		Value:  value,
		NoSend: true,
	}
}

func (c *defaultEthClient) GetTokenInfo(ctx context.Context, address string) (types.TokenInfo, error) {
	info, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		tokenInstance, err := erc20.NewErc20(common.HexToAddress(address), client)
		if err != nil {
			return nil, err
		}

		symbol, err := tokenInstance.Symbol(nil)
		if err != nil {
			return nil, err
		}

		decimals, err := tokenInstance.Decimals(nil)
		if err != nil {
			return nil, err
		}

		name, err := tokenInstance.Name(nil)
		if err != nil {
			return nil, err
		}

		return types.TokenInfo{Name: name, Symbol: symbol, Decimals: int(decimals)}, err
	})

	if err != nil {
		return types.TokenInfo{}, err
	}

	return info.(types.TokenInfo), err
}

func (c *defaultEthClient) ERC20BalanceOf(ctx context.Context, tokenAddress, accountAddress string) (*big.Int, error) {
	balance, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		tokenInstance, err := erc20.NewErc20(common.HexToAddress(tokenAddress), client)
		if err != nil {
			return nil, err
		}

		balance, err := tokenInstance.BalanceOf(nil, common.HexToAddress(accountAddress))
		if err != nil {
			return nil, err
		}

		return balance, nil
	})

	if err != nil {
		return nil, err
	}

	return balance.(*big.Int), err
}
