package internal

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
