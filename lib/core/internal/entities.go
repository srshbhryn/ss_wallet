package internal

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	UserID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	AvailableBalance int64     `gorm:"not null;default:0" json:"available_balance"`
	BlockedBalance   int64     `gorm:"not null;default:0" json:"blocked_balance"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Transaction struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	WalletID      uuid.UUID `gorm:"type:uuid;not null;index" json:"wallet_id"`
	Amount        int64     `gorm:"not null;default:0" json:"amount"`
	BlockedAmount int64     `gorm:"not null;default:0" json:"blocked_amount"`
	Reference     uuid.UUID `gorm:"type:uuid;index;not null" json:"reference"`
	Description   string    `gorm:"size:255" json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}
