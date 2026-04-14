package model

import (
	"time"

	"github.com/google/uuid"
)

type DebtSettlement struct {
	ID                 uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DebtorID           uuid.UUID          `gorm:"type:uuid;not null" json:"debtor_id"`
	DebtAmount         int                `gorm:"not null" json:"debt_amount"` // Negative debt in points (e.g., -7)
	MoneyAmount        int                `gorm:"not null" json:"money_amount"` // Total VND amount
	FundAmount         int                `gorm:"not null" json:"fund_amount"`  // Amount to fund (ToFund alias)
	WinnerDistribution int                `gorm:"not null" json:"winner_distribution"` // Amount to winners (ToWinners alias)
	OriginalDebtPoints int                `gorm:"not null" json:"original_debt_points"` // Original debt in points (positive)
	SettlementDate     time.Time          `gorm:"default:now()" json:"settlement_date"`
	CreatedAt          time.Time          `json:"created_at"`
	Debtor             User               `gorm:"foreignKey:DebtorID" json:"debtor,omitempty"`
	Winners            []SettlementWinner `gorm:"foreignKey:SettlementID" json:"winners,omitempty"`
}

func (DebtSettlement) TableName() string {
	return "debt_settlements"
}

type SettlementWinner struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SettlementID   uuid.UUID `gorm:"type:uuid;not null" json:"settlement_id"`
	WinnerID       uuid.UUID `gorm:"type:uuid;not null" json:"winner_id"`
	MoneyAmount    int       `gorm:"not null" json:"money_amount"` // VND amount winner receives
	PointsDeducted int       `gorm:"not null" json:"points_deducted"` // Points deducted from winner
	Winner         User      `gorm:"foreignKey:WinnerID" json:"winner,omitempty"`
}

func (SettlementWinner) TableName() string {
	return "settlement_winners"
}
