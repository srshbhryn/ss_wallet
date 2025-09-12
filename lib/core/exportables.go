package core

import (
	"context"
	"wallet/lib/core/internal"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wallet = internal.Wallet
type Transaction = internal.Transaction

type Repo interface {
	Wallet() WalletRepo
	Transaction() TransactionRepo

	GetDBTransaction() *gorm.DB
	Commit() error
	RollBack() error
}

type RepoFactory interface {
	New(tx *gorm.DB) Repo
}

type WalletRepo interface {
	GetOrCreate(ctx context.Context, userID uuid.UUID) (*Wallet, error)
	GetOrCreateForUpdate(ctx context.Context, userID uuid.UUID) (*Wallet, error)
	Update(ctx context.Context, wallet *Wallet) error
}

type TransactionRepo interface {
	Get(ctx context.Context, userID uuid.UUID, pageNumber int, pageSize int) (transactions []Transaction, hasMore bool, err error)
	Create(ctx context.Context, trx *Transaction) error
}

func NewFactory(db *gorm.DB) RepoFactory {
	return &repoFactory{
		db: db,
	}
}
