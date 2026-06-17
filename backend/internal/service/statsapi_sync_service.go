package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
)

const (
	oddsApiBaseURL = "https://api.odds-api.io/v3"
	oddsApiLeague  = "international-fifa-world-cup"
	oddsApiSport   = "football"
)

// StatsApiFixture is a WC2026 event from odds-api.io.
type StatsApiFixture struct {
	ID        string    `json:"id"`
	HomeTeam  string    `json:"home_team"`
	AwayTeam  string    `json:"away_team"`
	MatchDate time.Time `json:"match_date"`
}

// StatsApiOdds holds Asian handicap + O/U data.
type StatsApiOdds struct {
	StatsapiFixtureID string
	HandicapLine      float64
	HandicapHome      float64
	HandicapAway      float64
	OULine            float64
	OddsOver          float64
	OddsUnder         float64
}

// MappedMatch is one confirmed wc_match ↔ odds-api event_id pairing.
type MappedMatch struct {
	WcMatchID         uuid.UUID `json:"wc_match_id"`
	HomeTeam          string    `json:"home_team"`
	AwayTeam          string    `json:"away_team"`
	StatsapiFixtureID string    `json:"statsapi_fixture_id"`
	Confidence        string    `json:"confidence"`
}

// MappingResult is returned by the setup-mapping endpoint.
type MappingResult struct {
	Matched          []MappedMatch    `json:"matched"`
	UnmatchedLocal   []map[string]any `json:"unmatched_local"`
	UnmatchedAPI     []StatsApiFixture `json:"unmatched_api"`
	TotalAPIFixtures int              `json:"total_api_fixtures"`
}

// ImportHandicapPreview is the response for import-handicap preview.
type ImportHandicapPreview struct {
	MatchID           string `json:"match_id"`
	StatsapiFixtureID string `json:"statsapi_fixture_id"`
	Current           any    `json:"current"`
	Proposed          any    `json:"proposed"`
	Source            string `json:"source"`
	FetchedAt         string `json:"fetched_at"`
}

// ImportOUPreview is the response for import-ou preview.
type ImportOUPreview struct {
	MatchID   string `json:"match_id"`
	Current   any    `json:"current"`
	Proposed  any    `json:"proposed"`
	Source    string `json:"source"`
	FetchedAt string `json:"fetched_at"`
}

// oddsApiMarketLine is one handicap/O/U line from odds-api.io.
// Spread lines use Home/Away; Totals lines use Over/Under.
// All prices are strings in the API response.
type oddsApiMarketLine struct {
	Hdp   float64 `json:"hdp"`
	Home  string  `json:"home"`
	Away  string  `json:"away"`
	Over  string  `json:"over"`
	Under string  `json:"under"`
}

// oddsApiMarket is one market (Spread or Totals) from odds-api.io.
type oddsApiMarket struct {
	Name string              `json:"name"`
	Odds []oddsApiMarketLine `json:"odds"`
}

// preferredBookmakers: Sbobet first (sharpest Asian handicap), Bet365 as fallback.
// Free plan on odds-api.io supports exactly these two.
var preferredBookmakers = []string{"Sbobet", "Bet365"}

type StatsApiSyncService struct {
	repo   *repository.WcRepository
	client *http.Client
	apiKey string
}

func NewStatsApiSyncService(repo *repository.WcRepository, apiKey, _ string) *StatsApiSyncService {
	return &StatsApiSyncService{
		repo:   repo,
		client: &http.Client{Timeout: 15 * time.Second},
		apiKey: apiKey,
	}
}

// SetSportKey is kept for router.go compatibility; odds-api.io uses a fixed league slug.
func (s *StatsApiSyncService) SetSportKey(_ string) {}

func (s *StatsApiSyncService) get(path string) ([]byte, error) {
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	url := oddsApiBaseURL + path + sep + "apiKey=" + s.apiKey
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(string(body))
		if len(msg) > 300 {
			msg = msg[:300]
		}
		return nil, fmt.Errorf("odds-api.io %d: %s", resp.StatusCode, msg)
	}
	return body, nil
}

// FetchWC2026Fixtures returns all WC2026 events from odds-api.io.
// Endpoint: GET /v3/events?sport=football&league=international-fifa-world-cup
// Response: array of { id: int, home: string, away: string, date: string }
func (s *StatsApiSyncService) FetchWC2026Fixtures() ([]StatsApiFixture, error) {
	path := fmt.Sprintf("/events?sport=%s&league=%s", oddsApiSport, oddsApiLeague)
	body, err := s.get(path)
	if err != nil {
		return nil, err
	}
	var raw []struct {
		ID   int    `json:"id"`
		Home string `json:"home"`
		Away string `json:"away"`
		Date string `json:"date"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse events: %w", err)
	}
	fixtures := make([]StatsApiFixture, 0, len(raw))
	for _, d := range raw {
		t, _ := time.Parse(time.RFC3339, d.Date)
		fixtures = append(fixtures, StatsApiFixture{
			ID:        strconv.Itoa(d.ID),
			HomeTeam:  d.Home,
			AwayTeam:  d.Away,
			MatchDate: t,
		})
	}
	return fixtures, nil
}

// FetchOddsForMatch fetches handicap and O/U odds for a specific event.
// Endpoint: GET /v3/odds?eventId={id}&bookmakers=Sbobet,Bet365
// Response: { bookmakers: { "Sbobet": [{ name: "Spread"|"Totals", odds: [...] }] } }
func (s *StatsApiSyncService) FetchOddsForMatch(eventID string) (*StatsApiOdds, error) {
	path := fmt.Sprintf("/odds?eventId=%s&bookmakers=Sbobet,Bet365", eventID)
	body, err := s.get(path)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Bookmakers map[string][]oddsApiMarket `json:"bookmakers"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse event odds: %w", err)
	}

	odds := &StatsApiOdds{StatsapiFixtureID: eventID}

	for _, bkName := range preferredBookmakers {
		markets, ok := raw.Bookmakers[bkName]
		if !ok {
			continue
		}
		for _, market := range markets {
			switch market.Name {
			case "Spread":
				if odds.HandicapHome == 0 && len(market.Odds) > 0 {
					line := market.Odds[0]
					odds.HandicapLine = line.Hdp
					odds.HandicapHome = parseOddsStr(line.Home)
					odds.HandicapAway = parseOddsStr(line.Away)
				}
			case "Totals":
				if odds.OddsOver == 0 && len(market.Odds) > 0 {
					line := market.Odds[0]
					odds.OULine = line.Hdp
					odds.OddsOver = parseOddsStr(line.Over)
					odds.OddsUnder = parseOddsStr(line.Under)
				}
			}
		}
		if odds.HandicapHome > 0 && odds.OddsOver > 0 {
			break
		}
	}

	return odds, nil
}

// PreviewHandicap fetches handicap odds without writing to DB.
func (s *StatsApiSyncService) PreviewHandicap(matchID uuid.UUID) (*ImportHandicapPreview, error) {
	m, err := s.repo.GetMatch(matchID)
	if err != nil {
		return nil, fmt.Errorf("match not found")
	}
	if m.StatsapiFixtureID == nil {
		return nil, fmt.Errorf("match has no event ID — run Setup Mapping first")
	}
	odds, err := s.FetchOddsForMatch(*m.StatsapiFixtureID)
	if err != nil {
		return nil, err
	}
	return &ImportHandicapPreview{
		MatchID:           matchID.String(),
		StatsapiFixtureID: *m.StatsapiFixtureID,
		Current: map[string]any{
			"handicap_team":      m.HandicapTeam,
			"handicap_value":     m.HandicapValue,
			"odds_handicap_home": m.OddsHandicapHome,
			"odds_handicap_away": m.OddsHandicapAway,
		},
		Proposed: map[string]any{
			"handicap_team":      teamForHandicap(odds.HandicapLine),
			"handicap_value":     absF(odds.HandicapLine),
			"odds_handicap_home": odds.HandicapHome,
			"odds_handicap_away": odds.HandicapAway,
		},
		Source:    "odds-api.io",
		FetchedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// ImportHandicapForMatch writes handicap odds to DB (always overwrite).
func (s *StatsApiSyncService) ImportHandicapForMatch(matchID uuid.UUID, adminID uuid.UUID) error {
	m, err := s.repo.GetMatch(matchID)
	if err != nil {
		return fmt.Errorf("match not found")
	}
	if m.StatsapiFixtureID == nil {
		return fmt.Errorf("match has no event ID")
	}
	odds, err := s.FetchOddsForMatch(*m.StatsapiFixtureID)
	if err != nil {
		return err
	}
	if odds.HandicapHome == 0 {
		return fmt.Errorf("no handicap odds available for this event")
	}
	now := time.Now()
	if err := s.repo.UpdateMatch(matchID, map[string]any{
		"handicap_team":      teamForHandicap(odds.HandicapLine),
		"handicap_value":     absF(odds.HandicapLine),
		"odds_handicap_home": odds.HandicapHome,
		"odds_handicap_away": odds.HandicapAway,
		"odds_synced_at":     now,
	}); err != nil {
		return err
	}
	s.logSync(adminID, "manual", "handicap", 1, 0, nil)
	return nil
}

// PreviewOU fetches O/U odds without writing to DB.
func (s *StatsApiSyncService) PreviewOU(matchID uuid.UUID) (*ImportOUPreview, error) {
	m, err := s.repo.GetMatch(matchID)
	if err != nil {
		return nil, fmt.Errorf("match not found")
	}
	if m.StatsapiFixtureID == nil {
		return nil, fmt.Errorf("match has no event ID")
	}
	odds, err := s.FetchOddsForMatch(*m.StatsapiFixtureID)
	if err != nil {
		return nil, err
	}
	return &ImportOUPreview{
		MatchID: matchID.String(),
		Current: map[string]any{
			"ou_line":    m.OULine,
			"odds_over":  m.OddsOver,
			"odds_under": m.OddsUnder,
		},
		Proposed: map[string]any{
			"ou_line":    odds.OULine,
			"odds_over":  odds.OddsOver,
			"odds_under": odds.OddsUnder,
		},
		Source:    "odds-api.io",
		FetchedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// ImportOUForMatch writes O/U odds to DB (always overwrite).
func (s *StatsApiSyncService) ImportOUForMatch(matchID uuid.UUID, adminID uuid.UUID) error {
	m, err := s.repo.GetMatch(matchID)
	if err != nil {
		return fmt.Errorf("match not found")
	}
	if m.StatsapiFixtureID == nil {
		return fmt.Errorf("match has no event ID")
	}
	odds, err := s.FetchOddsForMatch(*m.StatsapiFixtureID)
	if err != nil {
		return err
	}
	if odds.OddsOver == 0 {
		return fmt.Errorf("no O/U odds available for this event")
	}
	now := time.Now()
	if err := s.repo.UpdateMatch(matchID, map[string]any{
		"ou_line":      odds.OULine,
		"odds_over":    odds.OddsOver,
		"odds_under":   odds.OddsUnder,
		"ou_synced_at": now,
	}); err != nil {
		return err
	}
	s.logSync(adminID, "manual", "ou", 1, 0, nil)
	return nil
}

// SetupMapping auto-matches local wc_matches with WC2026 events from odds-api.io.
// Pass 1: normalized name + same UTC day → confidence "exact"
// Pass 2: normalized name only (±1 day tolerance) → confidence "name_match"
func (s *StatsApiSyncService) SetupMapping(previewOnly bool, adminID uuid.UUID) (*MappingResult, error) {
	fixtures, err := s.FetchWC2026Fixtures()
	if err != nil {
		return nil, err
	}
	matches, err := s.repo.ListAllMatches()
	if err != nil {
		return nil, err
	}

	result := &MappingResult{TotalAPIFixtures: len(fixtures)}
	apiUsed := make(map[string]bool)
	localUsed := make(map[uuid.UUID]bool)

	tryMatch := func(dateCheck func(a, b time.Time) bool, confidence string) {
		for _, f := range fixtures {
			if apiUsed[f.ID] {
				continue
			}
			fHome := normalizeTeam(f.HomeTeam)
			fAway := normalizeTeam(f.AwayTeam)
			for _, m := range matches {
				if localUsed[m.ID] {
					continue
				}
				if normalizeTeam(m.HomeTeam) == fHome &&
					normalizeTeam(m.AwayTeam) == fAway &&
					dateCheck(m.MatchDate, f.MatchDate) {
					result.Matched = append(result.Matched, MappedMatch{
						WcMatchID:         m.ID,
						HomeTeam:          m.HomeTeam,
						AwayTeam:          m.AwayTeam,
						StatsapiFixtureID: f.ID,
						Confidence:        confidence,
					})
					apiUsed[f.ID] = true
					localUsed[m.ID] = true
					break
				}
			}
		}
	}

	// Pass 1: same UTC day (strict)
	tryMatch(sameDay, "exact")
	// Pass 2: ±1 day tolerance (handles timezone edge cases, pre-scheduled dates)
	tryMatch(nearDay, "near_date")

	for _, m := range matches {
		if !localUsed[m.ID] {
			result.UnmatchedLocal = append(result.UnmatchedLocal, map[string]any{
				"id": m.ID, "home_team": m.HomeTeam, "away_team": m.AwayTeam,
			})
		}
	}
	for _, f := range fixtures {
		if !apiUsed[f.ID] {
			result.UnmatchedAPI = append(result.UnmatchedAPI, f)
		}
	}

	if !previewOnly {
		for _, mm := range result.Matched {
			id := mm.StatsapiFixtureID
			if err := s.repo.UpdateMatch(mm.WcMatchID, map[string]any{"statsapi_fixture_id": id}); err != nil {
				log.Printf("failed to save mapping for %v: %v", mm.WcMatchID, err)
			}
		}
		s.logSync(adminID, "manual", "mapping", len(result.Matched), 0, nil)
	}
	return result, nil
}

// SyncUpcomingMatchesBlank fills blank handicap/O/U odds for upcoming matches (cron mode).
func (s *StatsApiSyncService) SyncUpcomingMatchesBlank() (updated, failed int, err error) {
	matches, err := s.repo.ListUpcomingMatchesWithStatsapiID()
	if err != nil {
		return 0, 0, err
	}
	for _, m := range matches {
		if m.StatsapiFixtureID == nil {
			continue
		}
		odds, fetchErr := s.FetchOddsForMatch(*m.StatsapiFixtureID)
		if fetchErr != nil {
			log.Printf("cron: failed to fetch odds for %v: %v", m.ID, fetchErr)
			failed++
			time.Sleep(2 * time.Second)
			continue
		}
		upd := map[string]any{}
		now := time.Now()
		if m.OddsHandicapHome == nil && odds.HandicapHome > 0 {
			upd["handicap_team"] = teamForHandicap(odds.HandicapLine)
			upd["handicap_value"] = absF(odds.HandicapLine)
			upd["odds_handicap_home"] = odds.HandicapHome
			upd["odds_handicap_away"] = odds.HandicapAway
			upd["odds_synced_at"] = now
		}
		if m.OddsOver == nil && odds.OddsOver > 0 {
			upd["ou_line"] = odds.OULine
			upd["odds_over"] = odds.OddsOver
			upd["odds_under"] = odds.OddsUnder
			upd["ou_synced_at"] = now
		}
		if len(upd) > 0 {
			if updErr := s.repo.UpdateMatch(m.ID, upd); updErr != nil {
				log.Printf("cron: failed to update match %v: %v", m.ID, updErr)
				failed++
			} else {
				updated++
			}
		}
		time.Sleep(2 * time.Second) // rate limit: ~30 req/min on free plan
	}
	s.logSync(uuid.Nil, "cron", "handicap+ou", updated, failed, nil)
	return
}

// StartCron runs SyncUpcomingMatchesBlank on the given interval.
func (s *StatsApiSyncService) StartCron(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	log.Printf("OddsAPI cron started (interval=%v, league=%s)", interval, oddsApiLeague)
	for range ticker.C {
		updated, failed, err := s.SyncUpcomingMatchesBlank()
		if err != nil {
			log.Printf("OddsAPI cron error: %v", err)
		} else {
			log.Printf("OddsAPI cron: updated=%d failed=%d", updated, failed)
		}
	}
}

// GetSyncLogs returns the last 20 sync log entries.
func (s *StatsApiSyncService) GetSyncLogs() ([]*model.WcSyncLog, error) {
	return s.repo.GetSyncLogs()
}

// UpsertScoreMultipliers bulk-upserts Poisson-generated odds into wc_score_multipliers.
func (s *StatsApiSyncService) UpsertScoreMultipliers(multipliers []model.WcScoreMultiplier) error {
	return s.repo.BulkUpsertScoreMultipliers(multipliers)
}

// TouchPoissonSync stamps poisson_synced_at on the match so the admin panel can show it.
func (s *StatsApiSyncService) TouchPoissonSync(matchID uuid.UUID) error {
	return s.repo.UpdateMatch(matchID, map[string]any{"poisson_synced_at": time.Now()})
}

func (s *StatsApiSyncService) logSync(adminID uuid.UUID, trigger, syncType string, updated, failed int, errDetail *string) {
	var triggeredBy *uuid.UUID
	if adminID != uuid.Nil {
		triggeredBy = &adminID
	}
	_ = s.repo.CreateSyncLog(&model.WcSyncLog{
		Trigger:        trigger,
		SyncType:       syncType,
		TriggeredBy:    triggeredBy,
		MatchesUpdated: updated,
		MatchesFailed:  failed,
		ErrorDetail:    errDetail,
	})
}

func parseOddsStr(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

func teamForHandicap(line float64) string {
	if line < 0 {
		return model.WcTeamHome
	}
	return model.WcTeamAway
}

func absF(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// diacriticReplacer strips common accented characters to their ASCII base.
var diacriticReplacer = strings.NewReplacer(
	"ô", "o", "Ô", "o",
	"é", "e", "É", "e",
	"è", "e", "È", "e",
	"ê", "e", "Ê", "e",
	"ë", "e", "Ë", "e",
	"à", "a", "À", "a",
	"â", "a", "Â", "a",
	"ä", "a", "Ä", "a",
	"ü", "u", "Ü", "u",
	"ö", "o", "Ö", "o",
	"ñ", "n", "Ñ", "n",
	"ç", "c", "Ç", "c",
	"'", "",             // right apostrophe (Côte d'Ivoire)
	"'", "",             // left apostrophe
	"-", " ",            // Bosnia-Herzegovina → Bosnia Herzegovina
)

// teamCanonical maps any known name variation to a shared canonical key.
// Both football-data.org and odds-api.io spellings are listed so both sides
// map to the same string after diacriticReplacer is applied.
var teamCanonical = map[string]string{
	// Korea
	"south korea":              "korea",
	"korea republic":           "korea",
	"republic of korea":        "korea",
	// Ivory Coast
	"ivory coast":              "cotedivoire",
	"cote divoire":             "cotedivoire",
	"cote d ivoire":            "cotedivoire",
	// Cape Verde
	"cape verde islands":       "capeverde",
	"cape verde":               "capeverde",
	"cabo verde":               "capeverde",
	// Congo
	"congo dr":                 "drcongo",
	"dr congo":                 "drcongo",
	"democratic republic of the congo": "drcongo",
	// Bosnia (after dash→space: "bosnia herzegovina")
	"bosnia herzegovina":       "bosnia",
	"bosnia and herzegovina":   "bosnia",
	// Turkey / Türkiye (ü stripped to u → "turkiye" vs "turkey")
	"turkey":                   "turkiye",
	// USA
	"united states":            "usa",
	"united states of america": "usa",
	// Curaçao (ç stripped to c → "curacao" naturally)
	// Palestine
	"state of palestine":       "palestine",
	// Czech Republic
	"czech republic":           "czechia",
	// North Macedonia
	"north macedonia":          "northmacedonia",
	"republic of north macedonia": "northmacedonia",
	// New Zealand
	"new zealand":              "newzealand",
	"nz":                       "newzealand",
}

// normalizeTeam returns a canonical lowercase key for matching.
// Steps: lowercase → strip diacritics → trim → alias lookup.
func normalizeTeam(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = diacriticReplacer.Replace(s)
	s = strings.TrimSpace(s)
	if canonical, ok := teamCanonical[s]; ok {
		return canonical
	}
	return s
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.UTC().Date()
	by, bm, bd := b.UTC().Date()
	return ay == by && am == bm && ad == bd
}

func nearDay(a, b time.Time) bool {
	diff := a.UTC().Truncate(24 * time.Hour).Sub(b.UTC().Truncate(24 * time.Hour))
	if diff < 0 {
		diff = -diff
	}
	return diff <= 24*time.Hour
}
