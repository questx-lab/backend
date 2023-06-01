package domain

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/blockchain/eth"
	interfaze "github.com/questx-lab/backend/pkg/blockchain/interface"
	"github.com/questx-lab/backend/pkg/blockchain/types"

	"github.com/questx-lab/backend/pkg/pubsub"
)

type DispatcherDomain interface {
	Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time)
}

type dispatcherDomain struct {
	cfg             config.ChainConfig
	dispatcher      interfaze.Dispatcher
	watcher         interfaze.Watcher
	transactionRepo repository.TransactionRepository
	client          eth.EthClient
}

func NewDispatcherDomain(
	cfg config.ChainConfig,
	dispatcher interfaze.Dispatcher,
	watcher interfaze.Watcher,
	ethClient eth.EthClient,
	transactionRepo repository.TransactionRepository) *dispatcherDomain {
	return &dispatcherDomain{
		dispatcher:      dispatcher,
		watcher:         watcher,
		transactionRepo: transactionRepo,
		client:          ethClient,
		cfg:             cfg,
	}
}

func (d *dispatcherDomain) getDispatchedTxRequest(tx *model.Transaction) *types.DispatchedTxRequest {

	return &types.DispatchedTxRequest{
		Chain: d.cfg.Chain,
	}
}

func (d *dispatcherDomain) Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time) {
	var tx model.Transaction
	if err := tx.Unmarshal(pack.Msg); err != nil {
		log.Printf("Unable to unmarshal transaction: %v\n", err.Error())
		return
	}
	dispatchedTxReq := d.getDispatchedTxRequest(&tx)
	result := d.dispatcher.Dispatch(dispatchedTxReq)
	if result.Err != types.ErrNil {
		log.Println("Unable to dispatch")
		return
	}
	e := &entity.Transaction{
		Base: entity.Base{
			ID: uuid.NewString(),
		},
		UserID:  tx.User.ID,
		Note:    tx.Note,
		Status:  entity.TransactionPending,
		Address: tx.Address,
		Token:   tx.Token,
		Amount:  tx.Amount,
		TxHash:  dispatchedTxReq.TxHash,
	}
	if tx.ClaimedQuestID != "" {
		if err := e.ClaimedQuestID.Scan(tx.ClaimedQuestID); err != nil {
			log.Printf("Unable to scan claim_quest_id: %v\n", err.Error())
			return
		}
	}

	if err := d.transactionRepo.Create(ctx, e); err != nil {
		log.Printf("Unable to create transaction: %v\n", err.Error())
		return
	}

	d.watcher.TrackTx(dispatchedTxReq.TxHash)
}
