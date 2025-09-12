package core

import (
	"wallet/lib/core/internal"

	"gorm.io/gorm"
)

type repo struct {
	tx              *gorm.DB
	walletRepo      WalletRepo
	transactionRepo TransactionRepo
}

func (r *repo) Wallet() WalletRepo {
	return r.walletRepo
}
func (r *repo) Transaction() TransactionRepo {
	return r.transactionRepo
}
func (r *repo) GetDBTransaction() *gorm.DB {
	return r.tx
}
func (r *repo) Commit() error {
	return r.tx.Commit().Error
}
func (r *repo) RollBack() error {
	return r.tx.Rollback().Error
}

type repoFactory struct {
	db *gorm.DB
}

func (rf *repoFactory) New(tx *gorm.DB) Repo {
	if tx == nil {
		tx = rf.db.Begin()
	}
	return &repo{
		tx:              tx,
		walletRepo:      internal.NewWalletRepo(tx),
		transactionRepo: internal.NewTransactionRepo(tx),
	}
}
