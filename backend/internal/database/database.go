package database

import (
	"fmt"
	"log"
	"os"

	"github.com/duyb/esport-score-tracker/internal/model"
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
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Seed initial config values if not exists
	seedConfig(db)

	log.Println("✅ Database connected successfully")
	return db, nil
}

func seedConfig(db *gorm.DB) {
	configs := []model.Config{
		{Key: "debt_threshold", Value: "-6", Description: "Score threshold that triggers debt settlement"},
		{Key: "point_to_vnd", Value: "22000", Description: "Conversion rate: 1 point = X VND"},
		{Key: "fund_split_percent", Value: "50", Description: "Percentage of debt that goes to fund (rest to winners)"},
		{Key: "auto_settlement", Value: "false", Description: "Automatically trigger settlement when debt threshold is reached (true/false)"},
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
