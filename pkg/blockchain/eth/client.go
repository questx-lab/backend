package eth

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/numberutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/crypto/sha3"
	"golang.org/x/net/html"

	"github.com/ethereum/go-ethereum/common"
)

var (
	RpcTimeOut = time.Second * 5
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
	GetSignedTransaction(ctx context.Context, privateKey *ecdsa.PrivateKey, from common.Address, to common.Address, amount *big.Int, gasPrice *big.Int) (*ethtypes.Transaction, error)
}

// Default implementation of ETH client. Since eth RPC often unstable, this client maintains a list
// of different RPC to connect to and uses the ones that is stable to dispatch a transaction.
type defaultEthClient struct {
	chain           string
	useExternalRpcs bool

	clients     []*ethclient.Client
	healthies   []bool
	initialRpcs []string
	rpcs        []string

	lock *sync.RWMutex
}

func NewEthClients(cfg config.ChainConfig, useExternalRpcs bool) EthClient {
	c := &defaultEthClient{
		chain:           cfg.Chain,
		useExternalRpcs: useExternalRpcs,
		initialRpcs:     cfg.Rpcs,
		lock:            &sync.RWMutex{},
	}

	return c
}

func (c *defaultEthClient) Start(ctx context.Context) {
	go c.loopCheck(ctx)
}

// loopCheck
func (c *defaultEthClient) loopCheck(ctx context.Context) {
	for {
		// Sleep a random time between 5 & 10 minutes
		mins := rand.Intn(5) + 5
		sleepTime := time.Second * time.Duration(60*mins)
		time.Sleep(sleepTime)

		c.updateRpcs(ctx)
	}
}

func (c *defaultEthClient) updateRpcs(ctx context.Context) {
	c.lock.RLock()
	rpcs := c.initialRpcs
	c.lock.RUnlock()

	if c.useExternalRpcs {
		// Get external rpcs.
		externals, err := c.GetExtraRpcs(ctx)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Failed to get external rpc info, err = ", err)
		} else {
			rpcs = append(rpcs, externals...)
		}
	}

	c.lock.RLock()
	oldClients := c.clients
	c.lock.RUnlock()

	rpcs, clients, healthies := c.getRpcsHealthiness(ctx, rpcs)

	// Close all the old clients
	c.lock.Lock()
	for _, client := range oldClients {
		client.Close()
	}

	c.rpcs, c.clients, c.healthies = rpcs, clients, healthies
	c.lock.Unlock()
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
			ctx, cancel := context.WithTimeout(context.Background(), RpcTimeOut)
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
	chainId := GetChainIntFromId(ctx, c.chain)
	url := fmt.Sprintf("https://chainlist.org/chain/%d", chainId)
	xcontext.Logger(ctx).Infof("Getting extra rpcs status from remote link %s for chain %s", url, c.chain)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Failed to get chain list data, status code = %d", res.StatusCode)
	}

	bz, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	ret := c.processData(string(bz))

	return ret, nil
}

func (c *defaultEthClient) shuffle() ([]*ethclient.Client, []bool, []string) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	n := len(c.clients)

	clients := make([]*ethclient.Client, n)
	healthy := make([]bool, n)
	rpcs := make([]string, n)

	copy(clients, c.clients)
	copy(healthy, c.healthies)
	copy(rpcs, c.rpcs)

	for i := 0; i < 20; i++ {
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
	c.lock.RLock()
	if c.clients == nil {
		c.lock.RUnlock()
		c.updateRpcs(ctx)
	} else {
		c.lock.RUnlock()
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
		return nil, fmt.Errorf("No healthy RPC for chain %s", c.chain)
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

	return num.(uint64), err
}

func (c *defaultEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*ethtypes.Block, error) {
	block, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		return client.BlockByNumber(ctx, number)
	})

	return block.(*ethtypes.Block), err
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

	return gas.(*big.Int), err
}

func (c *defaultEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	nonce, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		return client.PendingNonceAt(ctx, account)
	})

	return nonce.(uint64), err
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

	return balance.(*big.Int), err
}

func (c *defaultEthClient) GetSignedTransaction(
	ctx context.Context,
	privateKey *ecdsa.PrivateKey,
	from common.Address,
	to common.Address,
	amount *big.Int,
	gasPrice *big.Int,
) (*ethtypes.Transaction, error) {
	signedTx, err := c.execute(ctx, func(client *ethclient.Client, rpc string) (any, error) {
		nonce, err := client.PendingNonceAt(ctx, from)
		if err != nil {
			return nil, err
		}
		data := c.GetTransferData(ctx, to, amount)
		gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
			To:   &to,
			Data: data,
		})
		if err != nil {
			return nil, err
		}
		tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
		chainID := GetChainIntFromId(ctx, c.chain)
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
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

func (c *defaultEthClient) GetTransferData(ctx context.Context, to common.Address, amount *big.Int) []byte {
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]

	paddedAddress := common.LeftPadBytes(to.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	return data
}
