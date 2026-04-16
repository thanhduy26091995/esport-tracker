package api

import (
	"os"
	"strings"

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

	// Initialize services
	userService := service.NewUserService(userRepo)
	configService := service.NewConfigService(configRepo)
	fundService := service.NewFundService(fundRepo)
	settlementService := service.NewSettlementService(settlementRepo, userRepo, matchRepo, fundService, configService, db)
	matchService := service.NewMatchService(matchRepo, userRepo, settlementService, configService, db)
	tournamentService := service.NewTournamentService(tournamentRepo, userRepo, matchService, db)

	// Initialize handlers
	userHandler := NewUserHandler(userService)
	matchHandler := NewMatchHandler(matchService)
	configHandler := NewConfigHandler(configService)
	fundHandler := NewFundHandler(fundService)
	settlementHandler := NewSettlementHandler(settlementService)
	tournamentHandler := NewTournamentHandler(tournamentService)

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.GET("", userHandler.GetAll)              // GET /api/v1/users
			users.POST("", userHandler.Create)             // POST /api/v1/users
			users.GET("/leaderboard", userHandler.GetLeaderboard) // GET /api/v1/users/leaderboard
			users.GET("/:id", userHandler.GetByID)         // GET /api/v1/users/:id
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
			settlements.GET("/stats", settlementHandler.GetStats) // GET /api/v1/settlements/stats
			settlements.GET("/:id", settlementHandler.GetByID)    // GET /api/v1/settlements/:id
		}

		// User settlement history
		v1.GET("/users/:id/settlements", settlementHandler.GetByDebtorID) // GET /api/v1/users/:id/settlements

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

	return router
}
