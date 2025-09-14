package internal

import (
	"time"
	"wallet/lib/withdraws/enums"

	"github.com/google/uuid"
)

type Withdrawal struct {
	ID                      uuid.UUID          `gorm:"type:uuid;primaryKey" json:"id"`
	WalletID                uuid.UUID          `gorm:"type:uuid;index" json:"wallet_id"`
	Status                  enums.PayoutStatus `gorm:"type:varchar(32);index" json:"status"`
	Bank                    enums.BankType     `gorm:"type:varchar(32);index" json:"bank"`
	Iban                    string             `gorm:"type:varchar(24);not null;index" json:"iban"`
	BlockTransactionID      uint64             `gorm:"index" json:"block_transaction_id"`
	WithdrawalTransactionID uint64             `gorm:"index" json:"withdrawal_transaction_id"`
	ReverserTransactionID   uint64             `gorm:"index" json:"reverser_transaction_id"`
	Amount                  int64              `gorm:"not null" json:"amount"`
	CreatedAt               time.Time          `json:"created_at"`
	UpdatedAt               time.Time          `json:"updated_at"`
}
