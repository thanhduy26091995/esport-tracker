---
phase: implementation
title: WC Tournament Analytics — Implementation Guide
description: Technical implementation notes, patterns, and code guidelines
---

# Implementation Guide

## Development Setup

Add go-cache dependency (not yet in `go.mod`):
```bash
cd backend && go get github.com/patrickmn/go-cache
```

## Code Structure

```
backend/internal/service/
  wc_football_data_client.go        ← NEW: football-data.org /scorers
  wc_open_football_client.go        ← NEW: openfootball worldcup.json (bulk GET)

backend/internal/repository/
  wc_repository.go                  ← EXTEND: GetCompletedMatchStats()

backend/internal/service/
  wc_analytics_service.go           ← EXTEND: GetWorldCup2026Analytics() + cache/fd/of fields

backend/internal/api/
  wc_analytics_handler.go           ← EXTEND: GetWorldCup2026Analytics handler
  router.go                         ← EXTEND: GET /analytics/world-cup-2026 in wcAuth group

frontend/src/components/wc/
  WcAnalyticsPanel.vue              ← EXTEND: add "World Cup 2026" as first/default tab
frontend/src/components/wc/analytics/
  WcTournamentPanel.vue             ← NEW: 11-section content component
frontend/src/stores/
  wcAnalyticsStore.ts               ← EXTEND: wc2026Data + loadWC2026Analytics()
frontend/src/services/
  wcAnalyticsService.ts             ← EXTEND: getWC2026Analytics()
frontend/src/types/wc.ts            ← EXTEND: all new interfaces
```

No router changes needed — `WcAnalyticsPanel` already lives inside `WcPredictView.vue`.

---

## WcFootballDataClient (`wc_football_data_client.go`)

Separate from existing `wc_football_client.go` (which syncs match schedules). This only handles `/scorers`.

```go
package service

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "github.com/your-org/esport-tracker/internal/model"
)

const footballDataBaseURL = "https://api.football-data.org/v4"

type FootballDataClient struct {
    apiKey     string
    httpClient *http.Client
}

func NewFootballDataClient(apiKey string) *FootballDataClient {
    return &FootballDataClient{
        apiKey:     apiKey,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

func (c *FootballDataClient) GetWCScorers(limit int) ([]model.WcScorer, time.Time, error) {
    url := fmt.Sprintf("%s/competitions/WC/scorers?limit=%d", footballDataBaseURL, limit)
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("X-Auth-Token", c.apiKey)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, time.Time{}, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return nil, time.Time{}, fmt.Errorf("football-data.org returned %d", resp.StatusCode)
    }

    var result fdScorersResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, time.Time{}, err
    }

    scorers := make([]model.WcScorer, 0, len(result.Scorers))
    for i, s := range result.Scorers {
        scorers = append(scorers, model.WcScorer{
            Rank: i + 1, PlayerName: s.Player.Name,
            TeamName: s.Team.Name, TeamCode: s.Team.TLA, TeamCrest: s.Team.Crest,
            Goals: s.Goals, Assists: s.Assists, PlayedMatches: s.PlayedMatches,
        })
    }
    return scorers, time.Now(), nil
}

type fdScorersResponse struct{ Scorers []fdScorer `json:"scorers"` }
type fdScorer struct {
    Player        fdPlayer `json:"player"`
    Team          fdTeam   `json:"team"`
    Goals         int      `json:"goals"`
    Assists       *int     `json:"assists"`
    PlayedMatches int      `json:"playedMatches"`
}
type fdPlayer struct{ Name string `json:"name"` }
type fdTeam struct {
    Name  string `json:"name"`
    TLA   string `json:"tla"`
    Crest string `json:"crest"`
}
```

---

## OpenFootballClient (`open_football_client.go`)

Single bulk GET. Parses all goal events, aggregates into analytics-ready structs.

```go
package service

import (
    "encoding/json"
    "net/http"
    "sort"
    "strconv"
    "strings"
    "time"
    "github.com/your-org/esport-tracker/internal/model"
)

const openfootballURL = "https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.json"

type OpenFootballClient struct {
    httpClient *http.Client
}

func NewOpenFootballClient() *OpenFootballClient {
    return &OpenFootballClient{httpClient: &http.Client{Timeout: 15 * time.Second}}
}

// WcOpenFootballData contains all analytics derived from worldcup.json.
type WcOpenFootballData struct {
    GoalTiming        []model.WcGoalTimingBucket
    HalfTimeStats     model.WcHalfTimeStats
    TeamStats         []model.WcTeamStat
    GoalsByGroup      []model.WcGroupGoals
    TopScoringMatches []model.WcMatchDetail
    VenueStats        []model.WcVenueStat
}

func (c *OpenFootballClient) GetWCData() (*WcOpenFootballData, error) {
    resp, err := c.httpClient.Get(openfootballURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var raw ofWorldcup
    if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
        return nil, err
    }

    // Accumulation maps
    buckets := map[string]int{
        "1-15": 0, "16-30": 0, "31-45": 0, "45+": 0,
        "46-60": 0, "61-75": 0, "76-90": 0, "90+": 0,
    }
    var ownGoals, penaltyGoals, firstHalf, secondHalf int
    var comebacks, heldLead int

    teamGoalsFor := map[string]int{}
    teamGoalsAgainst := map[string]int{}
    teamMatches := map[string]int{}

    groupGoals := map[string]int{}
    groupMatches := map[string]int{}

    venueGoals := map[string]int{}
    venueMatches := map[string]int{}

    type matchScore struct {
        match model.WcMatchDetail
        total int
    }
    var allMatches []matchScore

    for _, m := range raw.Matches {
        if m.Score == nil { continue } // upcoming match — no score yet

        ft := m.Score.FT
        ht := m.Score.HT
        homeGoals := ft[0]
        awayGoals := ft[1]
        totalGoals := homeGoals + awayGoals

        // Team stats
        teamGoalsFor[m.Team1] += homeGoals
        teamGoalsAgainst[m.Team1] += awayGoals
        teamMatches[m.Team1]++
        teamGoalsFor[m.Team2] += awayGoals
        teamGoalsAgainst[m.Team2] += homeGoals
        teamMatches[m.Team2]++

        // Group stats
        groupGoals[m.Group] += totalGoals
        groupMatches[m.Group]++

        // Venue stats
        venueGoals[m.Ground] += totalGoals
        venueMatches[m.Ground]++

        // HT comeback detection (only if HT score present)
        if len(ht) == 2 {
            htHome, htAway := ht[0], ht[1]
            trailedHT1 := htHome < htAway && ft[0] > ft[1]  // team1 trailed HT, won FT
            trailedHT2 := htAway < htHome && ft[1] > ft[0]  // team2 trailed HT, won FT
            if trailedHT1 || trailedHT2 { comebacks++ }
            ledHT1 := htHome > htAway && ft[0] > ft[1]
            ledHT2 := htAway > htHome && ft[1] > ft[0]
            if ledHT1 || ledHT2 { heldLead++ }
        }

        // Goal events
        for _, g := range m.Goals1 {
            name := fixMojibake(g.Name)
            _ = name
            if g.OwnGoal { ownGoals++; continue }
            if g.Penalty { penaltyGoals++ }
            bucket := parseBucket(g.Minute)
            buckets[bucket]++
            if isFirstHalf(g.Minute) { firstHalf++ } else { secondHalf++ }
        }
        for _, g := range m.Goals2 {
            name := fixMojibake(g.Name)
            _ = name
            if g.OwnGoal { ownGoals++; continue }
            if g.Penalty { penaltyGoals++ }
            bucket := parseBucket(g.Minute)
            buckets[bucket]++
            if isFirstHalf(g.Minute) { firstHalf++ } else { secondHalf++ }
        }

        allMatches = append(allMatches, matchScore{
            match: model.WcMatchDetail{
                HomeTeam: m.Team1, AwayTeam: m.Team2,
                HomeScore: homeGoals, AwayScore: awayGoals,
                TotalGoals: totalGoals,
                Group: m.Group, Round: m.Round, Date: m.Date, Venue: m.Ground,
            },
            total: totalGoals,
        })
    }

    // Assemble bucket slice (ordered)
    bucketOrder := []string{"1-15", "16-30", "31-45", "45+", "46-60", "61-75", "76-90", "90+"}
    timingSlice := make([]model.WcGoalTimingBucket, len(bucketOrder))
    for i, label := range bucketOrder {
        timingSlice[i] = model.WcGoalTimingBucket{Label: label, Goals: buckets[label]}
    }

    // Team stats — sort by goals for desc
    teamSlice := make([]model.WcTeamStat, 0, len(teamMatches))
    for name, matches := range teamMatches {
        teamSlice = append(teamSlice, model.WcTeamStat{
            TeamName: name, GoalsFor: teamGoalsFor[name],
            GoalsAgainst: teamGoalsAgainst[name], Matches: matches,
        })
    }
    sort.Slice(teamSlice, func(i, j int) bool { return teamSlice[i].GoalsFor > teamSlice[j].GoalsFor })

    // Group goals — sort by group name
    groupSlice := make([]model.WcGroupGoals, 0, len(groupGoals))
    for g, goals := range groupGoals {
        groupSlice = append(groupSlice, model.WcGroupGoals{Group: g, Matches: groupMatches[g], Goals: goals})
    }
    sort.Slice(groupSlice, func(i, j int) bool { return groupSlice[i].Group < groupSlice[j].Group })

    // Venue stats
    venueSlice := make([]model.WcVenueStat, 0, len(venueGoals))
    for v, goals := range venueGoals {
        venueSlice = append(venueSlice, model.WcVenueStat{
            Venue: v, Matches: venueMatches[v], Goals: goals,
        })
    }
    sort.Slice(venueSlice, func(i, j int) bool { return venueSlice[i].Goals > venueSlice[j].Goals })

    // Top 5 scoring matches
    sort.Slice(allMatches, func(i, j int) bool { return allMatches[i].total > allMatches[j].total })
    top5 := make([]model.WcMatchDetail, 0, 5)
    for i, m := range allMatches {
        if i >= 5 { break }
        top5 = append(top5, m.match)
    }

    return &WcOpenFootballData{
        GoalTiming: timingSlice,
        HalfTimeStats: model.WcHalfTimeStats{
            FirstHalfGoals: firstHalf, SecondHalfGoals: secondHalf,
            OwnGoals: ownGoals, PenaltyGoals: penaltyGoals,
            Comebacks: comebacks, HeldLead: heldLead,
        },
        TeamStats:         teamSlice,
        GoalsByGroup:      groupSlice,
        TopScoringMatches: top5,
        VenueStats:        venueSlice,
    }, nil
}

// parseBucket assigns a goal minute string to a display bucket.
func parseBucket(minuteStr string) string {
    parts := strings.SplitN(minuteStr, "+", 2)
    base, _ := strconv.Atoi(parts[0])
    injury := len(parts) == 2
    switch {
    case injury && base == 45: return "45+"
    case injury && base >= 90: return "90+"
    case base <= 15:  return "1-15"
    case base <= 30:  return "16-30"
    case base <= 45:  return "31-45"
    case base <= 60:  return "46-60"
    case base <= 75:  return "61-75"
    default:          return "76-90"
    }
}

func isFirstHalf(minuteStr string) bool {
    parts := strings.SplitN(minuteStr, "+", 2)
    base, _ := strconv.Atoi(parts[0])
    injury := len(parts) == 2
    return base <= 45 && !injury || (base == 45 && injury)
}

// fixMojibake corrects accented names stored as Latin-1 codepoints in the JSON.
// "VinÃ­cius" → "Vinícius"
func fixMojibake(s string) string {
    runes := []rune(s)
    b := make([]byte, len(runes))
    for i, r := range runes {
        if r > 0xFF { return s }
        b[i] = byte(r)
    }
    return string(b)
}

// --- raw JSON structs ---

type ofWorldcup struct {
    Matches []ofMatch `json:"matches"`
}
type ofMatch struct {
    Round  string    `json:"round"`
    Date   string    `json:"date"`
    Team1  string    `json:"team1"`
    Team2  string    `json:"team2"`
    Score  *ofScore  `json:"score"`
    Goals1 []ofGoal  `json:"goals1"`
    Goals2 []ofGoal  `json:"goals2"`
    Group  string    `json:"group"`
    Ground string    `json:"ground"`
}
type ofScore struct {
    FT []int `json:"ft"`
    HT []int `json:"ht"`
}
type ofGoal struct {
    Name    string `json:"name"`
    Minute  string `json:"minute"`
    OwnGoal bool   `json:"owngoal"`
    Penalty bool   `json:"penalty"`
}
```

---

## WcService — Three-tier caching

Add fields to `WcService` struct and update constructor:

```go
type WcService struct {
    repo          WcRepository
    userRepo      UserRepository
    customBetRepo CustomBetRepository
    cache         *cache.Cache
    fdClient      *FootballDataClient
    ofClient      *OpenFootballClient
    hub           ws.HubBroadcaster
}

const (
    cacheKeyResponse     = "wc:analytics:response"
    cacheKeyScorers      = "wc:analytics:scorers"
    cacheKeyOpenFootball = "wc:analytics:openfootball"
    cacheTTLResponse     = 5 * time.Minute
    cacheTTLScorers      = 30 * time.Minute
    cacheTTLOpenFootball = 15 * time.Minute
)

type scorersCacheEntry struct {
    scorers   []model.WcScorer
    fetchedAt time.Time
}

func (s *WcService) GetTournamentAnalytics() (*model.WcAnalyticsResponse, error) {
    // Tier 1: full response cache
    if cached, found := s.cache.Get(cacheKeyResponse); found {
        r := cached.(model.WcAnalyticsResponse)
        return &r, nil
    }

    matchStats, err := s.repo.GetCompletedMatchStats()
    if err != nil { return nil, err }
    resp := &model.WcAnalyticsResponse{MatchStats: *matchStats}

    // Tier 2: football-data.org scorers
    if cached, found := s.cache.Get(cacheKeyScorers); found {
        entry := cached.(scorersCacheEntry)
        resp.TopScorers = entry.scorers
        t := entry.fetchedAt
        resp.ScorersUpdatedAt = &t
    } else {
        scorers, fetchedAt, err := s.fdClient.GetWCScorers(20)
        if err == nil {
            s.cache.Set(cacheKeyScorers, scorersCacheEntry{scorers, fetchedAt}, cacheTTLScorers)
            resp.TopScorers = scorers
            resp.ScorersUpdatedAt = &fetchedAt
        }
        // err → graceful degradation, top_scorers stays empty
    }

    // Tier 3: openfootball event data
    if cached, found := s.cache.Get(cacheKeyOpenFootball); found {
        ofData := cached.(*WcOpenFootballData)
        applyOpenFootballData(resp, ofData)
    } else {
        ofData, err := s.ofClient.GetWCData()
        if err == nil {
            s.cache.Set(cacheKeyOpenFootball, ofData, cacheTTLOpenFootball)
            applyOpenFootballData(resp, ofData)
        }
        // err → graceful degradation, openfootball fields stay nil (omitempty)
    }

    s.cache.Set(cacheKeyResponse, *resp, cacheTTLResponse)
    return resp, nil
}

func applyOpenFootballData(resp *model.WcAnalyticsResponse, d *WcOpenFootballData) {
    resp.GoalTiming = d.GoalTiming
    resp.HalfTimeStats = &d.HalfTimeStats
    resp.TeamStats = d.TeamStats
    resp.GoalsByGroup = d.GoalsByGroup
    resp.TopScoringMatches = d.TopScoringMatches
    resp.VenueStats = d.VenueStats
}
```

---

## WcRepository — GetCompletedMatchStats

```go
func (r *WcRepository) GetCompletedMatchStats() (*model.WcMatchStats, error) {
    type agg struct {
        TotalMatches int; TotalGoals int
        HomeWins int; AwayWins int; Draws int; CleanSheets int
    }
    var a agg
    err := r.db.Raw(`
        SELECT
            COUNT(*)                                                              AS total_matches,
            COALESCE(SUM(home_score + away_score), 0)                            AS total_goals,
            SUM(CASE WHEN home_score > away_score THEN 1 ELSE 0 END)             AS home_wins,
            SUM(CASE WHEN away_score > home_score THEN 1 ELSE 0 END)             AS away_wins,
            SUM(CASE WHEN home_score = away_score THEN 1 ELSE 0 END)             AS draws,
            SUM(CASE WHEN home_score = 0 OR away_score = 0 THEN 1 ELSE 0 END)   AS clean_sheets
        FROM wc_matches WHERE status = 'completed' AND home_score IS NOT NULL
    `).Scan(&a).Error
    if err != nil { return nil, err }

    type stageRow struct{ Stage string; Matches int; Goals int }
    var stageRows []stageRow
    r.db.Raw(`
        SELECT stage, COUNT(*) AS matches, SUM(home_score + away_score) AS goals
        FROM wc_matches WHERE status = 'completed' AND home_score IS NOT NULL
        GROUP BY stage ORDER BY MIN(match_date)
    `).Scan(&stageRows)
    stageGoals := make([]model.WcStageGoals, len(stageRows))
    for i, s := range stageRows {
        stageGoals[i] = model.WcStageGoals{Stage: s.Stage, Matches: s.Matches, Goals: s.Goals}
    }

    var top model.WcMatch
    r.db.Where("status = 'completed' AND home_score IS NOT NULL").
        Order("(home_score + away_score) DESC").First(&top)
    var highest *model.WcMatchResult
    if top.ID != uuid.Nil {
        total := *top.HomeScore + *top.AwayScore
        highest = &model.WcMatchResult{
            HomeTeam: top.HomeTeam, AwayTeam: top.AwayTeam,
            HomeScore: *top.HomeScore, AwayScore: *top.AwayScore,
            Stage: top.Stage, TotalGoals: total,
        }
    }
    avg := 0.0
    if a.TotalMatches > 0 {
        avg = math.Round(float64(a.TotalGoals)/float64(a.TotalMatches)*100) / 100
    }
    return &model.WcMatchStats{
        TotalMatches: a.TotalMatches, TotalGoals: a.TotalGoals, AvgGoalsPerMatch: avg,
        HomeWins: a.HomeWins, AwayWins: a.AwayWins, Draws: a.Draws, CleanSheets: a.CleanSheets,
        HighestScoringMatch: highest, GoalsByStage: stageGoals,
    }, nil
}
```

---

## Handler & Route

```go
// wc_handler.go
func (h *WcHandler) GetTournamentAnalytics(c *gin.Context) {
    result, err := h.wcService.GetTournamentAnalytics()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load analytics"})
        return
    }
    c.JSON(http.StatusOK, result)
}

// router.go
wc.GET("/analytics", wcAuthMiddleware, h.GetTournamentAnalytics)
```

---

## WcAnalyticsPanel.vue — Add first tab

```vue
<!-- Add BEFORE the existing "Của Tôi" tab pane -->
<el-tab-pane :label="t('wc.analytics.wc2026Tab')" name="wc2026">
  <WcTournamentPanel
    :data="store.wc2026Data"
    :loading="store.wc2026Loading"
  />
</el-tab-pane>
```

```ts
// Change default
const analyticsTab = ref('wc2026')

// Add import
import WcTournamentPanel from './analytics/WcTournamentPanel.vue'

// Update onMounted — load wc2026 instead of my (new default tab)
onMounted(async () => {
  if (!store.wc2026Data) await store.loadWC2026Analytics()
})

// Update watch — add wc2026 case
watch(analyticsTab, async (tab) => {
  if (tab === 'wc2026' && !store.wc2026Data) await store.loadWC2026Analytics()
  if (tab === 'my' && !store.myData) await store.loadMyAnalytics()
  if (tab === 'community' && !store.communityData) await store.loadCommunityAnalytics()
  if (tab === 'compare' && !store.compareData) await store.loadCompareAnalytics()
})
```

## wcAnalyticsStore.ts — Add wc2026 state

```ts
const wc2026Data = ref<WcAnalyticsResponse | null>(null)
const wc2026Loading = ref(false)

async function loadWC2026Analytics() {
  wc2026Loading.value = true
  try {
    wc2026Data.value = await wcAnalyticsService.getWC2026Analytics()
  } catch {
    // fail silently — component shows empty state
  } finally {
    wc2026Loading.value = false
  }
}

// Expose in return:
return { ..., wc2026Data, wc2026Loading, loadWC2026Analytics }
```

## wcAnalyticsService.ts — Add API call

```ts
getWC2026Analytics(): Promise<WcAnalyticsResponse> {
  return wcApi.get<WcAnalyticsResponse>('/analytics/world-cup-2026').then(r => r.data)
}
```

## WcTournamentPanel.vue — Props interface

```vue
<script setup lang="ts">
import type { WcAnalyticsResponse } from '@/types/wc'

defineProps<{
  data: WcAnalyticsResponse | null
  loading: boolean
}>()
</script>
```

## Frontend Types (`types/wc.ts` additions)

```ts
export interface WcAnalyticsResponse {
  match_stats: WcMatchStats
  top_scorers: WcScorer[]
  scorers_updated_at?: string
  goal_timing?: WcGoalTimingBucket[]
  half_time_stats?: WcHalfTimeStats
  team_stats?: WcTeamStat[]
  goals_by_group?: WcGroupGoals[]
  top_scoring_matches?: WcMatchDetail[]
  venue_stats?: WcVenueStat[]
}

export interface WcMatchStats {
  total_matches: number
  total_goals: number
  avg_goals_per_match: number
  home_wins: number
  away_wins: number
  draws: number
  clean_sheets: number
  highest_scoring_match?: WcMatchResult
  goals_by_stage: WcStageGoals[]
}

export interface WcScorer {
  rank: number; player_name: string; team_name: string
  team_code: string; team_crest: string
  goals: number; assists: number | null; played_matches: number
}

export interface WcGoalTimingBucket { label: string; goals: number }

export interface WcHalfTimeStats {
  first_half_goals: number; second_half_goals: number
  own_goals: number; penalty_goals: number
  comebacks: number; held_lead: number
}

export interface WcTeamStat {
  team_name: string; goals_for: number; goals_against: number; matches: number
}

export interface WcGroupGoals { group: string; matches: number; goals: number }

export interface WcMatchDetail {
  home_team: string; away_team: string
  home_score: number; away_score: number; total_goals: number
  group?: string; round: string; date: string; venue: string
}

export interface WcVenueStat { venue: string; matches: number; goals: number }

export interface WcMatchResult {
  home_team: string; away_team: string
  home_score: number; away_score: number; stage: string; total_goals: number
}

export interface WcStageGoals { stage: string; matches: number; goals: number }
```

---

## Frontend — wcApi.ts

```ts
getAnalytics(): Promise<WcAnalyticsResponse> {
  return this.client.get('/wc/analytics').then(r => r.data)
}
```

---

## Frontend — WcTournamentPanel.vue sections

| # | Section | Data field | Chart type |
|---|---------|-----------|------------|
| 1 | Tổng quan (6 stat cards) | `match_stats` + `half_time_stats` | Cards |
| 2 | **Vua phá lưới** | `top_scorers` | Table with crest |
| 3 | Kết quả trận đấu | `match_stats.home_wins/away_wins/draws` | Horizontal bar |
| 4 | Phân tích hiệp đấu | `half_time_stats` | Stats + text |
| 5 | Thời điểm ghi bàn | `goal_timing` | Vertical bar chart |
| 6 | Bàn theo vòng đấu | `match_stats.goals_by_stage` | Bar chart |
| 7 | Bàn theo bảng đấu | `goals_by_group` | Bar chart |
| 8 | Top 5 trận nhiều bàn | `top_scoring_matches` | Match cards list |
| 9 | Đội ghi nhiều bàn | `team_stats` | Sortable table |
| 10 | Sân vận động | `venue_stats` | Table |
| 11 | Phản lưới & Penalty | `half_time_stats.own_goals/penalty_goals` | 2 stat cards |

---

## DI / main.go

```go
wcCache := cache.New(30*time.Minute, 10*time.Minute)
fdClient := service.NewFootballDataClient(os.Getenv("FOOTBALL_DATA_API_KEY"))
ofClient := service.NewOpenFootballClient()
wcService := service.NewWcService(wcRepo, userRepo, customBetRepo, wcCache, fdClient, ofClient, hub)
```

---

## Error Handling

- DB error → HTTP 500 (match stats are core)
- football-data.org error → `top_scorers: []`, `scorers_updated_at: null`
- openfootball error → all `omitempty` openfootball fields absent from response
- 0 completed matches → all zeros, no `highest_scoring_match`

## Performance

- DB: ~10ms (64-100 rows, no index needed)
- Cache hit: < 50ms for any tier
- Cold start (both external APIs): football-data.org ~500ms + openfootball ~300ms, parallel if goroutines used
- Response cache (5 min): absorbs burst traffic
