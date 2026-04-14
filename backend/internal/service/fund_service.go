package service

import (
	"errors"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
)

type FundService struct {
	fundRepo *repository.FundRepository
}

func NewFundService(fundRepo *repository.FundRepository) *FundService {
	return &FundService{fundRepo: fundRepo}
}

// CreateDepositRequest represents deposit request
type CreateDepositRequest struct {
	Amount      int        `json:"amount" binding:"required"`
	Description string     `json:"description"`
	Date        *time.Time `json:"date,omitempty"`
}

// CreateWithdrawalRequest represents withdrawal request
type CreateWithdrawalRequest struct {
	Amount      int        `json:"amount" binding:"required"`
	Description string     `json:"description"`
	Date        *time.Time `json:"date,omitempty"`
}

// CreateDeposit creates a deposit transaction
func (s *FundService) CreateDeposit(req *CreateDepositRequest) (*model.FundTransaction, error) {
	if req.Amount <= 0 {
		return nil, errors.New("deposit amount must be positive")
	}

	transactionDate := time.Now()
	if req.Date != nil {
		transactionDate = *req.Date
	}

	transaction := &model.FundTransaction{
		Amount:          req.Amount,
		TransactionType: "deposit",
		Description:     req.Description,
		TransactionDate: transactionDate,
	}

	if err := s.fundRepo.Create(transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

// CreateWithdrawal creates a withdrawal transaction
func (s *FundService) CreateWithdrawal(req *CreateWithdrawalRequest) (*model.FundTransaction, error) {
	if req.Amount <= 0 {
		return nil, errors.New("withdrawal amount must be positive")
	}

	// Check if there's enough balance
	balance, err := s.fundRepo.GetBalance()
	if err != nil {
		return nil, err
	}

	if balance < req.Amount {
		return nil, errors.New("insufficient fund balance")
	}

	transactionDate := time.Now()
	if req.Date != nil {
		transactionDate = *req.Date
	}

	transaction := &model.FundTransaction{
		Amount:          req.Amount,
		TransactionType: "withdrawal",
		Description:     req.Description,
		TransactionDate: transactionDate,
	}

	if err := s.fundRepo.Create(transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

// CreateSettlementDeposit creates a deposit from debt settlement (internal use)
func (s *FundService) CreateSettlementDeposit(amount int, description string) error {
	if amount <= 0 {
		return errors.New("settlement amount must be positive")
	}

	transaction := &model.FundTransaction{
		Amount:          amount,
		TransactionType: "deposit",
		Description:     description,
		TransactionDate: time.Now(),
	}

	return s.fundRepo.Create(transaction)
}

// GetAllTransactions returns all fund transactions
func (s *FundService) GetAllTransactions(limit, offset int) ([]*model.FundTransaction, error) {
	return s.fundRepo.GetAll(limit, offset)
}

// GetBalance returns the current fund balance
func (s *FundService) GetBalance() (int, error) {
	return s.fundRepo.GetBalance()
}

// GetTransactionsByType returns transactions by type
func (s *FundService) GetTransactionsByType(transactionType string, limit int) ([]*model.FundTransaction, error) {
	if transactionType != "deposit" && transactionType != "withdrawal" {
		return nil, errors.New("transaction_type must be 'deposit' or 'withdrawal'")
	}
	return s.fundRepo.GetByType(transactionType, limit)
}

// GetFundStats returns fund statistics
func (s *FundService) GetFundStats() (map[string]interface{}, error) {
	balance, err := s.fundRepo.GetBalance()
	if err != nil {
		return nil, err
	}

	totalTransactions, err := s.fundRepo.CountTotal()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"current_balance":     balance,
		"total_transactions":  totalTransactions,
	}, nil
}
