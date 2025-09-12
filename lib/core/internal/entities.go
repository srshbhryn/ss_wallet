package internal

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	UserID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	AvailableBalance int64     `gorm:"not null;default:0"`
	BlockedBalance   int64     `gorm:"not null;default:0"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Transaction struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"`
	WalletID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Amount        int64     `gorm:"not null;default:0"`       // effect on available balance
	BlockedAmount int64     `gorm:"not null;default:0"`       // effect on blocked funds
	Reference     uuid.UUID `gorm:"type:uuid;index;not null"` // external reference
	Description   string    `gorm:"size:255"`
	CreatedAt     time.Time
}
