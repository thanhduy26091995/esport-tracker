package api

import (
	"log"
	"os"
	"strings"

	"github.com/duyb/esport-score-tracker/internal/middleware"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	corsConfig := cors.Config{
		AllowOrigins:     strings.Split(os.Getenv("CORS_ORIGINS"), ","),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}
	router.Use(cors.New(corsConfig))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "FC25 Esport Score Tracker API"})
	})

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	matchRepo := repository.NewMatchRepository(db)
	configRepo := repository.NewConfigRepository(db)
	fundRepo := repository.NewFundRepository(db)
	settlementRepo := repository.NewSettlementRepository(db)
	tournamentRepo := repository.NewTournamentRepository(db)
	wcRepo := repository.NewWcRepository(db)
	wcUserRepo := repository.NewWcUserRepository(db)

	// Initialize services
	configService := service.NewConfigService(configRepo)
	userService := service.NewUserService(userRepo, configService)
	fundService := service.NewFundService(fundRepo)
	settlementService := service.NewSettlementService(settlementRepo, userRepo, matchRepo, fundService, configService, db)
	tierService := service.NewTierService(userRepo, configService)
	matchService := service.NewMatchService(matchRepo, userRepo, settlementService, configService, tierService, db)
	tournamentService := service.NewTournamentService(tournamentRepo, userRepo, matchService, db)
	wcAuthService := service.NewWcAuthService(wcUserRepo, wcRepo)
	wcService := service.NewWcService(wcRepo, wcUserRepo)
	bonusRepo := repository.NewScoreBonusRepository(db)
	bonusService := service.NewScoreBonusService(bonusRepo, userRepo, tierService, db)

	// Backfill tiers from existing match history on startup.
	if err := tierService.RecalculateAllTiers(); err != nil {
		log.Printf("⚠️  Failed to backfill tiers on startup: %v", err)
	}

	// Initialize handlers
	userHandler := NewUserHandler(userService)
	matchHandler := NewMatchHandler(matchService, bonusService)
	bonusHandler := NewScoreBonusHandler(bonusService)
	configHandler := NewConfigHandler(configService, tierService)
	fundHandler := NewFundHandler(fundService)
	settlementHandler := NewSettlementHandler(settlementService)
	tournamentHandler := NewTournamentHandler(tournamentService)
	wcAuthHandler := NewWcAuthHandler(wcAuthService)
	wcHandler := NewWcHandler(wcService, wcAuthService)

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.GET("", userHandler.GetAll)                            // GET /api/v1/users
			users.POST("", userHandler.Create)                           // POST /api/v1/users
			users.GET("/leaderboard", userHandler.GetLeaderboard)        // GET /api/v1/users/leaderboard
			users.GET("/payment-ranking", userHandler.GetPaymentRanking) // GET /api/v1/users/payment-ranking
			users.GET("/:id", userHandler.GetByID)                       // GET /api/v1/users/:id
			users.GET("/:id/matches", matchHandler.GetByUserID) // GET /api/v1/users/:id/matches
			users.PUT("/:id", userHandler.Update)          // PUT /api/v1/users/:id
			users.DELETE("/:id", userHandler.Delete)       // DELETE /api/v1/users/:id
		}

		// Match routes
		matches := v1.Group("/matches")
		{
			matches.GET("", matchHandler.GetAll)           // GET /api/v1/matches
			matches.POST("", matchHandler.Create)          // POST /api/v1/matches
			matches.GET("/recent", matchHandler.GetRecent) // GET /api/v1/matches/recent
			matches.GET("/stats", matchHandler.GetStats)   // GET /api/v1/matches/stats
			matches.GET("/:id", matchHandler.GetByID)      // GET /api/v1/matches/:id
			matches.DELETE("/:id", matchHandler.Delete)    // DELETE /api/v1/matches/:id
		}

		// Config routes
		config := v1.Group("/config")
		{
			config.GET("", configHandler.GetAll)           // GET /api/v1/config
			config.PUT("", configHandler.UpdateAll)         // PUT /api/v1/config  (bulk)
			config.GET("/:key", configHandler.GetByKey)     // GET /api/v1/config/:key
			config.PUT("/:key", configHandler.Update)       // PUT /api/v1/config/:key
		}

		// Fund routes
		fund := v1.Group("/fund")
		{
			fund.GET("/balance", fundHandler.GetBalance)   // GET /api/v1/fund/balance
			fund.GET("/stats", fundHandler.GetStats)       // GET /api/v1/fund/stats
			fund.GET("/transactions", fundHandler.GetTransactions) // GET /api/v1/fund/transactions
			fund.POST("/deposit", fundHandler.CreateDeposit)       // POST /api/v1/fund/deposit
			fund.POST("/withdraw", fundHandler.CreateWithdrawal)  // POST /api/v1/fund/withdraw
		}

		// Settlement routes
		settlements := v1.Group("/settlements")
		{
			settlements.GET("", settlementHandler.GetAll)    // GET /api/v1/settlements
			settlements.POST("/trigger", settlementHandler.TriggerSettlement) // POST /api/v1/settlements/trigger
			settlements.GET("/stats", settlementHandler.GetStats)                   // GET /api/v1/settlements/stats
			settlements.GET("/fund-contributors", settlementHandler.GetFundContributors)     // GET /api/v1/settlements/fund-contributors
			settlements.GET("/winner-contributors", settlementHandler.GetWinnerContributors) // GET /api/v1/settlements/winner-contributors
			settlements.GET("/:id", settlementHandler.GetByID)                              // GET /api/v1/settlements/:id
		}

		// User settlement history
		v1.GET("/users/:id/settlements", settlementHandler.GetByDebtorID) // GET /api/v1/users/:id/settlements

		// Score bonus routes (POST/DELETE only; GET is merged into /matches)
		bonuses := v1.Group("/score-bonuses")
		{
			bonuses.POST("", bonusHandler.Create)      // POST /api/v1/score-bonuses
			bonuses.DELETE("/:id", bonusHandler.Delete) // DELETE /api/v1/score-bonuses/:id
		}

		// Tournament routes
		tournaments := v1.Group("/tournaments")
		{
			tournaments.GET("", tournamentHandler.GetAll)
			tournaments.POST("", tournamentHandler.Create)
			tournaments.GET("/:id", tournamentHandler.GetByID)
			tournaments.DELETE("/:id", tournamentHandler.Delete)
			tournaments.PUT("/:id/complete", tournamentHandler.Complete)
			tournaments.POST("/:id/matches/:matchId/result", tournamentHandler.RecordResult)
		}
	}

	// WC2026 routes — /api/v1/wc
	wc := router.Group("/api/v1/wc")
	{
		// Config endpoints — always accessible (exempt from feature flag)
		wc.GET("/config", wcHandler.GetPublicConfig)
		wc.GET("/admin/config", middleware.WcJWTMiddleware(wcAuthService), middleware.WcAdminMiddleware(), wcHandler.GetConfig)
		wc.PUT("/admin/config", middleware.WcJWTMiddleware(wcAuthService), middleware.WcAdminMiddleware(), wcHandler.UpdateConfig)

		// Matches — always accessible for score tracking (exempt from feature flag)
		wc.GET("/matches", wcHandler.ListMatches)
		wc.GET("/matches/:id", wcHandler.GetMatch)

		// Auth — public (no feature flag, no JWT required)
		auth := wc.Group("/auth")
		{
			auth.POST("/register", wcAuthHandler.Register)
			auth.POST("/login", wcAuthHandler.Login)
			auth.POST("/reset-password", wcAuthHandler.ResetPassword)
		}

		// Admin — always accessible regardless of feature flag
		wcAdminAlways := wc.Group("/admin", middleware.WcJWTMiddleware(wcAuthService), middleware.WcAdminMiddleware())
		{
			wcAdminAlways.POST("/sync", wcHandler.SyncMatches)
		}

		// All remaining WC routes require the feature to be enabled
		wcFeature := wc.Group("", middleware.WcFeatureMiddleware(wcRepo))
		{
			// Public (feature enabled, no auth)
			wcFeature.GET("/matches/:id/score-multipliers", wcHandler.GetScoreMultipliers)
			wcFeature.GET("/matches/:id/predictions", wcHandler.GetMatchPredictions)
			wcFeature.GET("/matches/:id/bets", wcHandler.GetMatchBets)
			wcFeature.GET("/leaderboard", wcHandler.GetLeaderboard)

			// JWT required
			wcAuth := wcFeature.Group("", middleware.WcJWTMiddleware(wcAuthService))
			{
				wcAuth.GET("/wallet", wcHandler.GetWallet)
				wcAuth.POST("/predictions", wcHandler.SubmitPrediction)
				wcAuth.GET("/predictions", wcHandler.ListPredictions)
				wcAuth.DELETE("/predictions/:id", wcHandler.DeletePrediction)
				wcAuth.PUT("/predictions/:id", wcHandler.UpdatePrediction)
				wcAuth.GET("/bets", wcHandler.ListBets)
				wcAuth.POST("/bets", wcHandler.PlaceBet)
				wcAuth.PUT("/bets/:id", wcHandler.UpdateBet)
				wcAuth.DELETE("/bets/:id", wcHandler.DeleteBet)
			}

			// Admin required
			wcAdmin := wcFeature.Group("/admin", middleware.WcJWTMiddleware(wcAuthService), middleware.WcAdminMiddleware())
			{
				wcAdmin.PUT("/matches/:id", wcHandler.UpdateMatch)
				wcAdmin.POST("/matches/:id/open", wcHandler.OpenMatch)
				wcAdmin.POST("/matches/:id/close", wcHandler.CloseMatch)
				wcAdmin.POST("/matches/:id/score-multipliers", wcHandler.AddScoreMultiplier)
				wcAdmin.PUT("/score-multipliers/:id", wcHandler.UpdateScoreMultiplier)
				wcAdmin.DELETE("/score-multipliers/:id", wcHandler.DeleteScoreMultiplier)
				wcAdmin.POST("/matches/:id/finalize", wcHandler.FinalizeMatch)
				wcAdmin.POST("/matches/:id/settle", wcHandler.SettleMatch)
				wcAdmin.POST("/matches/:id/score-odds", wcHandler.AddScoreOdds)
				wcAdmin.PUT("/score-odds/:id", wcHandler.UpdateScoreOdds)
				wcAdmin.DELETE("/score-odds/:id", wcHandler.DeleteScoreOdds)
				wcAdmin.GET("/users", wcHandler.ListUsers)
				wcAdmin.PUT("/users/:wc_user_id/role", wcHandler.SetUserRole)
				wcAdmin.GET("/wallets", wcHandler.ListAllWallets)
				wcAdmin.PUT("/wallets/:wc_user_id", wcHandler.AdminTopUp)
				wcAdmin.GET("/wallets/:wc_user_id/logs", wcHandler.GetWalletLogs)
				wcAdmin.GET("/settlements/preview", wcHandler.PreviewSettlement)
				wcAdmin.POST("/settlements", wcHandler.CreateSettlement)
				wcAdmin.GET("/settlements", wcHandler.ListSettlements)
				wcAdmin.GET("/settlements/:id", wcHandler.GetSettlement)
				wcAdmin.PUT("/settlements/:id/details/:wc_user_id", wcHandler.MarkSettlementDone)
			}
		}
	}

	return router
}
