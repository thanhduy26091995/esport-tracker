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
		Preload("Winners.Winner").
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
		Preload("Winners.Winner").
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
		Preload("Winners.Winner").
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

// GetFundContributors returns per-user fund contribution totals, ordered by total_fund_amount desc
func (r *SettlementRepository) GetFundContributors() ([]*model.FundContributor, error) {
	var rows []*model.FundContributor
	err := r.db.Raw(`
		SELECT
			u.id          AS user_id,
			u.name        AS user_name,
			COUNT(ds.id)  AS settlement_count,
			SUM(ds.fund_amount) AS total_fund_amount
		FROM debt_settlements ds
		JOIN users u ON u.id = ds.debtor_id
		GROUP BY u.id, u.name
		ORDER BY total_fund_amount DESC
	`).Scan(&rows).Error
	return rows, err
}

// GetWinnerContributors aggregates each winner's fund contributions across all settlements.
// A winner's fund share per settlement = points_deducted * fund_amount / original_debt_points.
func (r *SettlementRepository) GetWinnerContributors() ([]*model.WinnerContributor, error) {
	var rows []*model.WinnerContributor
	err := r.db.Raw(`
		SELECT
			u.id                                                                                     AS user_id,
			u.name                                                                                   AS user_name,
			COUNT(DISTINCT sw.settlement_id)                                                         AS settlement_count,
			SUM(sw.points_deducted)                                                                  AS total_points_deducted,
			ROUND(SUM(sw.points_deducted::numeric * ds.fund_amount / NULLIF(ds.money_amount, 0)))::int    AS total_points_contributed,
			ROUND(SUM(sw.points_deducted::numeric * ds.fund_amount / NULLIF(ds.original_debt_points, 0)))::int AS total_fund_amount
		FROM settlement_winners sw
		JOIN users            u  ON u.id  = sw.winner_id
		JOIN debt_settlements ds ON ds.id = sw.settlement_id
		GROUP BY u.id, u.name
		ORDER BY total_fund_amount DESC
	`).Scan(&rows).Error
	return rows, err
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
