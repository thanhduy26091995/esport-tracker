package database

import (
	"fmt"
	"log"
	"os"

	"github.com/duyb/esport-score-tracker/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run schema migrations that GORM AutoMigrate cannot handle (column type changes)
	if err := runSchemaMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run schema migrations: %w", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&model.User{},
		&model.Match{},
		&model.MatchParticipant{},
		&model.DebtSettlement{},
		&model.SettlementWinner{},
		&model.FundTransaction{},
		&model.Config{},
		&model.Tournament{},
		&model.TournamentParticipant{},
		&model.TournamentTeam{},
		&model.TournamentMatch{},
		// WC2026 models
		&model.WcUser{},
		&model.WcConfig{},
		&model.WcMatch{},
		&model.WcScoreMultiplier{},
		&model.WcScoreOdds{},
		&model.WcWallet{},
		&model.WcWalletLog{},
		&model.WcPrediction{},
		&model.WcBet{},
		&model.WcSettlement{},
		&model.WcSettlementDetail{},
		&model.WcSyncLog{},
		// Champion prediction
		&model.WcChampionTeam{},
		&model.WcChampionConfig{},
		&model.WcChampionPrediction{},
		&model.ScoreBonus{},
		// Custom proposition bets
		&model.WcCustomBet{},
		&model.WcCustomBetOption{},
		&model.WcCustomBetEntry{},
		// Live chat
		&model.WcChatMessage{},
		&model.WcChatMention{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Seed initial config values if not exists
	seedConfig(db)
	seedWcConfig(db)
	seedWcChampion(db)

	log.Println("✅ Database connected successfully")
	return db, nil
}

func runSchemaMigrations(db *gorm.DB) error {
	sqls := []string{
		// tournament_round_robin_top4: add format + champion to tournaments
		`ALTER TABLE tournaments ADD COLUMN IF NOT EXISTS format VARCHAR(30) NOT NULL DEFAULT 'classic'`,
		`ALTER TABLE tournaments ADD COLUMN IF NOT EXISTS champion_team_id UUID`,
		`ALTER TABLE tournaments ADD COLUMN IF NOT EXISTS knockout_size INT NOT NULL DEFAULT 4`,
		// tournament_round_robin_top4: add stage + team FK columns to tournament_matches
		`ALTER TABLE tournament_matches ADD COLUMN IF NOT EXISTS stage VARCHAR(20) NOT NULL DEFAULT 'group'`,
		`ALTER TABLE tournament_matches ADD COLUMN IF NOT EXISTS team1_team_id UUID`,
		`ALTER TABLE tournament_matches ADD COLUMN IF NOT EXISTS team2_team_id UUID`,
		`ALTER TABLE wc_bets ALTER COLUMN payout TYPE NUMERIC(10,2)`,
		`ALTER TABLE wc_wallets ALTER COLUMN balance TYPE NUMERIC(10,2)`,
		`ALTER TABLE wc_wallet_logs ALTER COLUMN delta TYPE NUMERIC(10,2)`,
		`ALTER TABLE wc_wallet_logs ALTER COLUMN balance_before TYPE NUMERIC(10,2)`,
		`ALTER TABLE wc_wallet_logs ALTER COLUMN balance_after TYPE NUMERIC(10,2)`,
		`ALTER TABLE wc_settlement_details ALTER COLUMN final_balance TYPE NUMERIC(10,2)`,
		// Float points: allow fractional points_earned for win_half / multiplier precision
		`ALTER TABLE wc_predictions ALTER COLUMN points_earned TYPE NUMERIC(10,2) USING points_earned::numeric`,
		// Google OAuth: make password_hash nullable for Google-only accounts
		`ALTER TABLE wc_users ALTER COLUMN password_hash DROP NOT NULL`,
		// Partial unique index so NULL google_id values don't conflict
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_wc_users_google_id ON wc_users (google_id) WHERE google_id IS NOT NULL`,
		// Champion multi-pick: drop old per-user unique constraint, add (user, team) composite
		`DROP INDEX IF EXISTS idx_wc_champion_predictions_wc_user_id`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_wc_champion_pred_user_team ON wc_champion_predictions (wc_user_id, team_id)`,
		// Configurable bet limits
		`ALTER TABLE wc_config ADD COLUMN IF NOT EXISTS min_points INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE wc_config ADD COLUMN IF NOT EXISTS max_points INTEGER NOT NULL DEFAULT 5`,
		// Fix wc_predictions.handicap_snapshot precision: numeric(4,1) rounded 0.25→0.3, breaking quarter-ball detection.
		// USING clause rounds existing data to nearest 0.25 (e.g. 0.3→0.25, 0.8→0.75) before widening type.
		`ALTER TABLE wc_predictions ALTER COLUMN handicap_snapshot TYPE NUMERIC(5,2) USING ROUND(handicap_snapshot::numeric * 4) / 4.0`,
		// Fix wc_matches.ou_line precision: numeric(4,1) rounded 2.75→2.8, breaking quarter-ball O/U settlement.
		// Same rounding recovery as handicap_snapshot above (e.g. 2.8→2.75, 2.3→2.25).
		`ALTER TABLE wc_matches ALTER COLUMN ou_line TYPE NUMERIC(4,2) USING ROUND(ou_line::numeric * 4) / 4.0`,
		`ALTER TABLE wc_bets ALTER COLUMN handicap_snapshot TYPE NUMERIC(5,2) USING ROUND(handicap_snapshot::numeric * 4) / 4.0`,
		// Bet cancel penalty: soft-delete + penalty tracking
		`ALTER TABLE wc_bets ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ NULL`,
		`ALTER TABLE wc_bets ADD COLUMN IF NOT EXISTS cancel_penalty INTEGER NULL`,
		`ALTER TABLE wc_bets ADD COLUMN IF NOT EXISTS original_stake INTEGER NULL`,
		`ALTER TABLE wc_custom_bet_entries ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ NULL`,
		`ALTER TABLE wc_custom_bet_entries ADD COLUMN IF NOT EXISTS cancel_penalty INTEGER NULL`,
		`ALTER TABLE wc_custom_bet_entries ADD COLUMN IF NOT EXISTS original_stake INTEGER NULL`,
		`ALTER TABLE wc_config ADD COLUMN IF NOT EXISTS cancel_penalty_enabled BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE wc_config ADD COLUMN IF NOT EXISTS cancel_penalty_percent INTEGER NOT NULL DEFAULT 20`,
		`ALTER TABLE wc_config ADD COLUMN IF NOT EXISTS bet_reduce_max_percent INTEGER NOT NULL DEFAULT 50`,
		`ALTER TABLE wc_config ADD COLUMN IF NOT EXISTS bet_reduce_penalty_percent INTEGER NOT NULL DEFAULT 20`,
		// Prediction cancel penalty: soft-delete + penalty tracking
		`ALTER TABLE wc_predictions ADD COLUMN IF NOT EXISTS original_points INTEGER NULL`,
		`ALTER TABLE wc_predictions ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ NULL`,
		`ALTER TABLE wc_predictions ADD COLUMN IF NOT EXISTS cancel_penalty INTEGER NULL`,
		`ALTER TABLE wc_predictions ADD COLUMN IF NOT EXISTS reduce_penalty NUMERIC(10,2) NOT NULL DEFAULT 0`,
		// Change cancel_penalty from INTEGER to NUMERIC(10,2) to store float penalties
		`ALTER TABLE wc_predictions ALTER COLUMN cancel_penalty TYPE NUMERIC(10,2) USING cancel_penalty::numeric`,
		`ALTER TABLE wc_bets ALTER COLUMN cancel_penalty TYPE NUMERIC(10,2) USING cancel_penalty::numeric`,
		`ALTER TABLE wc_custom_bet_entries ALTER COLUMN cancel_penalty TYPE NUMERIC(10,2) USING cancel_penalty::numeric`,
		// Replace standard unique indexes with partial ones so cancelled predictions don't block re-placement
		`DROP INDEX IF EXISTS idx_prediction_hc_dedup`,
		`DROP INDEX IF EXISTS idx_prediction_es_dedup`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_prediction_hc_dedup ON wc_predictions(wc_user_id, match_id, prediction_type, prediction_choice) WHERE cancelled_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_prediction_es_dedup ON wc_predictions(wc_user_id, match_id, predicted_home_score, predicted_away_score) WHERE cancelled_at IS NULL`,
		// Bot user flag
		`ALTER TABLE wc_users ADD COLUMN IF NOT EXISTS is_bot BOOLEAN NOT NULL DEFAULT FALSE`,
	}
	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("Schema migration skipped (may not exist yet or already applied): %v", err)
		}
	}
	return nil
}

func seedWcConfig(db *gorm.DB) {
	// Ensure single wc_config row exists (is_enabled = false by default)
	var cfg model.WcConfig
	if err := db.First(&cfg, 1).Error; err != nil {
		db.Create(&model.WcConfig{ID: 1, IsEnabled: false})
		log.Println("Seeded wc_config: is_enabled = false")
	}

	// Seed first WC admin from env vars (optional — skip if not set)
	adminName := os.Getenv("WC_ADMIN_NAME")
	adminPassword := os.Getenv("WC_ADMIN_PASSWORD")
	if adminName != "" && adminPassword != "" {
		var existing model.WcUser
		if err := db.Where("name = ?", adminName).First(&existing).Error; err != nil {
			hash, hashErr := bcrypt.GenerateFromPassword([]byte(adminPassword), 12)
			if hashErr != nil {
				log.Printf("⚠️  Failed to hash WC admin password: %v", hashErr)
				return
			}
			hashStr := string(hash)
			db.Create(&model.WcUser{
				Name:         adminName,
				PasswordHash: &hashStr,
				IsAdmin:      true,
			})
			log.Printf("Seeded WC admin user: %s", adminName)
		}
	}
}

func seedWcChampion(db *gorm.DB) {
	// Singleton champion config
	var cfg model.WcChampionConfig
	if err := db.First(&cfg, 1).Error; err != nil {
		db.Create(&model.WcChampionConfig{ID: 1, IsOpen: false})
		log.Println("Seeded wc_champion_config: is_open = false")
	}

	// Seed teams only if table is empty
	var count int64
	db.Model(&model.WcChampionTeam{}).Count(&count)
	if count > 0 {
		return
	}
	// WC 2026 — 48 teams, odds tiered by strength.
	// sum(1/odds) ≈ 1.05 (slight house edge). Admin can update any odds via API.
	teams := []model.WcChampionTeam{
		// Tier 1 — Favourites (3–6x)
		{Name: "Argentina", Code: "ARG", FlagEmoji: "🇦🇷", Odds: 3.50},
		{Name: "France", Code: "FRA", FlagEmoji: "🇫🇷", Odds: 4.00},
		{Name: "Brazil", Code: "BRA", FlagEmoji: "🇧🇷", Odds: 4.50},
		{Name: "England", Code: "ENG", FlagEmoji: "🏴󠁧󠁢󠁥󠁮󠁧󠁿", Odds: 5.00},
		{Name: "Spain", Code: "ESP", FlagEmoji: "🇪🇸", Odds: 5.00},
		// Tier 2 — Strong (6–12x)
		{Name: "Germany", Code: "GER", FlagEmoji: "🇩🇪", Odds: 7.00},
		{Name: "Portugal", Code: "POR", FlagEmoji: "🇵🇹", Odds: 8.00},
		{Name: "Netherlands", Code: "NED", FlagEmoji: "🇳🇱", Odds: 9.00},
		{Name: "Colombia", Code: "COL", FlagEmoji: "🇨🇴", Odds: 10.00},
		{Name: "Uruguay", Code: "URU", FlagEmoji: "🇺🇾", Odds: 12.00},
		// Tier 3 — Dark horses (12–25x)
		{Name: "Belgium", Code: "BEL", FlagEmoji: "🇧🇪", Odds: 15.00},
		{Name: "Morocco", Code: "MAR", FlagEmoji: "🇲🇦", Odds: 15.00},
		{Name: "USA", Code: "USA", FlagEmoji: "🇺🇸", Odds: 18.00},
		{Name: "Japan", Code: "JPN", FlagEmoji: "🇯🇵", Odds: 20.00},
		{Name: "Croatia", Code: "CRO", FlagEmoji: "🇭🇷", Odds: 20.00},
		{Name: "Italy", Code: "ITA", FlagEmoji: "🇮🇹", Odds: 20.00},
		{Name: "Denmark", Code: "DEN", FlagEmoji: "🇩🇰", Odds: 22.00},
		{Name: "Switzerland", Code: "SUI", FlagEmoji: "🇨🇭", Odds: 22.00},
		{Name: "Senegal", Code: "SEN", FlagEmoji: "🇸🇳", Odds: 25.00},
		// Tier 4 — Mid (25–55x)
		{Name: "Mexico", Code: "MEX", FlagEmoji: "🇲🇽", Odds: 25.00},
		{Name: "Canada", Code: "CAN", FlagEmoji: "🇨🇦", Odds: 30.00},
		{Name: "South Korea", Code: "KOR", FlagEmoji: "🇰🇷", Odds: 30.00},
		{Name: "Australia", Code: "AUS", FlagEmoji: "🇦🇺", Odds: 30.00},
		{Name: "Turkey", Code: "TUR", FlagEmoji: "🇹🇷", Odds: 30.00},
		{Name: "Austria", Code: "AUT", FlagEmoji: "🇦🇹", Odds: 35.00},
		{Name: "Serbia", Code: "SRB", FlagEmoji: "🇷🇸", Odds: 35.00},
		{Name: "Ecuador", Code: "ECU", FlagEmoji: "🇪🇨", Odds: 40.00},
		{Name: "Ivory Coast", Code: "CIV", FlagEmoji: "🇨🇮", Odds: 40.00},
		{Name: "Iran", Code: "IRN", FlagEmoji: "🇮🇷", Odds: 45.00},
		{Name: "Egypt", Code: "EGY", FlagEmoji: "🇪🇬", Odds: 45.00},
		{Name: "Nigeria", Code: "NGA", FlagEmoji: "🇳🇬", Odds: 50.00},
		// Tier 5 — Underdogs (55–120x)
		{Name: "Scotland", Code: "SCO", FlagEmoji: "🏴󠁧󠁢󠁳󠁣󠁴󠁿", Odds: 55.00},
		{Name: "Hungary", Code: "HUN", FlagEmoji: "🇭🇺", Odds: 55.00},
		{Name: "Paraguay", Code: "PAR", FlagEmoji: "🇵🇾", Odds: 60.00},
		{Name: "South Africa", Code: "RSA", FlagEmoji: "🇿🇦", Odds: 60.00},
		{Name: "Ghana", Code: "GHA", FlagEmoji: "🇬🇭", Odds: 65.00},
		{Name: "Cameroon", Code: "CMR", FlagEmoji: "🇨🇲", Odds: 65.00},
		{Name: "Saudi Arabia", Code: "KSA", FlagEmoji: "🇸🇦", Odds: 70.00},
		{Name: "DR Congo", Code: "COD", FlagEmoji: "🇨🇩", Odds: 75.00},
		{Name: "Mali", Code: "MLI", FlagEmoji: "🇲🇱", Odds: 75.00},
		{Name: "Algeria", Code: "ALG", FlagEmoji: "🇩🇿", Odds: 75.00},
		{Name: "Panama", Code: "PAN", FlagEmoji: "🇵🇦", Odds: 80.00},
		{Name: "Honduras", Code: "HON", FlagEmoji: "🇭🇳", Odds: 90.00},
		{Name: "Iraq", Code: "IRQ", FlagEmoji: "🇮🇶", Odds: 90.00},
		{Name: "Uzbekistan", Code: "UZB", FlagEmoji: "🇺🇿", Odds: 100.00},
		{Name: "New Zealand", Code: "NZL", FlagEmoji: "🇳🇿", Odds: 100.00},
		{Name: "Jamaica", Code: "JAM", FlagEmoji: "🇯🇲", Odds: 110.00},
		{Name: "Venezuela", Code: "VEN", FlagEmoji: "🇻🇪", Odds: 110.00},
		{Name: "Tunisia", Code: "TUN", FlagEmoji: "🇹🇳", Odds: 120.00},
	}
	if err := db.Create(&teams).Error; err != nil {
		log.Printf("⚠️  Failed to seed champion teams: %v", err)
	} else {
		log.Printf("Seeded %d WC champion teams", len(teams))
	}
}

func seedConfig(db *gorm.DB) {
	configs := []model.Config{
		{Key: "debt_threshold", Value: "-6", Description: "Score threshold that triggers debt settlement"},
		{Key: "point_to_vnd", Value: "22000", Description: "Conversion rate: 1 point = X VND"},
		{Key: "fund_split_percent", Value: "50", Description: "Percentage of debt that goes to fund (rest to winners)"},
		{Key: "auto_settlement", Value: "false", Description: "Automatically trigger settlement when debt threshold is reached (true/false)"},
		{Key: "min_matches_for_tier", Value: "5", Description: "Minimum matches a player must play before tier evaluation applies"},
		{Key: "pro_win_rate_threshold", Value: "0.60", Description: "Win rate threshold (0-1) required to reach the Pro tier"},
		{Key: "normal_win_rate_threshold", Value: "0.40", Description: "Win rate threshold (0-1) required to reach the Normal tier (below this = Noob)"},
	}

	for _, cfg := range configs {
		var existing model.Config
		if err := db.Where("key = ?", cfg.Key).First(&existing).Error; err != nil {
			// Config doesn't exist, create it
			db.Create(&cfg)
			log.Printf("Seeded config: %s = %s", cfg.Key, cfg.Value)
		}
	}
}
