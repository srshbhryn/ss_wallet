package internal

import (
	"time"

	"github.com/google/uuid"
)

type Deposit struct {
	ID                 uuid.UUID `gorm:"primaryKey"`
	UserID             uuid.UUID `gorm:"index"`
	CreatedAt          time.Time `gorm:"index"`
	ApplyAt            time.Time `gorm:"index"`
	Amount             int64
	Description        string `gorm:"size:255"`
	BlockTransactionID uint64 `gorm:"index"`
	ApplyTransactionID uint64 `gorm:"index"`
}
