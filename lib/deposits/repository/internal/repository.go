package internal

import (
	"context"

	"gorm.io/gorm"
)

type depositRepo struct {
	tx *gorm.DB
}

func NewDepositRepo(tx *gorm.DB) *depositRepo {
	return &depositRepo{tx: tx}
}

func (r *depositRepo) Create(ctx context.Context, dep *Deposit) error {
	return r.tx.WithContext(ctx).Create(dep).Error
}

func (r *depositRepo) Update(ctx context.Context, dep *Deposit) error {
	return r.tx.WithContext(ctx).
		Model(&Deposit{}).
		Where("id = ?", dep.ID).
		Updates(dep).Error
}

func (r *depositRepo) GetDBTransaction() *gorm.DB {
	return r.tx
}

func (r *depositRepo) Commit() error {
	return r.tx.Commit().Error
}

func (r *depositRepo) RollBack() error {
	return r.tx.Rollback().Error
}
