package domain

import (
	"context"
	"time"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type PayRewardDomain interface {
	GetMyPayRewards(context.Context, *model.GetMyPayRewardRequest) (*model.GetMyPayRewardResponse, error)
}

type payRewardDomain struct {
	payRewardRepo repository.PayRewardRepository
}

func NewPayRewardDomain(payRewardRepo repository.PayRewardRepository) *payRewardDomain {
	return &payRewardDomain{
		payRewardRepo: payRewardRepo,
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

	return &model.GetMyPayRewardResponse{PayRewards: clientTxs}, nil
}
