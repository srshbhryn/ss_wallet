package internal

import (
	"time"
	"wallet/lib/withdraws/enums"

	"github.com/google/uuid"
)

type Withdrawal struct {
	ID                      uuid.UUID          `gorm:"primaryKey"`
	WalletID                uuid.UUID          `gorm:"index"`
	Status                  enums.PayoutStatus `gorm:"index"`
	Bank                    enums.BankType     `gorm:"index"`
	BlockTransactionID      uint64             `gorm:"index"`
	WithdrawalTransactionID uint64             `gorm:"index"`
	ReverserTransactionID   uint64             `gorm:"index"`
	Amount                  int64
	CreatedAt               time.Time
	UpdatedAt               time.Time
}
