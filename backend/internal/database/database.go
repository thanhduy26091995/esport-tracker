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
		&model.ScoreBonus{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Seed initial config values if not exists
	seedConfig(db)
	seedWcConfig(db)

	log.Println("✅ Database connected successfully")
	return db, nil
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
			db.Create(&model.WcUser{
				Name:         adminName,
				PasswordHash: string(hash),
				IsAdmin:      true,
			})
			log.Printf("Seeded WC admin user: %s", adminName)
		}
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
