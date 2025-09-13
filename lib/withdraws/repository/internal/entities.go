package internal

import (
	"time"
	"wallet/lib/withdraws/enums"

	"github.com/google/uuid"
)

type Withdrawal struct {
	ID                      uuid.UUID          `gorm:"type:uuid;primaryKey"`
	WalletID                uuid.UUID          `gorm:"type:uuid;index"`
	Status                  enums.PayoutStatus `gorm:"type:varchar(32);index"`
	Bank                    enums.BankType     `gorm:"type:varchar(32);index"`
	Iban                    string             `gorm:"type:varchar(24);not null;index"`
	BlockTransactionID      uint64             `gorm:"index"`
	WithdrawalTransactionID uint64             `gorm:"index"`
	ReverserTransactionID   uint64             `gorm:"index"`
	Amount                  int64              `gorm:"not null"`
	CreatedAt               time.Time
	UpdatedAt               time.Time
}
