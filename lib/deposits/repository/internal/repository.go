package internal

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type depositRepo struct {
	tx *gorm.DB
}

func NewDepositRepo(tx *gorm.DB) *depositRepo {
	return &depositRepo{tx: tx}
}

// Create inserts a new deposit and fills in ID automatically.
func (r *depositRepo) Create(ctx context.Context, dep *Deposit) error {
	if dep.ID == uuid.Nil {
		dep.ID = uuid.New()
	}
	return r.tx.WithContext(ctx).Create(dep).Error
}

// Update updates an existing deposit by ID.
func (r *depositRepo) Update(ctx context.Context, dep *Deposit) error {
	return r.tx.WithContext(ctx).
		Model(&Deposit{}).
		Where("id = ?", dep.ID).
		Updates(dep).Error
}

// GetDBTransaction returns the underlying gorm.DB (transaction).
func (r *depositRepo) GetDBTransaction() *gorm.DB {
	return r.tx
}

// Commit commits the transaction.
func (r *depositRepo) Commit() error {
	return r.tx.Commit().Error
}

// RollBack rolls back the transaction.
func (r *depositRepo) RollBack() error {
	return r.tx.Rollback().Error
}

// GetApplicableDeposits fetches deposits that:
// - ID starts with IDPrefix
// - ApplyAt is before now
// - ApplyTransactionID = 0
func (r *depositRepo) GetApplicableDeposits(ctx context.Context, IDPrefix string) ([]Deposit, error) {
	var deposits []Deposit

	if err := r.tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id::text LIKE ?", IDPrefix+"%").
		Where("apply_at < NOW()").
		Where("apply_transaction_id = 0").
		Find(&deposits).Error; err != nil {
		return nil, err
	}

	return deposits, nil
}
