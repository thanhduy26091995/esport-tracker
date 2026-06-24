package repository

import (
	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WcCustomBetRepository struct {
	db *gorm.DB
}

func NewWcCustomBetRepository(db *gorm.DB) *WcCustomBetRepository {
	return &WcCustomBetRepository{db: db}
}

func (r *WcCustomBetRepository) Create(bet *model.WcCustomBet, options []model.WcCustomBetOption) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(bet).Error; err != nil {
			return err
		}
		for i := range options {
			options[i].CustomBetID = bet.ID
		}
		return tx.Create(&options).Error
	})
}

func (r *WcCustomBetRepository) GetByID(id uuid.UUID) (*model.WcCustomBet, error) {
	var bet model.WcCustomBet
	err := r.db.Where("id = ?", id).First(&bet).Error
	return &bet, err
}

func (r *WcCustomBetRepository) GetOptions(betID uuid.UUID) ([]model.WcCustomBetOption, error) {
	var options []model.WcCustomBetOption
	err := r.db.Where("custom_bet_id = ?", betID).Order("display_order ASC").Find(&options).Error
	return options, err
}

func (r *WcCustomBetRepository) GetOptionByID(optionID uuid.UUID) (*model.WcCustomBetOption, error) {
	var opt model.WcCustomBetOption
	err := r.db.Where("id = ?", optionID).First(&opt).Error
	return &opt, err
}

func (r *WcCustomBetRepository) listForMatch(matchID uuid.UUID) ([]*model.WcCustomBet, error) {
	var bets []*model.WcCustomBet
	err := r.db.Where("match_id = ?", matchID).Order("created_at ASC").Find(&bets).Error
	return bets, err
}

func (r *WcCustomBetRepository) GetPublicEntriesForBet(betID uuid.UUID) ([]model.WcCustomBetEntryPublic, error) {
	var result []model.WcCustomBetEntryPublic
	err := r.db.Table("wc_custom_bet_entries e").
		Select("e.id, e.wc_user_id, e.option_id, o.label AS option_label, u.name, u.avatar_url, e.stake, e.odds_snapshot, e.status, e.payout, e.created_at").
		Joins("JOIN wc_users u ON u.id = e.wc_user_id").
		Joins("JOIN wc_custom_bet_options o ON o.id = e.option_id").
		Where("e.custom_bet_id = ?", betID).
		Order("e.created_at ASC").
		Scan(&result).Error
	return result, err
}

func (r *WcCustomBetRepository) listWithOptionsAndEntries(bets []*model.WcCustomBet) (
	optsByBet map[uuid.UUID][]model.WcCustomBetOption,
	entriesByBet map[uuid.UUID][]model.WcCustomBetEntryPublic,
	err error,
) {
	if len(bets) == 0 {
		return map[uuid.UUID][]model.WcCustomBetOption{}, map[uuid.UUID][]model.WcCustomBetEntryPublic{}, nil
	}
	ids := make([]uuid.UUID, len(bets))
	for i, b := range bets {
		ids[i] = b.ID
	}

	var allOpts []model.WcCustomBetOption
	if err = r.db.Where("custom_bet_id IN ?", ids).Order("display_order ASC").Find(&allOpts).Error; err != nil {
		return
	}

	type entryWithBetID struct {
		model.WcCustomBetEntryPublic
		CustomBetID uuid.UUID `json:"-"`
	}
	var allEntries []entryWithBetID
	if err = r.db.Table("wc_custom_bet_entries e").
		Select("e.custom_bet_id, e.id, e.wc_user_id, e.option_id, o.label AS option_label, u.name, u.avatar_url, e.stake, e.odds_snapshot, e.status, e.payout, e.created_at").
		Joins("JOIN wc_users u ON u.id = e.wc_user_id").
		Joins("JOIN wc_custom_bet_options o ON o.id = e.option_id").
		Where("e.custom_bet_id IN ?", ids).
		Order("e.created_at ASC").
		Scan(&allEntries).Error; err != nil {
		return
	}

	optsByBet = make(map[uuid.UUID][]model.WcCustomBetOption, len(bets))
	for _, o := range allOpts {
		optsByBet[o.CustomBetID] = append(optsByBet[o.CustomBetID], o)
	}
	entriesByBet = make(map[uuid.UUID][]model.WcCustomBetEntryPublic, len(bets))
	for _, e := range allEntries {
		entriesByBet[e.CustomBetID] = append(entriesByBet[e.CustomBetID], e.WcCustomBetEntryPublic)
	}
	return
}

func (r *WcCustomBetRepository) ListForMatchAdmin(matchID uuid.UUID) ([]*model.WcCustomBetWithOptions, error) {
	bets, err := r.listForMatch(matchID)
	if err != nil {
		return nil, err
	}
	optsByBet, entriesByBet, err := r.listWithOptionsAndEntries(bets)
	if err != nil {
		return nil, err
	}
	result := make([]*model.WcCustomBetWithOptions, 0, len(bets))
	for _, bet := range bets {
		entries := entriesByBet[bet.ID]
		result = append(result, &model.WcCustomBetWithOptions{
			WcCustomBet: *bet,
			Options:     optsByBet[bet.ID],
			EntryCount:  len(entries),
			Entries:     entries,
		})
	}
	return result, nil
}

func (r *WcCustomBetRepository) ListForMatchWithMyEntry(matchID, userID uuid.UUID) ([]*model.WcCustomBetWithOptions, error) {
	bets, err := r.listForMatch(matchID)
	if err != nil {
		return nil, err
	}
	optsByBet, entriesByBet, err := r.listWithOptionsAndEntries(bets)
	if err != nil {
		return nil, err
	}
	result := make([]*model.WcCustomBetWithOptions, 0, len(bets))
	for _, bet := range bets {
		entries := entriesByBet[bet.ID]
		var myEntry *model.WcCustomBetEntry
		for i := range entries {
			if entries[i].WcUserID == userID {
				e := entries[i]
				myEntry = &model.WcCustomBetEntry{
					ID:           e.ID,
					CustomBetID:  bet.ID,
					OptionID:     e.OptionID,
					WcUserID:     e.WcUserID,
					Stake:        e.Stake,
					OddsSnapshot: e.OddsSnapshot,
					Payout:       e.Payout,
					Status:       e.Status,
					CreatedAt:    e.CreatedAt,
				}
				break
			}
		}
		result = append(result, &model.WcCustomBetWithOptions{
			WcCustomBet: *bet,
			Options:     optsByBet[bet.ID],
			MyEntry:     myEntry,
			EntryCount:  len(entries),
			Entries:     entries,
		})
	}
	return result, nil
}

// ListCustomEntriesForMatchPublic returns all custom bet entries for a match
// adapted to WcPredictionPublic so they can be merged into the predictions list.
func (r *WcCustomBetRepository) ListCustomEntriesForMatchPublic(matchID uuid.UUID) ([]*model.WcPredictionPublic, error) {
	var result []*model.WcPredictionPublic
	err := r.db.Table("wc_custom_bet_entries e").
		Select(`e.id, e.wc_user_id, u.name, u.avatar_url,
		        'custom'                                      AS prediction_type,
		        o.label                                       AS prediction_choice,
		        e.stake                                       AS points,
		        e.odds_snapshot                               AS multiplier_snapshot,
		        CASE e.status
		            WHEN 'won'  THEN 'correct'
		            WHEN 'lost' THEN 'incorrect'
		            WHEN 'void' THEN 'void'
		            ELSE NULL
		        END                                           AS result,
		        CASE WHEN e.status = 'won' THEN e.payout - e.stake ELSE NULL END AS points_earned,
		        e.created_at,
		        b.title                                       AS bet_title`).
		Joins("JOIN wc_users u ON u.id = e.wc_user_id").
		Joins("JOIN wc_custom_bet_options o ON o.id = e.option_id").
		Joins("JOIN wc_custom_bets b ON b.id = e.custom_bet_id").
		Where("b.match_id = ?", matchID).
		Order("e.created_at ASC").
		Scan(&result).Error
	return result, err
}

// ListCustomEntriesForUserAsHistory returns a user's custom bet entries adapted
// to WcPredictionWithMatch so they merge cleanly into the predictions history list.
func (r *WcCustomBetRepository) ListCustomEntriesForUserAsHistory(userID uuid.UUID) ([]*model.WcPredictionWithMatch, error) {
	var result []*model.WcPredictionWithMatch
	err := r.db.Table("wc_custom_bet_entries e").
		Select(`e.id, e.wc_user_id, b.match_id,
		        'custom'                                             AS prediction_type,
		        o.label                                              AS prediction_choice,
		        b.title                                              AS bet_title,
		        e.stake                                              AS points,
		        e.odds_snapshot                                      AS multiplier_snapshot,
		        CASE e.status
		            WHEN 'won'  THEN 'correct'
		            WHEN 'lost' THEN 'incorrect'
		            WHEN 'void' THEN 'void'
		            ELSE NULL
		        END                                                  AS result,
		        CASE WHEN e.status = 'won' THEN ROUND(CAST(e.stake AS numeric) * e.odds_snapshot, 2) ELSE NULL END AS points_earned,
		        e.created_at, e.created_at AS updated_at,
		        m.home_team, m.away_team, m.match_date,
		        m.status                                             AS match_status,
		        false                                                AS predictions_open`).
		Joins("JOIN wc_custom_bets b ON b.id = e.custom_bet_id").
		Joins("JOIN wc_custom_bet_options o ON o.id = e.option_id").
		Joins("JOIN wc_matches m ON m.id = b.match_id").
		Where("e.wc_user_id = ?", userID).
		Order("e.created_at DESC").
		Scan(&result).Error
	return result, err
}

func (r *WcCustomBetRepository) GetEntriesForUser(userID uuid.UUID) ([]model.WcCustomBetEntryHistory, error) {
	var result []model.WcCustomBetEntryHistory
	err := r.db.Table("wc_custom_bet_entries e").
		Select(`e.*, b.title AS bet_title, b.line AS bet_line, o.label AS option_label,
		        m.home_team, m.away_team, m.match_date`).
		Joins("JOIN wc_custom_bets b ON b.id = e.custom_bet_id").
		Joins("JOIN wc_custom_bet_options o ON o.id = e.option_id").
		Joins("JOIN wc_matches m ON m.id = b.match_id").
		Where("e.wc_user_id = ?", userID).
		Order("e.created_at DESC").
		Scan(&result).Error
	return result, err
}

func (r *WcCustomBetRepository) CountUnsettledForMatch(matchID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.WcCustomBet{}).
		Where("match_id = ? AND status IN ('open', 'closed')", matchID).
		Count(&count).Error
	return count, err
}

func (r *WcCustomBetRepository) CountMatchesWithUnsettled() (int64, error) {
	var count int64
	err := r.db.Model(&model.WcCustomBet{}).
		Where("status IN ('open', 'closed')").
		Distinct("match_id").
		Count(&count).Error
	return count, err
}

func (r *WcCustomBetRepository) UpdateBet(tx *gorm.DB, betID uuid.UUID, fields map[string]interface{}) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&model.WcCustomBet{}).Where("id = ?", betID).Updates(fields).Error
}

func (r *WcCustomBetRepository) MarkOptionWinner(tx *gorm.DB, optionID uuid.UUID) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&model.WcCustomBetOption{}).Where("id = ?", optionID).Update("is_winner", true).Error
}

func (r *WcCustomBetRepository) GetPendingEntries(tx *gorm.DB, betID uuid.UUID) ([]model.WcCustomBetEntry, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var entries []model.WcCustomBetEntry
	err := db.Where("custom_bet_id = ? AND status = ?", betID, model.WcCustomBetEntryStatusPending).Find(&entries).Error
	return entries, err
}

func (r *WcCustomBetRepository) UpdateEntryResult(tx *gorm.DB, entryID uuid.UUID, status string, payout float64) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&model.WcCustomBetEntry{}).Where("id = ?", entryID).
		Updates(map[string]interface{}{"status": status, "payout": payout}).Error
}

func (r *WcCustomBetRepository) CreateEntry(tx *gorm.DB, entry *model.WcCustomBetEntry) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(entry).Error
}

func (r *WcCustomBetRepository) GetEntry(entryID uuid.UUID) (*model.WcCustomBetEntry, error) {
	var entry model.WcCustomBetEntry
	err := r.db.Where("id = ?", entryID).First(&entry).Error
	return &entry, err
}

func (r *WcCustomBetRepository) DeleteEntry(tx *gorm.DB, entryID uuid.UUID) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Delete(&model.WcCustomBetEntry{}, "id = ?", entryID).Error
}
