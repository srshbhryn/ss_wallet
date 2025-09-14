package internal

import (
	"time"

	"github.com/google/uuid"
)

type Deposit struct {
	ID                 uuid.UUID `gorm:"primaryKey" json:"id"`
	UserID             uuid.UUID `gorm:"index" json:"user_id"`
	CreatedAt          time.Time `gorm:"index" json:"created_at"`
	ApplyAt            time.Time `gorm:"index" json:"apply_at"`
	Amount             int64     `json:"amount"`
	Description        string    `gorm:"size:255" json:"description"`
	BlockTransactionID uint64    `gorm:"index" json:"block_transaction_id"`
	ApplyTransactionID uint64    `gorm:"index" json:"apply_transaction_id"`
}
