package internal

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type transactionRepo struct {
	tx *gorm.DB
}

func NewTransactionRepo(tx *gorm.DB) *transactionRepo {
	return &transactionRepo{tx: tx}
}

// Get returns transactions for a given userID with pagination.
// It also returns whether there are more records beyond this page.
func (r *transactionRepo) Get(ctx context.Context, userID uuid.UUID, pageNumber int, pageSize int) ([]Transaction, bool, error) {
	var (
		transactions []Transaction
		count        int64
	)

	// first resolve wallet ID for the given user
	var wallet Wallet
	if err := r.tx.WithContext(ctx).
		Select("id").
		Where("user_id = ?", userID).
		First(&wallet).Error; err != nil {
		return nil, false, err
	}

	// count total transactions for pagination
	if err := r.tx.WithContext(ctx).
		Model(&Transaction{}).
		Where("wallet_id = ?", wallet.UserID).
		Count(&count).Error; err != nil {
		return nil, false, err
	}

	// fetch transactions with offset/limit
	offset := (pageNumber - 1) * pageSize
	if err := r.tx.WithContext(ctx).
		Where("wallet_id = ?", wallet.UserID).
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&transactions).Error; err != nil {
		return nil, false, err
	}

	hasMore := int64(offset+pageSize) < count
	return transactions, hasMore, nil
}

func (r *transactionRepo) Create(ctx context.Context, trx *Transaction) error {
	return r.tx.WithContext(ctx).Create(trx).Error
}
