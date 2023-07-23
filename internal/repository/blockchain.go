package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm/clause"
)

type BlockChainRepository interface {
	// Blockchain
	Upsert(context.Context, *entity.Blockchain) error
	Check(ctx context.Context, chain string) error
	Get(ctx context.Context, chain string) (*entity.Blockchain, error)
	GetAll(ctx context.Context) ([]entity.Blockchain, error)
	GetAllNames(ctx context.Context) ([]string, error)

	// Blockchain Connection
	CreateBlockchainConnection(context.Context, *entity.BlockchainConnection) error
	GetBlockchainConnectionsByChain(ctx context.Context, chain string) ([]entity.BlockchainConnection, error)
	DeleteBlockchainConnection(ctx context.Context, chain, url string) error

	// Token
	CreateToken(context.Context, *entity.BlockchainToken) error
	GetToken(ctx context.Context, chain, address string) (*entity.BlockchainToken, error)
	GetTokenByID(ctx context.Context, id string) (*entity.BlockchainToken, error)
	GetTokenByIDs(ctx context.Context, ids []string) ([]entity.BlockchainToken, error)

	// Transaction
	CreateTransaction(ctx context.Context, e *entity.BlockchainTransaction) error
	UpdateStatusByTxHash(ctx context.Context, txHash, chain string, newStatus entity.BlockchainTransactionStatusType) error
	GetTransactionByID(ctx context.Context, id string) (*entity.BlockchainTransaction, error)
	GetTransactionByTxHash(ctx context.Context, txHash, chain string) (*entity.BlockchainTransaction, error)
}

type blockChainRepository struct{}

func NewBlockChainRepository() *blockChainRepository {
	return &blockChainRepository{}
}

func (r *blockChainRepository) Upsert(ctx context.Context, chain *entity.Blockchain) error {
	return xcontext.DB(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "name"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"id":                     chain.ID,
				"use_eip1559":            chain.UseEip1559,
				"block_time":             chain.BlockTime,
				"adjust_time":            chain.AdjustTime,
				"threshold_update_block": chain.ThresholdUpdateBlock,
			}),
		}).Create(chain).Error
}

func (r *blockChainRepository) Get(ctx context.Context, chain string) (*entity.Blockchain, error) {
	var result entity.Blockchain
	err := xcontext.DB(ctx).Take(&result, "name=?", chain).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *blockChainRepository) Check(ctx context.Context, chain string) error {
	return xcontext.DB(ctx).Select("name").Take(&entity.Blockchain{}, "name=?", chain).Error
}

func (r *blockChainRepository) GetAll(ctx context.Context) ([]entity.Blockchain, error) {
	var result []entity.Blockchain
	err := xcontext.DB(ctx).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *blockChainRepository) GetAllNames(ctx context.Context) ([]string, error) {
	var result []string
	err := xcontext.DB(ctx).Select("name").Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *blockChainRepository) CreateBlockchainConnection(
	ctx context.Context, conn *entity.BlockchainConnection,
) error {
	if err := xcontext.DB(ctx).Create(conn).Error; err != nil {
		return err
	}

	return nil
}

func (r *blockChainRepository) GetBlockchainConnectionsByChain(
	ctx context.Context, chain string,
) ([]entity.BlockchainConnection, error) {
	var result []entity.BlockchainConnection
	err := xcontext.DB(ctx).Find(&result, "chain=?", chain).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *blockChainRepository) DeleteBlockchainConnection(
	ctx context.Context, chain, url string,
) error {
	err := xcontext.DB(ctx).
		Delete(&entity.BlockchainConnection{}, "chain=? AND url=?", chain, url).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *blockChainRepository) CreateTransaction(ctx context.Context, e *entity.BlockchainTransaction) error {
	if err := xcontext.DB(ctx).Create(e).Error; err != nil {
		return err
	}
	return nil
}

func (r *blockChainRepository) UpdateStatusByTxHash(
	ctx context.Context, txHash, chain string, newStatus entity.BlockchainTransactionStatusType,
) error {
	return xcontext.DB(ctx).Model(&entity.BlockchainTransaction{}).
		Where("tx_hash = ? AND chain = ?", txHash, chain).
		Update("status", newStatus).Error
}

func (r *blockChainRepository) GetTransactionByTxHash(ctx context.Context, txHash, chain string) (*entity.BlockchainTransaction, error) {
	var result entity.BlockchainTransaction
	if err := xcontext.DB(ctx).Take(&result, "tx_hash = ? AND chain = ?", txHash, chain).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *blockChainRepository) GetTransactionByID(ctx context.Context, id string) (*entity.BlockchainTransaction, error) {
	var result entity.BlockchainTransaction
	if err := xcontext.DB(ctx).Take(&result, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *blockChainRepository) CreateToken(ctx context.Context, token *entity.BlockchainToken) error {
	return xcontext.DB(ctx).Create(token).Error
}

func (r *blockChainRepository) GetToken(ctx context.Context, chain, address string) (*entity.BlockchainToken, error) {
	var result entity.BlockchainToken
	if err := xcontext.DB(ctx).Take(&result, "chain = ? AND address = ?", chain, address).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *blockChainRepository) GetTokenByID(ctx context.Context, id string) (*entity.BlockchainToken, error) {
	var result entity.BlockchainToken
	if err := xcontext.DB(ctx).Take(&result, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *blockChainRepository) GetTokenByIDs(ctx context.Context, ids []string) ([]entity.BlockchainToken, error) {
	var result []entity.BlockchainToken
	if err := xcontext.DB(ctx).Find(&result, "id IN (?)", ids).Error; err != nil {
		return nil, err
	}

	return result, nil
}
