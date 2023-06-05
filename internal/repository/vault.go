package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm/clause"
)

type VaultRepository interface {
	UpsertVault(ctx context.Context, e *entity.Vault) error
	GetVaultsByChain(ctx context.Context, chain string) ([]entity.Vault, error)
}
type vaultRepository struct {
}

func NewVaultRepository() *vaultRepository {
	return &vaultRepository{}
}

// Vault address
func (r *vaultRepository) UpsertVault(ctx context.Context, e *entity.Vault) error {
	if err := xcontext.DB(ctx).Model(e).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "chain"},
			{Name: "type"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"address": e.Address,
		}),
	}).Create(e).Error; err != nil {
		return err
	}
	return nil
}
func (r *vaultRepository) GetVaultsByChain(ctx context.Context, chain string) ([]entity.Vault, error) {
	result := []entity.Vault{}
	err := xcontext.DB(ctx).
		Find(&result, "chain = ?", chain).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
