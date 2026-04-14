package repository

import (
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SettlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository(db *gorm.DB) *SettlementRepository {
	return &SettlementRepository{db: db}
}

// Create creates a new debt settlement
func (r *SettlementRepository) Create(settlement *model.DebtSettlement) error {
	return r.db.Create(settlement).Error
}

// GetByID returns a settlement by ID with debtor and winners
func (r *SettlementRepository) GetByID(id uuid.UUID) (*model.DebtSettlement, error) {
	var settlement model.DebtSettlement
	err := r.db.Preload("Debtor").
		Preload("Winners").
		First(&settlement, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &settlement, nil
}

// GetAll returns all settlements with pagination
func (r *SettlementRepository) GetAll(limit, offset int) ([]*model.DebtSettlement, error) {
	var settlements []*model.DebtSettlement
	query := r.db.Preload("Debtor").
		Preload("Winners").
		Order("settlement_date DESC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&settlements).Error
	return settlements, err
}

// GetByDebtorID returns all settlements for a specific debtor
func (r *SettlementRepository) GetByDebtorID(debtorID uuid.UUID, limit int) ([]*model.DebtSettlement, error) {
	var settlements []*model.DebtSettlement
	query := r.db.Preload("Debtor").
		Preload("Winners").
		Where("debtor_id = ?", debtorID).
		Order("settlement_date DESC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&settlements).Error
	return settlements, err
}

// CountTotal returns the total number of settlements
func (r *SettlementRepository) CountTotal() (int64, error) {
	var count int64
	err := r.db.Model(&model.DebtSettlement{}).Count(&count).Error
	return count, err
}

// CountToday returns the number of settlements today
func (r *SettlementRepository) CountToday() (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	err := r.db.Model(&model.DebtSettlement{}).
		Where("settlement_date >= ? AND settlement_date < ?", today, tomorrow).
		Count(&count).Error
	return count, err
}
