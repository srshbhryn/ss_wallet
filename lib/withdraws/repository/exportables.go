package repository

import (
	"context"
	"wallet/lib/withdraws/enums"
	"wallet/lib/withdraws/repository/internal"

	"gorm.io/gorm"
)

type Withdrawal = internal.Withdrawal

type Repo interface {
	Create(context.Context, *Withdrawal) error
	Update(context.Context, *Withdrawal) error
	GetUnFinishedWithdraws(ctx context.Context, IDPrefix string, bankType enums.BankType) ([]Withdrawal, error)

	GetDBTransaction() *gorm.DB
	Commit() error
	RollBack() error
}

type RepoFactory interface {
	New(tx *gorm.DB) Repo
}

func NewFactory(db *gorm.DB) RepoFactory {
	return &repoFactory{
		db: db,
	}
}

type repoFactory struct {
	db *gorm.DB
}

func (rf *repoFactory) New(tx *gorm.DB) Repo {
	if tx == nil {
		tx = rf.db.Begin()
	}
	return internal.NewWithdrawalRepo(tx)
}
