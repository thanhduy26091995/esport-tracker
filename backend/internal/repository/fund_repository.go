package repository

import (
	"github.com/duyb/esport-score-tracker/internal/model"
	"gorm.io/gorm"
)

type FundRepository struct {
	db *gorm.DB
}

func NewFundRepository(db *gorm.DB) *FundRepository {
	return &FundRepository{db: db}
}

// Create creates a new fund transaction
func (r *FundRepository) Create(transaction *model.FundTransaction) error {
	return r.db.Create(transaction).Error
}

// GetAll returns all fund transactions ordered by date DESC
func (r *FundRepository) GetAll(limit, offset int) ([]*model.FundTransaction, error) {
	var transactions []*model.FundTransaction
	query := r.db.Order("transaction_date DESC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&transactions).Error
	return transactions, err
}

// GetBalance calculates the current fund balance
func (r *FundRepository) GetBalance() (int, error) {
	var balance int
	err := r.db.Model(&model.FundTransaction{}).
		Select("COALESCE(SUM(CASE WHEN transaction_type = 'deposit' THEN amount WHEN transaction_type = 'withdrawal' THEN -amount END), 0)").
		Scan(&balance).Error
	return balance, err
}

// GetByType returns transactions by type
func (r *FundRepository) GetByType(transactionType string, limit int) ([]*model.FundTransaction, error) {
	var transactions []*model.FundTransaction
	query := r.db.Where("transaction_type = ?", transactionType).
		Order("transaction_date DESC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&transactions).Error
	return transactions, err
}

// CountTotal returns the total number of transactions
func (r *FundRepository) CountTotal() (int64, error) {
	var count int64
	err := r.db.Model(&model.FundTransaction{}).Count(&count).Error
	return count, err
}

// GetTotalByType returns the sum of amounts for a given transaction type
func (r *FundRepository) GetTotalByType(transactionType string) (int, error) {
	var total int
	err := r.db.Model(&model.FundTransaction{}).
		Where("transaction_type = ?", transactionType).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// CountSettlementDeposits returns the count of deposits linked to a settlement
func (r *FundRepository) CountSettlementDeposits() (int64, error) {
	var count int64
	err := r.db.Model(&model.FundTransaction{}).
		Where("transaction_type = 'deposit' AND related_settlement_id IS NOT NULL").
		Count(&count).Error
	return count, err
}
