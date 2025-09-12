package internal

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type walletRepo struct {
	tx *gorm.DB
}

func NewWalletRepo(tx *gorm.DB) *walletRepo {
	return &walletRepo{tx: tx}
}

// GetOrCreate fetches wallet by userID, creates if not exists.
func (r *walletRepo) GetOrCreate(ctx context.Context, userID uuid.UUID) (*Wallet, error) {
	var wallet Wallet

	err := r.tx.WithContext(ctx).First(&wallet, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		wallet = Wallet{
			UserID:           userID,
			AvailableBalance: 0,
			BlockedBalance:   0,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := r.tx.WithContext(ctx).Create(&wallet).Error; err != nil {
			return nil, err
		}
		return &wallet, nil
	}
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

// GetOrCreateForUpdate fetches wallet with row-level locking.
func (r *walletRepo) GetOrCreateForUpdate(ctx context.Context, userID uuid.UUID) (*Wallet, error) {
	var wallet Wallet

	err := r.tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&wallet, "user_id = ?", userID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		wallet = Wallet{
			UserID:           userID,
			AvailableBalance: 0,
			BlockedBalance:   0,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := r.tx.WithContext(ctx).Create(&wallet).Error; err != nil {
			return nil, err
		}
		return &wallet, nil
	}

	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

// Update updates balances and updated_at
func (r *walletRepo) Update(ctx context.Context, wallet *Wallet) error {
	wallet.UpdatedAt = time.Now()
	return r.tx.WithContext(ctx).
		Model(&Wallet{}).
		Where("user_id = ?", wallet.UserID).
		Updates(map[string]any{
			"available_balance": wallet.AvailableBalance,
			"blocked_balance":   wallet.BlockedBalance,
			"updated_at":        wallet.UpdatedAt,
		}).Error
}
