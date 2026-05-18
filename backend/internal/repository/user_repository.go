package repository

import (
	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

const winRateSelect = `u.*,
	COUNT(mp.id) FILTER (WHERE mp.point_change != 0)            AS total_matches,
	COUNT(mp.id) FILTER (WHERE mp.point_change > 0)             AS won_matches,
	CASE WHEN COUNT(mp.id) FILTER (WHERE mp.point_change != 0) = 0 THEN 0
	     ELSE COUNT(mp.id) FILTER (WHERE mp.point_change > 0)::float
	          / COUNT(mp.id) FILTER (WHERE mp.point_change != 0)
	END AS win_rate`

// GetAll returns all active users with computed win rate, ordered by current_score DESC.
func (r *UserRepository) GetAll() ([]*model.UserWithStats, error) {
	var users []*model.UserWithStats
	err := r.db.Table("users u").
		Select(winRateSelect).
		Joins("LEFT JOIN match_participants mp ON mp.user_id = u.id").
		Where("u.is_active = ?", true).
		Group("u.id").
		Order("u.current_score DESC, u.name ASC").
		Find(&users).Error
	return users, err
}

// GetByID returns a single active user with computed win rate.
func (r *UserRepository) GetByID(id uuid.UUID) (*model.UserWithStats, error) {
	var user model.UserWithStats
	err := r.db.Table("users u").
		Select(winRateSelect).
		Joins("LEFT JOIN match_participants mp ON mp.user_id = u.id").
		Where("u.id = ? AND u.is_active = ?", id, true).
		Group("u.id").
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByName returns a user by name (for duplicate checking)
func (r *UserRepository) GetByName(name string) (*model.User, error) {
	var user model.User
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update updates an existing user
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// SoftDelete marks a user as inactive (soft delete)
func (r *UserRepository) SoftDelete(id uuid.UUID) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// GetLeaderboard returns active users with computed win rate, optionally limited.
func (r *UserRepository) GetLeaderboard(limit int) ([]*model.UserWithStats, error) {
	var users []*model.UserWithStats
	query := r.db.Table("users u").
		Select(winRateSelect).
		Joins("LEFT JOIN match_participants mp ON mp.user_id = u.id").
		Where("u.is_active = ?", true).
		Group("u.id").
		Order("u.current_score DESC, u.name ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&users).Error
	return users, err
}

// GetWinRatesBatch returns win rate stats for a specific set of user IDs in one query.
func (r *UserRepository) GetWinRatesBatch(ids []uuid.UUID) (map[uuid.UUID]model.UserWithStats, error) {
	var rows []model.UserWithStats
	err := r.db.Table("users u").
		Select(winRateSelect).
		Joins("LEFT JOIN match_participants mp ON mp.user_id = u.id").
		Where("u.id IN ?", ids).
		Group("u.id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]model.UserWithStats, len(rows))
	for _, row := range rows {
		result[row.ID] = row
	}
	return result, nil
}

// UpdateTier persists a computed tier value for a user.
func (r *UserRepository) UpdateTier(id uuid.UUID, tier string) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", id).
		Update("tier", tier).Error
}

// GetAllIDs returns IDs of all active users (used for startup backfill).
func (r *UserRepository) GetAllIDs() ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&model.User{}).
		Where("is_active = ?", true).
		Pluck("id", &ids).Error
	return ids, err
}

// UpdateScore updates a user's current score
func (r *UserRepository) UpdateScore(id uuid.UUID, scoreChange int) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", id).
		UpdateColumn("current_score", gorm.Expr("current_score + ?", scoreChange)).Error
}

// GetPaymentRanking returns active users sorted by total historical settlement money paid DESC
func (r *UserRepository) GetPaymentRanking() ([]*model.UserWithPaymentTotal, error) {
	var results []*model.UserWithPaymentTotal
	err := r.db.Raw(`
		SELECT u.*,
		       COALESCE((
		           SELECT SUM(s.money_amount)
		           FROM debt_settlements s
		           WHERE s.debtor_id = u.id
		       ), 0) AS total_paid,
		       COALESCE((
		           SELECT SUM(sw.points_deducted)
		           FROM settlement_winners sw
		           JOIN debt_settlements ds ON sw.settlement_id = ds.id
		           WHERE ds.debtor_id = u.id
		       ), 0) AS total_debt_points,
		       COUNT(mp.id) FILTER (WHERE mp.point_change != 0)           AS total_matches,
		       COUNT(mp.id) FILTER (WHERE mp.point_change > 0)            AS won_matches,
		       CASE WHEN COUNT(mp.id) FILTER (WHERE mp.point_change != 0) = 0 THEN 0
		            ELSE COUNT(mp.id) FILTER (WHERE mp.point_change > 0)::float
		                 / COUNT(mp.id) FILTER (WHERE mp.point_change != 0)
		       END AS win_rate
		FROM users u
		LEFT JOIN match_participants mp ON mp.user_id = u.id
		WHERE u.is_active = true
		GROUP BY u.id
		ORDER BY total_paid DESC, u.name ASC
	`).Scan(&results).Error
	return results, err
}
