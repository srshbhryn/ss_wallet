package internal

import (
	"time"

	"github.com/google/uuid"
)

type Deposit struct {
	ID                 uuid.UUID `gorm:"primaryKey"`
	CreatedAt          time.Time `gorm:"index"`
	ApplyAt            time.Time `gorm:"index"`
	Amount             int64
	Description        string    `gorm:"size:255"`
	BlockTransactionID uuid.UUID `gorm:"index"`
	ApplyTransactionID uuid.UUID `gorm:"index"`
}
