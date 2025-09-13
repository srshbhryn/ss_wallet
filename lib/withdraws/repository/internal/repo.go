package internal

import (
	"context"
	"wallet/lib/withdraws/enums"

	"gorm.io/gorm"
)

type withdrawalRepo struct {
	tx *gorm.DB
}

func NewWithdrawalRepo(tx *gorm.DB) *withdrawalRepo {
	return &withdrawalRepo{tx: tx}
}

func (r *withdrawalRepo) Create(ctx context.Context, w *Withdrawal) error {
	return r.tx.WithContext(ctx).Create(w).Error
}

func (r *withdrawalRepo) Update(ctx context.Context, w *Withdrawal) error {
	return r.tx.WithContext(ctx).Save(w).Error
}

// GetUnFinishedWithdraws returns withdrawals with status "new" or "sent"
func (r *withdrawalRepo) GetUnFinishedWithdraws(ctx context.Context, idPrefix string, bankType enums.BankType) ([]Withdrawal, error) {
	var withdraws []Withdrawal

	query := r.tx.WithContext(ctx).Model(&Withdrawal{}).
		Where("status IN ?", []enums.PayoutStatus{enums.NEW, enums.SENT}).
		Where("bank = ?", bankType)

	if idPrefix != "" {
		query = query.Where("CAST(id AS TEXT) LIKE ?", idPrefix+"%")
	}

	err := query.Find(&withdraws).Error
	return withdraws, err
}

func (r *withdrawalRepo) GetDBTransaction() *gorm.DB {
	return r.tx
}

func (r *withdrawalRepo) Commit() error {
	return r.tx.Commit().Error
}

func (r *withdrawalRepo) RollBack() error {
	return r.tx.Rollback().Error
}
