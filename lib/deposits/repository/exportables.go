package repository

import (
	"wallet/lib/deposits/repository/internal"

	"gorm.io/gorm"
)

type Deposit = internal.Deposit

type Repo interface {
	Create(*Deposit) error
	Update(*Deposit) error

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
	return internal.NewDepositRepo(tx)
}
