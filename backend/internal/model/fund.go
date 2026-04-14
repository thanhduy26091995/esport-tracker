package model

import (
	"time"

	"github.com/google/uuid"
)

type FundTransaction struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Amount              int        `gorm:"not null" json:"amount"` // Amount in VND
	TransactionType     string     `gorm:"type:varchar(20);not null" json:"transaction_type"` // "deposit" or "withdrawal"
	Description         string     `gorm:"type:text" json:"description"`
	RelatedSettlementID *uuid.UUID `gorm:"type:uuid" json:"related_settlement_id,omitempty"`
	TransactionDate     time.Time  `gorm:"default:now()" json:"transaction_date"`
	CreatedAt           time.Time  `json:"created_at"`
}

func (FundTransaction) TableName() string {
	return "fund_transactions"
}
