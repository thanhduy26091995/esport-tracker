package api

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/duyb/esport-score-tracker/internal/cache"
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
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Internal-Key"},
		AllowCredentials: true,
	}
	router.Use(cors.New(corsConfig))

	// Serve uploaded avatar files (after CORS so browser can fetch cross-origin)
	router.Static("/uploads", "./uploads")

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "FC25 Esport Score Tracker API"})
	})

	// Initialize cache — soft startup: Redis failure falls back to go-cache (no crash)
	var cacheStore cache.CacheStore
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		rc, err := cache.NewRedisCache(redisURL)
		if err != nil {
			log.Printf("⚠️  Redis unavailable (%v) — falling back to go-cache", err)
			cacheStore = cache.NewGoCacheStore(10*time.Minute, 5*time.Minute)
		} else {
			cacheStore = rc
			log.Println("Cache backend: Redis")
		}
	} else {
		cacheStore = cache.NewGoCacheStore(10*time.Minute, 5*time.Minute)
		log.Println("Cache backend: go-cache (dev mode)")
	}

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
	wcAnalyticsRepo := repository.NewWcAnalyticsRepository(db)

	// Initialize services
	configService := service.NewConfigService(configRepo, cacheStore)
	userService := service.NewUserService(userRepo, matchRepo, configService, cacheStore)
	fundService := service.NewFundService(fundRepo, cacheStore)
	settlementService := service.NewSettlementService(settlementRepo, userRepo, matchRepo, fundService, configService, db, cacheStore)
	tierService := service.NewTierService(userRepo, configService)
	matchService := service.NewMatchService(matchRepo, userRepo, settlementService, configService, tierService, db, cacheStore)
	tournamentService := service.NewTournamentService(tournamentRepo, userRepo, matchService, db)
	wcAuthService := service.NewWcAuthService(wcUserRepo, wcRepo)
	wcProfileService := service.NewWcProfileService(wcUserRepo)

	wsHub := ws.NewHub()
	go wsHub.Run()

	chatHub := ws.NewHub()
	go chatHub.Run()

	wcChatService := service.NewWcChatService(wcChatRepo, chatHub)

	wcService := service.NewWcService(wcRepo, wcUserRepo, wcCustomBetRepo, wsHub, cacheStore)
	wcChampionService := service.NewWcChampionService(wcChampionRepo, wcRepo, wcUserRepo, wsHub)
	wcCustomBetService := service.NewWcCustomBetService(wcCustomBetRepo, wcRepo, wcUserRepo, wsHub)
	statsApiKey := os.Getenv("ODDSAPI_KEY")
	statsApiSyncService := service.NewStatsApiSyncService(wcRepo, statsApiKey, "")
	poissonService := service.NewPoissonService()
	bonusRepo := repository.NewScoreBonusRepository(db)
	bonusService := service.NewScoreBonusService(bonusRepo, userRepo, tierService, db, cacheStore)

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
	wcFdClient := service.NewWcFootballDataClient(os.Getenv("FOOTBALL_DATA_API_KEY"))
	wcOfClient := service.NewWcOpenFootballClient()
	wcAnalyticsService := service.NewWcAnalyticsService(wcAnalyticsRepo, wcRepo, cacheStore, wcFdClient, wcOfClient)
	wcAnalyticsHandler := NewWcAnalyticsHandler(wcAnalyticsService)
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

	// API v1 group — non-WC routes protected by X-Internal-Key
	v1 := router.Group("/api/v1", middleware.InternalKeyMiddleware())
	{
		// User routes
		users := v1.Group("/users")
		{
			users.GET("", userHandler.GetAll)                            // GET /api/v1/users
			users.POST("", userHandler.Create)                           // POST /api/v1/users
			users.GET("/leaderboard", userHandler.GetLeaderboard)        // GET /api/v1/users/leaderboard
			users.GET("/payment-ranking", userHandler.GetPaymentRanking) // GET /api/v1/users/payment-ranking
			users.GET("/head-to-head", userHandler.GetHeadToHead)        // GET /api/v1/users/head-to-head?player1=&player2=
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

	// setupTournamentRoutes registers all WC/AC routes for a given tournament prefix and type.
	// Both /wc and /ac share the same handlers — tournament context is set by TournamentMiddleware.
	setupTournamentRoutes := func(prefix, tournamentTypeVal string) {
		t := router.Group("/api/v1/"+prefix, middleware.TournamentMiddleware(tournamentTypeVal))
		{
			// Config endpoints — always accessible (exempt from feature flag)
			t.GET("/config", wcHandler.GetPublicConfig)
			t.GET("/admin/config", middleware.WcJWTMiddleware(wcAuthService), middleware.WcAdminMiddleware(), wcHandler.GetConfig)
			t.PUT("/admin/config", middleware.WcJWTMiddleware(wcAuthService), middleware.WcAdminMiddleware(), wcHandler.UpdateConfig)

			// Matches & standings — always accessible (exempt from feature flag)
			t.GET("/matches", wcHandler.ListMatches)
			t.GET("/matches/:id", wcHandler.GetMatch)
			t.GET("/standings", wcHandler.GetGroupStandings)

			// Live chat history — public, no auth required
			t.GET("/chat/messages", wcChatHandler.ListMessages)

			// Auth — public (no feature flag, no JWT required)
			auth := t.Group("/auth")
			{
				auth.POST("/login", wcAuthHandler.Login)
				auth.POST("/google", wcAuthHandler.GoogleLoginOrCreate)
				auth.POST("/google/link", middleware.WcJWTMiddleware(wcAuthService), wcAuthHandler.GoogleLink)
			}

			// Admin — always accessible regardless of feature flag
			tAdminAlways := t.Group("/admin", middleware.WcJWTMiddleware(wcAuthService), middleware.WcAdminMiddleware())
			{
				if tournamentTypeVal == "world_cup" {
					tAdminAlways.POST("/sync", wcHandler.SyncMatches)
				}
				// Manual match creation — for tournaments without API sync
				tAdminAlways.POST("/matches", wcHandler.CreateMatch)
			}

			// All remaining routes require the feature to be enabled
			tFeature := t.Group("", middleware.WcFeatureMiddleware(wcRepo))
			{
				// Public (feature enabled, no auth)
				tFeature.GET("/matches/:id/score-multipliers", wcHandler.GetScoreMultipliers)
				tFeature.GET("/matches/:id/predictions", wcHandler.GetMatchPredictions)
				tFeature.GET("/matches/:id/bets", wcHandler.GetMatchBets)
				tFeature.GET("/leaderboard", wcHandler.GetLeaderboard)
				tFeature.GET("/champion/config", wcChampionHandler.GetConfig)
				tFeature.GET("/champion/teams", wcChampionHandler.GetTeams)
				tFeature.GET("/champion/predictions", wcChampionHandler.GetAllPredictions)

				// JWT + Google-linked required (player profile)
				tPlayerAuth := tFeature.Group("", middleware.WcJWTMiddleware(wcAuthService), middleware.WcGoogleLinkedMiddleware(db))
				{
					tPlayerAuth.GET("/profile", wcProfileHandler.GetProfile)
					tPlayerAuth.PUT("/profile", wcProfileHandler.UpdateProfile)
				}

				// JWT + Google-linked required (player actions)
				tAuth := tFeature.Group("", middleware.WcJWTMiddleware(wcAuthService), middleware.WcGoogleLinkedMiddleware(db))
				{
					tAuth.GET("/wallet", wcHandler.GetWallet)
					tAuth.POST("/predictions", wcHandler.SubmitPrediction)
					tAuth.GET("/predictions", wcHandler.ListPredictions)
					tAuth.DELETE("/predictions/:id", wcHandler.DeletePrediction)
					tAuth.PUT("/predictions/:id", wcHandler.UpdatePrediction)
					tAuth.GET("/predictions/:id/reduce-preview", wcHandler.PreviewReducePredictionPoints)
					tAuth.GET("/bets", wcHandler.ListBets)
					tAuth.GET("/bets/history", wcHandler.GetBetHistory)
					tAuth.POST("/bets", wcHandler.PlaceBet)
					tAuth.PUT("/bets/:id", wcHandler.UpdateBet)
					tAuth.DELETE("/bets/:id", wcHandler.DeleteBet)
					tAuth.GET("/bets/:id/reduce-preview", wcHandler.PreviewReduceStake)
					tAuth.GET("/champion/my-prediction", wcChampionHandler.GetMyPrediction)
					tAuth.POST("/champion/predict", wcChampionHandler.PlacePredict)
					tAuth.DELETE("/champion/predict/:id", wcChampionHandler.DeletePredict)
					// Custom bets (player)
					tAuth.GET("/matches/:id/custom-bets", wcCustomBetHandler.ListCustomBets)
					tAuth.GET("/custom-bet-entries", wcCustomBetHandler.GetMyCustomBetEntries)
					tAuth.POST("/custom-bets/:id/entry", wcCustomBetHandler.PlaceEntry)
					tAuth.DELETE("/custom-bet-entries/:id", wcCustomBetHandler.CancelEntry)
					// Analytics
					tAuth.GET("/analytics/my", wcAnalyticsHandler.GetMyAnalytics)
					tAuth.GET("/analytics/community", wcAnalyticsHandler.GetCommunityAnalytics)
					tAuth.GET("/analytics/compare", wcAnalyticsHandler.GetCompareAnalytics)
					tAuth.GET("/analytics/world-cup-2026", wcAnalyticsHandler.GetWorldCup2026Analytics)
					// Chat mention
					tAuth.GET("/users", wcHandler.ListUsersForMention)
					tAuth.GET("/chat/mentions/unread-count", wcChatHandler.GetUnreadMentionCount)
					tAuth.POST("/chat/mentions/read", wcChatHandler.MarkMentionsRead)
				}

				// Admin required
				tAdmin := tFeature.Group("/admin", middleware.WcJWTMiddleware(wcAuthService), middleware.WcAdminMiddleware())
				{
					tAdmin.PUT("/matches/:id", wcHandler.UpdateMatch)
					tAdmin.POST("/matches/:id/open", wcHandler.OpenMatch)
					tAdmin.POST("/matches/:id/close", wcHandler.CloseMatch)
					tAdmin.POST("/matches/:id/score-multipliers", wcHandler.AddScoreMultiplier)
					tAdmin.PUT("/score-multipliers/:id", wcHandler.UpdateScoreMultiplier)
					tAdmin.DELETE("/score-multipliers/:id", wcHandler.DeleteScoreMultiplier)
					tAdmin.GET("/matches/finalize-all-preview", wcHandler.PreviewFinalizeAll)
					tAdmin.GET("/matches/refinalize-all-preview", wcHandler.PreviewRefinalizeAll)
					tAdmin.GET("/matches/:id/finalize-preview", wcHandler.PreviewFinalizeMatch)
					tAdmin.POST("/matches/finalize-all", wcHandler.FinalizeAll)
					tAdmin.POST("/matches/refinalize-all", wcHandler.RefinalizeAll)
					tAdmin.POST("/matches/:id/finalize", wcHandler.FinalizeMatch)
					tAdmin.POST("/matches/:id/settle", wcHandler.SettleMatch)
					tAdmin.POST("/matches/:id/score-odds", wcHandler.AddScoreOdds)
					tAdmin.PUT("/score-odds/:id", wcHandler.UpdateScoreOdds)
					tAdmin.DELETE("/score-odds/:id", wcHandler.DeleteScoreOdds)
					tAdmin.GET("/users", wcHandler.ListUsers)
					tAdmin.PUT("/users/:wc_user_id/bot", wcHandler.SetUserBot)
					tAdmin.PUT("/users/:wc_user_id/role", wcHandler.SetUserRole)
					tAdmin.PUT("/users/:wc_user_id/block", wcHandler.BlockUser)
					tAdmin.PUT("/users/:wc_user_id/unblock", wcHandler.UnblockUser)
					tAdmin.GET("/wallets", wcHandler.ListAllWallets)
					tAdmin.PUT("/wallets/:wc_user_id", wcHandler.AdminTopUp)
					tAdmin.GET("/wallets/:wc_user_id/logs", wcHandler.GetWalletLogs)
					tAdmin.GET("/house-pnl", wcHandler.GetHousePnL)
					if tournamentTypeVal == "world_cup" {
						tAdmin.POST("/setup-statsapi-mapping", wcSyncHandler.SetupMapping)
						tAdmin.POST("/matches/:id/import-handicap", wcSyncHandler.ImportHandicap)
						tAdmin.POST("/matches/:id/import-ou", wcSyncHandler.ImportOU)
						tAdmin.POST("/matches/:id/generate-poisson", wcSyncHandler.GeneratePoisson)
						tAdmin.GET("/sync-logs", wcSyncHandler.GetSyncLogs)
					}
					tAdmin.GET("/settlements/preview", wcHandler.PreviewSettlement)
					tAdmin.POST("/settlements", wcHandler.CreateSettlement)
					tAdmin.GET("/settlements", wcHandler.ListSettlements)
					tAdmin.GET("/settlements/:id", wcHandler.GetSettlement)
					tAdmin.PUT("/settlements/:id/details/:wc_user_id", wcHandler.MarkSettlementDone)
					tAdmin.POST("/backfill-original-points", wcHandler.BackfillOriginalPoints)
					// Champion prediction admin
					tAdmin.PUT("/champion/config", wcChampionHandler.AdminUpdateConfig)
					tAdmin.POST("/champion/teams", wcChampionHandler.AdminCreateTeam)
					tAdmin.PUT("/champion/teams/:id", wcChampionHandler.AdminUpdateTeamOdds)
					tAdmin.POST("/champion/settle", wcChampionHandler.AdminSettle)
					// Custom bets admin
					tAdmin.GET("/matches/:id/custom-bets", wcCustomBetHandler.AdminListCustomBets)
					tAdmin.POST("/matches/:id/custom-bets", wcCustomBetHandler.AdminCreateCustomBet)
					tAdmin.PUT("/custom-bets/:id", wcCustomBetHandler.AdminUpdateCustomBet)
					tAdmin.POST("/custom-bets/:id/settle", wcCustomBetHandler.AdminSettleCustomBet)
					tAdmin.PUT("/custom-bets/:id/void", wcCustomBetHandler.AdminVoidCustomBet)
				}
			}
		}
	}

	setupTournamentRoutes("wc", "world_cup")
	setupTournamentRoutes("ac", "asean_cup")

	return router
}
