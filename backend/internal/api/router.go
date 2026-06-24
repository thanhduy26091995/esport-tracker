package api

import (
	"log"
	"os"
	"strings"

	"github.com/duyb/esport-score-tracker/internal/cron"
	"github.com/duyb/esport-score-tracker/internal/middleware"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/duyb/esport-score-tracker/internal/ws"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.MaxMultipartMemory = 5 << 20 // 5 MB limit for avatar uploads

	// CORS middleware — must be registered before any routes (including Static)
	corsConfig := cors.Config{
		AllowOrigins:     strings.Split(os.Getenv("CORS_ORIGINS"), ","),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}
	router.Use(cors.New(corsConfig))

	// Serve uploaded avatar files (after CORS so browser can fetch cross-origin)
	router.Static("/uploads", "./uploads")

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
	wcChampionRepo := repository.NewWcChampionRepository(db)
	wcCustomBetRepo := repository.NewWcCustomBetRepository(db)
	wcChatRepo := repository.NewWcChatRepository(db)

	// Initialize services
	configService := service.NewConfigService(configRepo)
	userService := service.NewUserService(userRepo, configService)
	fundService := service.NewFundService(fundRepo)
	settlementService := service.NewSettlementService(settlementRepo, userRepo, matchRepo, fundService, configService, db)
	tierService := service.NewTierService(userRepo, configService)
	matchService := service.NewMatchService(matchRepo, userRepo, settlementService, configService, tierService, db)
	tournamentService := service.NewTournamentService(tournamentRepo, userRepo, matchService, db)
	wcAuthService := service.NewWcAuthService(wcUserRepo, wcRepo)
	wcProfileService := service.NewWcProfileService(wcUserRepo)

	wsHub := ws.NewHub()
	go wsHub.Run()

	chatHub := ws.NewHub()
	go chatHub.Run()

	wcChatService := service.NewWcChatService(wcChatRepo, chatHub)

	wcService := service.NewWcService(wcRepo, wcUserRepo, wcCustomBetRepo, wsHub)
	wcChampionService := service.NewWcChampionService(wcChampionRepo, wcRepo, wcUserRepo, wsHub)
	wcCustomBetService := service.NewWcCustomBetService(wcCustomBetRepo, wcRepo, wcUserRepo, wsHub)
	statsApiKey := os.Getenv("ODDSAPI_KEY")
	statsApiSyncService := service.NewStatsApiSyncService(wcRepo, statsApiKey, "")
	poissonService := service.NewPoissonService()
	bonusRepo := repository.NewScoreBonusRepository(db)
	bonusService := service.NewScoreBonusService(bonusRepo, userRepo, tierService, db)

	// Backfill tiers from existing match history on startup.
	if err := tierService.RecalculateAllTiers(); err != nil {
		log.Printf("⚠️  Failed to backfill tiers on startup: %v", err)
	}

	// Start WC match auto-sync cron job.
	go cron.StartWcMatchSync(wcService)
	if statsApiKey != "" {
		go statsApiSyncService.StartCron(30 * 60 * 1e9) // 30min
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
	wcProfileHandler := NewWcProfileHandler(wcProfileService)
	wcHandler := NewWcHandler(wcService, wcAuthService)
	wcSyncHandler := NewWcSyncHandler(statsApiSyncService, poissonService)
	wcChampionHandler := NewWcChampionHandler(wcChampionService)
	wcCustomBetHandler := NewWcCustomBetHandler(wcCustomBetService)
	wcChatHandler := NewWcChatHandler(wcChatService)
	wsHandler := ws.NewHandler(wsHub)

	// Token verifier adapter for ChatHandler
	tokenVerifier := ws.NewWcAuthTokenVerifier(func(tokenStr string) (uuid.UUID, string, error) {
		claims, err := wcAuthService.VerifyToken(tokenStr)
		if err != nil {
			return uuid.UUID{}, "", err
		}
		return claims.WcUserID, claims.Name, nil
	})
	// Avatar URL fetcher adapter for ChatHandler
	avatarFetcher := ws.NewWcUserAvatarFetcher(func(userID uuid.UUID) string {
		user, err := wcUserRepo.GetByID(userID)
		if err != nil || user.AvatarURL == nil {
			return ""
		}
		return *user.AvatarURL
	})
	chatWsHandler := ws.NewChatHandler(chatHub, tokenVerifier, avatarFetcher, wcChatService)

	// WebSocket endpoints — outside /api/v1, proxied by Nginx with Upgrade headers
	router.GET("/ws", wsHandler.Handle)
	router.GET("/ws/chat", chatWsHandler.Handle)

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
			users.PUT("/:id", userHandler.Update)                    // PUT /api/v1/users/:id
			users.DELETE("/:id", userHandler.Delete)                 // DELETE /api/v1/users/:id
			users.PUT("/:id/avatar", userHandler.UploadAvatar)         // PUT /api/v1/users/:id/avatar
			users.PUT("/:id/avatar/url", userHandler.SetAvatarURL)   // PUT /api/v1/users/:id/avatar/url
			users.DELETE("/:id/avatar", userHandler.DeleteAvatar)    // DELETE /api/v1/users/:id/avatar
			users.PUT("/:id/club", userHandler.UpdateClub)           // PUT /api/v1/users/:id/club
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
			tournaments.POST("/:id/generate-knockouts", tournamentHandler.GenerateKnockouts)
		}
	}

	// WC2026 routes — /api/v1/wc
	wc := router.Group("/api/v1/wc")
	{
		// Config endpoints — always accessible (exempt from feature flag)
		wc.GET("/config", wcHandler.GetPublicConfig)
		wc.GET("/admin/config", middleware.WcJWTMiddleware(wcAuthService), middleware.WcAdminMiddleware(), wcHandler.GetConfig)
		wc.PUT("/admin/config", middleware.WcJWTMiddleware(wcAuthService), middleware.WcAdminMiddleware(), wcHandler.UpdateConfig)

		// Matches & standings — always accessible (exempt from feature flag)
		wc.GET("/matches", wcHandler.ListMatches)
		wc.GET("/matches/:id", wcHandler.GetMatch)
		wc.GET("/standings", wcHandler.GetGroupStandings)

		// Live chat history — public, no auth required
		wc.GET("/chat/messages", wcChatHandler.ListMessages)

		// Auth — public (no feature flag, no JWT required)
		auth := wc.Group("/auth")
		{
			auth.POST("/login", wcAuthHandler.Login)
			auth.POST("/google", wcAuthHandler.GoogleLoginOrCreate)
			// Google link: requires JWT only (this IS the endpoint that satisfies the link requirement)
			auth.POST("/google/link", middleware.WcJWTMiddleware(wcAuthService), wcAuthHandler.GoogleLink)
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
			wcFeature.GET("/champion/config", wcChampionHandler.GetConfig)
			wcFeature.GET("/champion/teams", wcChampionHandler.GetTeams)
			wcFeature.GET("/champion/predictions", wcChampionHandler.GetAllPredictions)

			// JWT + Google-linked required (player profile)
			wcPlayerAuth := wcFeature.Group("", middleware.WcJWTMiddleware(wcAuthService), middleware.WcGoogleLinkedMiddleware(db))
			{
				wcPlayerAuth.GET("/profile", wcProfileHandler.GetProfile)
				wcPlayerAuth.PUT("/profile", wcProfileHandler.UpdateProfile)
			}

			// JWT + Google-linked required
			wcAuth := wcFeature.Group("", middleware.WcJWTMiddleware(wcAuthService), middleware.WcGoogleLinkedMiddleware(db))
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
				wcAuth.GET("/champion/my-prediction", wcChampionHandler.GetMyPrediction)
				wcAuth.POST("/champion/predict", wcChampionHandler.PlacePredict)
				wcAuth.DELETE("/champion/predict/:id", wcChampionHandler.DeletePredict)
				// Custom bets (player)
				wcAuth.GET("/matches/:id/custom-bets", wcCustomBetHandler.ListCustomBets)
				wcAuth.GET("/custom-bet-entries", wcCustomBetHandler.GetMyCustomBetEntries)
				wcAuth.POST("/custom-bets/:id/entry", wcCustomBetHandler.PlaceEntry)
				wcAuth.DELETE("/custom-bet-entries/:id", wcCustomBetHandler.CancelEntry)
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
				wcAdmin.GET("/matches/finalize-all-preview", wcHandler.PreviewFinalizeAll)
				wcAdmin.GET("/matches/refinalize-all-preview", wcHandler.PreviewRefinalizeAll)
				wcAdmin.GET("/matches/:id/finalize-preview", wcHandler.PreviewFinalizeMatch)
				wcAdmin.POST("/matches/finalize-all", wcHandler.FinalizeAll)
				wcAdmin.POST("/matches/refinalize-all", wcHandler.RefinalizeAll)
				wcAdmin.POST("/matches/:id/finalize", wcHandler.FinalizeMatch)
				wcAdmin.POST("/matches/:id/settle", wcHandler.SettleMatch)
				wcAdmin.POST("/matches/:id/score-odds", wcHandler.AddScoreOdds)
				wcAdmin.PUT("/score-odds/:id", wcHandler.UpdateScoreOdds)
				wcAdmin.DELETE("/score-odds/:id", wcHandler.DeleteScoreOdds)
				wcAdmin.GET("/users", wcHandler.ListUsers)
				wcAdmin.PUT("/users/:wc_user_id/role", wcHandler.SetUserRole)
				wcAdmin.PUT("/users/:wc_user_id/block", wcHandler.BlockUser)
				wcAdmin.PUT("/users/:wc_user_id/unblock", wcHandler.UnblockUser)
				wcAdmin.GET("/wallets", wcHandler.ListAllWallets)
				wcAdmin.PUT("/wallets/:wc_user_id", wcHandler.AdminTopUp)
				wcAdmin.GET("/wallets/:wc_user_id/logs", wcHandler.GetWalletLogs)
				wcAdmin.GET("/house-pnl", wcHandler.GetHousePnL)
				wcAdmin.POST("/setup-statsapi-mapping", wcSyncHandler.SetupMapping)
				wcAdmin.POST("/matches/:id/import-handicap", wcSyncHandler.ImportHandicap)
				wcAdmin.POST("/matches/:id/import-ou", wcSyncHandler.ImportOU)
				wcAdmin.POST("/matches/:id/generate-poisson", wcSyncHandler.GeneratePoisson)
				wcAdmin.GET("/sync-logs", wcSyncHandler.GetSyncLogs)
				wcAdmin.GET("/settlements/preview", wcHandler.PreviewSettlement)
				wcAdmin.POST("/settlements", wcHandler.CreateSettlement)
				wcAdmin.GET("/settlements", wcHandler.ListSettlements)
				wcAdmin.GET("/settlements/:id", wcHandler.GetSettlement)
				wcAdmin.PUT("/settlements/:id/details/:wc_user_id", wcHandler.MarkSettlementDone)
				// Champion prediction admin
				wcAdmin.PUT("/champion/config", wcChampionHandler.AdminUpdateConfig)
				wcAdmin.POST("/champion/teams", wcChampionHandler.AdminCreateTeam)
				wcAdmin.PUT("/champion/teams/:id", wcChampionHandler.AdminUpdateTeamOdds)
				wcAdmin.POST("/champion/settle", wcChampionHandler.AdminSettle)
				// Custom bets admin
				wcAdmin.GET("/matches/:id/custom-bets", wcCustomBetHandler.AdminListCustomBets)
				wcAdmin.POST("/matches/:id/custom-bets", wcCustomBetHandler.AdminCreateCustomBet)
				wcAdmin.PUT("/custom-bets/:id", wcCustomBetHandler.AdminUpdateCustomBet)
				wcAdmin.POST("/custom-bets/:id/settle", wcCustomBetHandler.AdminSettleCustomBet)
				wcAdmin.PUT("/custom-bets/:id/void", wcCustomBetHandler.AdminVoidCustomBet)
			}
		}
	}

	return router
}
