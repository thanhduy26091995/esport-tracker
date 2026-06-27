---
phase: design
title: WC Tournament Analytics — System Design
description: Architecture, data flow, API design, and component breakdown for WC2026 analytics page
---

# System Design & Architecture

## Architecture Overview

```mermaid
graph TD
    User["WC User (Browser)"] -->|GET /api/v1/wc/analytics/world-cup-2026| WcAnalyticsHandler["WcAnalyticsHandler\nGetWorldCup2026Analytics"]
    WcAnalyticsHandler --> WcAnalyticsSvc["WcAnalyticsService\nGetWorldCup2026Analytics()"]

    WcAnalyticsSvc -->|QueryMatchStats| WcRepo["WcRepository\nGetCompletedMatchStats()"]
    WcRepo --> DB[(PostgreSQL\nwc_matches)]

    WcAnalyticsSvc -->|cache hit?| RespCache["go-cache\nwc:analytics:wc2026\n5min TTL"]

    RespCache -->|miss| ScorersCache["go-cache\nwc:analytics:scorers\n30min TTL"]
    ScorersCache -->|miss| FDClient["WcFootballDataClient\nGET /competitions/WC/scorers"]
    FDClient -->|X-Auth-Token| FDAPI[("football-data.org\nfree tier")]
    FDAPI --> FDClient
    FDClient --> ScorersCache

    RespCache -->|miss| OFCache["go-cache\nwc:analytics:openfootball\n15min TTL"]
    OFCache -->|miss| OFClient["WcOpenFootballClient\nGET worldcup.json"]
    OFClient --> OFAPI[("github raw\nopenfootball/worldcup.json\nno auth")]
    OFAPI --> OFClient
    OFClient --> OFCache

    ScorersCache --> WcAnalyticsSvc
    OFCache --> WcAnalyticsSvc
    WcAnalyticsSvc --> RespCache
    RespCache --> WcAnalyticsHandler
    WcAnalyticsHandler --> User

    VuePage["WcAnalyticsPage.vue"] -->|wcApi.getWC2026Analytics()| WcAnalyticsHandler
```

**Data sources by field group:**

| Field group | Source | Cache TTL |
|-------------|--------|-----------|
| total_matches, total_goals, avg, home/away/draws, clean_sheets, goals_by_stage, highest_scoring_match | DB `wc_matches` | — (always fresh, ~10ms) |
| top_scorers (rank, goals, assists, team crest) | football-data.org `/scorers` | 30 min |
| goal_timing, own_goals, penalty_goals, 1st/2nd half, team_stats, goals_by_group, ht_stats, top_matches | openfootball `worldcup.json` | 15 min |
| Full assembled response | go-cache response | 5 min |

**openfootball JSON schema (confirmed):**
```json
{
  "matches": [
    {
      "round": "Matchday 1",
      "date":  "2026-06-11",
      "team1": "Mexico",
      "team2": "South Africa",
      "score": { "ft": [2, 0], "ht": [1, 0] },
      "goals1": [{ "name": "Julián Quiñones", "minute": "67" }],
      "goals2": [],
      "group":  "Group A",
      "ground": "Mexico City"
    }
  ]
}
```
- `minute` is a **string** — can be `"45+3"`, `"90+7"` (injury time)
- `owngoal: true` flag — `name` is the player who scored the own goal
- `penalty: true` flag
- No `status` field — completed match inferred by presence of `score` key
- `team1`/`team2` are plain strings, not objects

**New dependency required:** `github.com/patrickmn/go-cache` — add via `go get github.com/patrickmn/go-cache`.

**Why NOT odds-api.io:** Only keeps ~44 most recent events, not historical analytics.

---

## Data Models

### Backend Response DTO

```go
type WcAnalyticsResponse struct {
    // From DB — always returned
    MatchStats  WcMatchStats  `json:"match_stats"`
    // From football-data.org — graceful degradation if unavailable
    TopScorers       []WcScorer `json:"top_scorers"`
    ScorersUpdatedAt *time.Time `json:"scorers_updated_at,omitempty"`
    // From openfootball — graceful degradation if unavailable
    GoalTiming        []WcGoalTimingBucket `json:"goal_timing,omitempty"`
    HalfTimeStats     *WcHalfTimeStats     `json:"half_time_stats,omitempty"`
    TeamStats         []WcTeamStat         `json:"team_stats,omitempty"`
    GoalsByGroup      []WcGroupGoals       `json:"goals_by_group,omitempty"`
    TopScoringMatches []WcMatchDetail      `json:"top_scoring_matches,omitempty"` // top 5
    VenueStats        []WcVenueStat        `json:"venue_stats,omitempty"`
}

// --- DB-sourced ---

type WcMatchStats struct {
    TotalMatches        int               `json:"total_matches"`
    TotalGoals          int               `json:"total_goals"`
    AvgGoalsPerMatch    float64           `json:"avg_goals_per_match"`
    HomeWins            int               `json:"home_wins"`
    AwayWins            int               `json:"away_wins"`
    Draws               int               `json:"draws"`
    CleanSheets         int               `json:"clean_sheets"`
    HighestScoringMatch *WcMatchResult    `json:"highest_scoring_match,omitempty"`
    GoalsByStage        []WcStageGoals    `json:"goals_by_stage"`
}

type WcMatchResult struct {
    HomeTeam   string `json:"home_team"`
    AwayTeam   string `json:"away_team"`
    HomeScore  int    `json:"home_score"`
    AwayScore  int    `json:"away_score"`
    Stage      string `json:"stage"`
    TotalGoals int    `json:"total_goals"`
}

type WcStageGoals struct {
    Stage   string `json:"stage"`
    Matches int    `json:"matches"`
    Goals   int    `json:"goals"`
}

// --- football-data.org sourced ---

type WcScorer struct {
    Rank          int    `json:"rank"`
    PlayerName    string `json:"player_name"`
    TeamName      string `json:"team_name"`
    TeamCode      string `json:"team_code"`
    TeamCrest     string `json:"team_crest"`
    Goals         int    `json:"goals"`
    Assists       *int   `json:"assists"`    // nullable
    PlayedMatches int    `json:"played_matches"`
}

// --- openfootball sourced ---

type WcGoalTimingBucket struct {
    Label string `json:"label"` // "1-15" | "16-30" | "31-45" | "45+" | "46-60" | "61-75" | "76-90" | "90+"
    Goals int    `json:"goals"`
}

type WcHalfTimeStats struct {
    FirstHalfGoals  int `json:"first_half_goals"`   // goals with minute <= 45 (inc 45+X)
    SecondHalfGoals int `json:"second_half_goals"`  // goals with minute >= 46
    OwnGoals        int `json:"own_goals"`
    PenaltyGoals    int `json:"penalty_goals"`
    Comebacks       int `json:"comebacks"`           // trailed at HT, won at FT
    HeldLead        int `json:"held_lead"`           // led at HT, won at FT
}

type WcTeamStat struct {
    TeamName     string `json:"team_name"`
    GoalsFor     int    `json:"goals_for"`
    GoalsAgainst int    `json:"goals_against"`
    Matches      int    `json:"matches"`
}

type WcGroupGoals struct {
    Group   string `json:"group"`
    Matches int    `json:"matches"`
    Goals   int    `json:"goals"`
}

type WcMatchDetail struct {
    HomeTeam   string `json:"home_team"`
    AwayTeam   string `json:"away_team"`
    HomeScore  int    `json:"home_score"`
    AwayScore  int    `json:"away_score"`
    TotalGoals int    `json:"total_goals"`
    Group      string `json:"group,omitempty"`
    Round      string `json:"round"`
    Date       string `json:"date"`
    Venue      string `json:"venue"`
}

type WcVenueStat struct {
    Venue   string `json:"venue"`
    Matches int    `json:"matches"`
    Goals   int    `json:"goals"`
}
```

---

## API Design

### Endpoint

```
GET /api/v1/wc/analytics/world-cup-2026
```

**Auth:** `WcJWTMiddleware` + `WcGoogleLinkedMiddleware` (same as existing `/analytics/my` route)
**Query params:** none
**Response:** `WcAnalyticsResponse` JSON, HTTP 200

**Error cases:**
- DB unavailable → 500
- football-data.org unavailable → `top_scorers: []`, `scorers_updated_at: null`
- openfootball unavailable → all openfootball fields omitted (`omitempty`)
- Both external APIs down → DB stats still served

### football-data.org (internal)

```
GET https://api.football-data.org/v4/competitions/WC/scorers?limit=20
X-Auth-Token: <FOOTBALL_DATA_API_KEY>
```

**Cache:** `"wc:analytics:scorers"`, TTL 30 min.

### openfootball (internal)

```
GET https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.json
```

No auth. Single GET → all 104 matches with goal events.

**Cache:** `"wc:analytics:openfootball"`, TTL 15 min.

### Full response cache

**Cache:** `"wc:analytics:response"`, TTL 5 min. Wraps fully assembled response.

---

## Component Breakdown

### Backend

**`backend/internal/service/wc_football_data_client.go`** (new — distinct from existing `wc_football_client.go` which syncs match schedules)
```go
type WcFootballDataClient struct { apiKey string; httpClient *http.Client }
func NewWcFootballDataClient(apiKey string) *WcFootballDataClient
func (c *WcFootballDataClient) GetWCScorers(limit int) ([]model.WcScorer, time.Time, error)
```

**`backend/internal/service/wc_open_football_client.go`** (new)
```go
type WcOpenFootballClient struct { httpClient *http.Client }
func NewWcOpenFootballClient() *WcOpenFootballClient
func (c *WcOpenFootballClient) GetWCData() (*WcOpenFootballData, error)
```
Returns fully parsed struct with all openfootball-derived analytics.

**`backend/internal/repository/wc_repository.go`** (extend)
```go
func (r *WcRepository) GetCompletedMatchStats() (*model.WcMatchStats, error)
```

**`backend/internal/service/wc_analytics_service.go`** (extend existing)
```go
// New fields added to WcAnalyticsService:
// cache    *cache.Cache             // github.com/patrickmn/go-cache
// fdClient *WcFootballDataClient
// ofClient *WcOpenFootballClient

func (s *WcAnalyticsService) GetWorldCup2026Analytics() (*model.WcAnalyticsResponse, error)
```
Three-tier cache: response (5 min) → scorers (30 min) → openfootball (15 min).

**`backend/internal/api/wc_analytics_handler.go`** (extend existing)
```go
func (h *WcAnalyticsHandler) GetWorldCup2026Analytics(c *gin.Context)
```

**`backend/internal/api/router.go`** (extend — add inside existing `wcAuth` group)
```go
wcAuth.GET("/analytics/world-cup-2026", wcAnalyticsHandler.GetWorldCup2026Analytics)
```

### Frontend

The analytics page is **not a new route**. It's a new tab inside the existing `WcAnalyticsPanel.vue` which lives in `WcPredictView.vue`.

**`frontend/src/components/wc/WcAnalyticsPanel.vue`** (extend)
- Add "World Cup 2026" as the **first tab** (name=`"wc2026"`)
- Change default `analyticsTab` from `'my'` → `'wc2026'`
- Add `watch` case: `if (tab === 'wc2026' && !store.wc2026Data) store.loadWC2026Analytics()`
- Load wc2026 data in `onMounted` (since it's the default tab)

Tab order after change:
```
[World Cup 2026] [Của Tôi] [Cộng đồng] [So sánh]
     ↑ default
```

**`frontend/src/components/wc/analytics/WcTournamentPanel.vue`** (new)
- 11-section content component, receives `data: WcAnalyticsResponse | null` + `loading: boolean` props
- Renders all analytics sections (see UI layout below)

**`frontend/src/stores/wcAnalyticsStore.ts`** (extend)
- Add `wc2026Data: Ref<WcAnalyticsResponse | null>`
- Add `wc2026Loading: Ref<boolean>` (separate loading state — doesn't block other tabs)
- Add `loadWC2026Analytics()` action

**`frontend/src/services/wcAnalyticsService.ts`** (extend)
```ts
getWC2026Analytics(): Promise<WcAnalyticsResponse> {
  return wcApi.get<WcAnalyticsResponse>('/analytics/world-cup-2026').then(r => r.data)
}
```

**`frontend/src/types/wc.ts`** (extend) — add all new interfaces

**i18n locale files** (extend)
- `wc.analytics.wc2026Tab` → `"World Cup 2026"` / `"World Cup 2026"`

---

## UI Layout (11 sections)

```
┌─────────────────────────────────────────────────────────────┐
│  Thống kê WC2026                                             │
│                                                              │
│  ── Section 1: Tổng quan ────────────────────────────────── │
│  [Bàn thắng] [TB/trận] [Trận] [Sạch lưới] [Phản lưới] [PK] │
│     188        2.94     64      18           5          22   │
│                                                              │
│  ── Section 2: Vua phá lưới ─────────────────────────────── │
│  #  Cầu thủ        [Logo] Đội    Bàn  Kiến tạo              │
│  1  L. Messi         🇦🇷   ARG    5      —                  │
│  2  Vinicius Jr.     🇧🇷   BRA    4      1                  │
│  3  K. Mbappé        🇫🇷   FRA    4      2                  │
│  ...                                                         │
│  Cập nhật lúc: 14:30 hôm nay                                │
│                                                              │
│  ── Section 3: Kết quả trận đấu ─────────────────────────── │
│  Thắng sân nhà ██████████ 28 (43.8%)                        │
│  Thắng sân khách ███████  20 (31.2%)                        │
│  Hoà             █████    16 (25.0%)                        │
│                                                              │
│  ── Section 4: Phân tích hiệp đấu ───────────────────────── │
│  Hiệp 1: 82 bàn (43.6%)  vs  Hiệp 2: 106 bàn (56.4%)       │
│  Lội ngược dòng (thua H1 → thắng cả trận): 8 trận           │
│  Giữ lead (dẫn H1 → thắng cả trận): 32 trận                 │
│                                                              │
│  ── Section 5: Thời điểm ghi bàn ────────────────────────── │
│  1-15  ████ 18                                               │
│  16-30 ███████ 29                                            │
│  31-45 █████ 21                                              │
│  45+   ██ 14  (injury time H1)                               │
│  46-60 ████████ 34                                           │
│  61-75 ███████ 28                                            │
│  76-90 ██████████ 42                                         │
│  90+   ████ 22  (injury time H2)                             │
│                                                              │
│  ── Section 6: Bàn thắng theo vòng ──────────────────────── │
│  Group R32 R16 QF SF Final  (bar chart per stage)            │
│                                                              │
│  ── Section 7: Bàn thắng theo bảng đấu ──────────────────── │
│  Group A ████ 18   Group B ██ 12   Group C ██████ 24 ...    │
│                                                              │
│  ── Section 8: Trận nhiều bàn nhất ──────────────────────── │
│  Top 5 trận:                                                 │
│  1. Spain 4–4 Portugal · Group A · 8 bàn · 11/06            │
│  2. France 5–2 Belgium · R16 · 7 bàn · 01/07               │
│  3. ...                                                      │
│                                                              │
│  ── Section 9: Đội ghi nhiều bàn nhất ───────────────────── │
│  Đội          GF   GA   GD   Trận                           │
│  Spain         12    4   +8    5                             │
│  France        10    6   +4    5                             │
│  Brazil         9    3   +6    5                             │
│  ...                                                         │
│                                                              │
│  ── Section 10: Sân vận động ────────────────────────────── │
│  Venue                       Trận  Bàn  TB/trận             │
│  Mexico City                   8    24    3.0                │
│  Los Angeles (Inglewood)       6    18    3.0                │
│  New York/New Jersey           6    16    2.7                │
│  ...                                                         │
│                                                              │
│  ── Section 11: Phản lưới & Penalty ─────────────────────── │
│  [Phản lưới nhà: 5]  [Bàn từ penalty: 22]                   │
│  (% bàn từ PK: 11.7%)                                       │
└─────────────────────────────────────────────────────────────┘
```

**Mobile:** Stat cards wrap 2×3, charts collapse to full-width, team stats table scrolls horizontally.

---

## OpenFootballClient — parsing logic

```go
type WcOpenFootballData struct {
    GoalTiming        []model.WcGoalTimingBucket
    HalfTimeStats     model.WcHalfTimeStats
    TeamStats         []model.WcTeamStat
    GoalsByGroup      []model.WcGroupGoals
    TopScoringMatches []model.WcMatchDetail    // top 5 by total goals
    VenueStats        []model.WcVenueStat
}

func (c *OpenFootballClient) GetWCData() (*WcOpenFootballData, error) {
    // 1. GET worldcup.json
    // 2. Filter: only matches where score key exists (completed)
    // 3. Parse all goal events:
    //    - minute parsing: "90+4" → base=90, isInjury=true
    //    - bucket assignment (see below)
    //    - own goal / penalty counting
    // 4. Aggregate per team: goalsFor, goalsAgainst, matches
    // 5. Aggregate per group: matches, goals
    // 6. Aggregate per venue: matches, goals
    // 7. HT stats: compare score.ht vs score.ft for comebacks/held-lead
    // 8. Sort top 5 matches by total goals desc
}
```

**Minute parsing and bucket assignment:**
```go
// "90+4" → base 90, injury = true → bucket "90+"
// "45+3" → base 45, injury = true → bucket "45+"
// "67"   → base 67 → bucket "61-75"
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
```

**HT comeback detection:**
```go
// trailed at HT (ht[0] < ht[1]) but won at FT (ft[0] > ft[1]) → comeback for team1
// trailed at HT (ht[0] > ht[1]) but won at FT (ft[0] < ft[1]) → comeback for team2
```

**Mojibake fix (accented names):**
```go
func fixMojibake(s string) string {
    runes := []rune(s)
    b := make([]byte, len(runes))
    for i, r := range runes {
        if r > 0xFF { return s }
        b[i] = byte(r)
    }
    return string(b)
}
```
Apply to all `goal.Name` fields before aggregating.

---

## Design Decisions

### 1. DB for match-level stats, openfootball for event-level stats
DB (`wc_matches`) is authoritative for total matches, H/A/D, goals per stage. openfootball enriches with per-minute goal events, group field, HT score — none of which are in DB.

### 2. Three-tier go-cache (response 5min / scorers 30min / openfootball 15min)
- Response cache: absorbs traffic bursts (no redundant DB+API calls within 5 min window).
- Scorers cache: football-data.org free tier limit (10 req/min → max 2 calls/hour).
- openfootball cache: GitHub raw CDN, 15 min is sufficient freshness.

### 3. openfootball as single bulk GET
One HTTP call fetches all 104 matches. Much cheaper than ESPN's 64 separate calls per match. Trade-off: slightly less detailed stats (no passes, possession), but goal-level events are sufficient for this analytics page.

### 4. football-data.org is primary scorer source
Has assists + team crest URL. openfootball has goal events (for timing/own goals/penalties) but not assists. FD wins for the scorers table.

### 5. Graceful degradation at every tier
Match stats (DB) always return. Scorers and openfootball data are additive — each degrades independently. User always sees something useful.

### 6. ESPN team stats — deferred to Phase 2
ESPN unofficial API provides 28 team stats per match (possession, passes, shots on target...). Requires 64 calls + `wc_match_stats` DB table + cron sync. Out of scope for Phase 1.

---

## Non-Functional Requirements

- **Performance:** DB < 200ms + cache hit < 50ms. Cold start (both external APIs): < 4s total. Cache hit: < 300ms.
- **Rate limits:** football-data.org ≤ 2 calls/hour; openfootball: no limit (GitHub raw CDN).
- **Security:** `FOOTBALL_DATA_API_KEY` server-side only. openfootball is public, no credentials.
- **Reliability:** Two external providers both fail → DB stats still served (match stats always present in response).
- **Responsive:** Mobile layout: cards 2×3, charts full-width, team table horizontal scroll.
