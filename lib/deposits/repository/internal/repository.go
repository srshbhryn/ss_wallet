package internal

import (
	"gorm.io/gorm"
)

type depositRepo struct {
	tx *gorm.DB
}

// constructor
func NewDepositRepo(tx *gorm.DB) *depositRepo {
	return &depositRepo{tx: tx}
}

func (r *depositRepo) Create(dep Deposit) error {
	return r.tx.Create(&dep).Error
}

func (r *depositRepo) Update(dep Deposit) error {
	return r.tx.Model(&Deposit{}).
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
