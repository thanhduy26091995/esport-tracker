package cron

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/duyb/esport-score-tracker/internal/service"
)

// StartWcMatchSync runs SyncMatches immediately then on a fixed interval.
// Interval is read from WC_SYNC_INTERVAL_MINUTES (default: 30).
func StartWcMatchSync(svc *service.WcService) {
	intervalMinutes := 30
	if v := os.Getenv("WC_SYNC_INTERVAL_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			intervalMinutes = n
		}
	}

	sync := func() {
		count, err := svc.SyncMatches()
		if err != nil {
			log.Printf("⚠️  WC sync failed: %v", err)
			return
		}
		log.Printf("✅ WC sync: %d matches upserted", count)
	}

	// Run once immediately at startup
	sync()

	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	go func() {
		for range ticker.C {
			sync()
		}
	}()

	log.Printf("🔄 WC match sync scheduled every %d minutes", intervalMinutes)
}
