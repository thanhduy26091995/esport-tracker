package cron

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/duyb/esport-score-tracker/internal/service"
)

// StartWcMatchSync runs SyncMatches immediately then on an adaptive schedule.
// It checks every minute whether enough time has elapsed based on current match state:
//   - Live matches present → sync every 5 minutes
//   - No live matches      → sync every intervalMinutes (default: 30)
//
// This means the interval adapts within 1 minute of matches going live,
// rather than waiting until the current sleep window expires.
func StartWcMatchSync(svc *service.WcService) {
	intervalMinutes := 30
	if v := os.Getenv("WC_SYNC_INTERVAL_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			intervalMinutes = n
		}
	}

	idleInterval := time.Duration(intervalMinutes) * time.Minute
	liveInterval := 5 * time.Minute

	sync := func() {
		count, err := svc.SyncMatches()
		if err != nil {
			log.Printf("⚠️  WC sync failed: %v", err)
			return
		}
		log.Printf("✅ WC sync: %d matches upserted", count)
	}

	desiredInterval := func() time.Duration {
		summary, err := svc.GetMatchScheduleSummary()
		if err != nil || summary.LiveCount > 0 {
			return liveInterval
		}
		return idleInterval
	}

	// Run once immediately at startup
	sync()

	go func() {
		lastSync := time.Now()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if time.Since(lastSync) >= desiredInterval() {
				sync()
				lastSync = time.Now()
			}
		}
	}()

	log.Printf("🔄 WC match sync started (idle: %d min, live: 5 min)", intervalMinutes)
}
