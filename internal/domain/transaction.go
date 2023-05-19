package domain

import (
	"time"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type TransactionDomain interface {
	GetMyTransactions(xcontext.Context, *model.GetMyTransactionRequest) (*model.GetMyTransactionResponse, error)
}

type transactionDomain struct {
	transactionRepo repository.TransactionRepository
}

func NewTransactionDomain(transactionRepo repository.TransactionRepository) *transactionDomain {
	return &transactionDomain{
		transactionRepo: transactionRepo,
	}
}

func (d *transactionDomain) GetMyTransactions(
	ctx xcontext.Context, req *model.GetMyTransactionRequest,
) (*model.GetMyTransactionResponse, error) {
	txs, err := d.transactionRepo.GetByUserID(ctx, xcontext.GetRequestUserID(ctx))
	if err != nil {
		ctx.Logger().Errorf("Cannot get transaction by user id: %v", err)
		return nil, errorx.Unknown
	}

	clientTxs := []model.Transaction{}
	for _, tx := range txs {
		clientTxs = append(clientTxs, model.Transaction{
			ID:             tx.ID,
			CreatedAt:      tx.CreatedAt.Format(time.RFC3339Nano),
			ClaimedQuestID: tx.ClaimedQuestID.String,
			Note:           tx.Note,
			Status:         string(tx.Status),
			Address:        tx.Address,
			Token:          tx.Token,
			Amount:         tx.Amount,
		})
	}

	return &model.GetMyTransactionResponse{Transactions: clientTxs}, nil
}
