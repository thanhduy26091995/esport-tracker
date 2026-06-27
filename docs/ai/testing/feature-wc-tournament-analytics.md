---
phase: testing
title: WC Tournament Analytics — Testing Strategy
description: Unit tests, integration tests, and manual validation for the analytics feature
---

# Testing Strategy

## Test Coverage Goals

- Unit tests: `FootballDataClient` mapper, `GetCompletedMatchStats` SQL logic (via repo integration test), `GetTournamentAnalytics` service (cache hit/miss + FD failure)
- Integration tests: `GET /api/v1/wc/analytics` endpoint with seeded DB data
- Manual: verify numbers match actual WC2026 standings

## Unit Tests

### FootballDataClient
- [ ] Maps FD API response correctly to `[]WcScorer` (rank, name, team, goals, assists)
- [ ] Handles `assists: null` → Go `*int` nil
- [ ] Returns error on non-200 HTTP status
- [ ] Returns error on network failure (use httptest.Server)

### WcService.GetTournamentAnalytics
- [ ] Cache hit: returns cached scorers without calling FD client
- [ ] Cache miss: calls FD client, stores in cache, returns response
- [ ] FD client failure: returns match stats with empty `top_scorers`, no error returned to caller
- [ ] DB failure: returns error (propagated from repo)
- [ ] Zero matches played: all stats = 0, `highest_scoring_match` = nil

### WcRepository.GetCompletedMatchStats
- [ ] Correct `total_goals` for mixed match results
- [ ] Correct `home_wins` / `away_wins` / `draws` counts
- [ ] `clean_sheets`: counts matches where home=0 OR away=0 (not AND)
- [ ] `goals_by_stage`: groups correctly by stage, ordered by date
- [ ] `highest_scoring_match`: returns match with highest total goals
- [ ] Excludes matches with `status != 'completed'` or `home_score IS NULL`

## Integration Tests

- [ ] `GET /wc/analytics` with valid JWT → 200 + correct JSON shape
- [ ] `GET /wc/analytics` without JWT → 401
- [ ] Response with 0 completed matches → all zeros, no error
- [ ] Response with seeded matches → correct aggregates match hand-calculated values

## Test Data (seed helpers)

```go
// Helper: insert a completed match with given scores and stage
func seedCompletedMatch(db *gorm.DB, home, away int, stage string) *model.WcMatch {
    m := &model.WcMatch{
        HomeTeam: "Team A", AwayTeam: "Team B",
        HomeScore: &home, AwayScore: &away,
        Status: "completed", Stage: stage,
        MatchDate: time.Now(),
        ExternalID: uuid.New().String(),
    }
    db.Create(m)
    return m
}
```

Seed set for aggregate verification:
```
Match 1: 3-1 (home win, 4 goals, group)
Match 2: 0-2 (away win, 2 goals, group, home clean sheet)
Match 3: 1-1 (draw, 2 goals, group, neither clean)
Match 4: 4-4 (draw, 8 goals, r16 — highest scoring)
Expected: total=16, home_wins=1, away_wins=1, draws=2, clean_sheets=1
Expected highest: Match 4 (8 goals)
```

## Manual Testing

- [ ] Open `/world-cup/analytics` → page loads in < 2s
- [ ] Top scorers matches current WC2026 standings (Messi ~5, Vinicius ~4)
- [ ] Highest scoring match is plausible (high-scoring actual match)
- [ ] "Cập nhật lúc" timestamp shows recent time
- [ ] Mobile view: stat cards wrap to 2×2, scorers table scrolls horizontally
- [ ] If FD client fails (test with wrong API key temporarily): page still shows match stats, scorers shows empty state
- [ ] `vi` / `en` language switch: all labels translate correctly

## Performance Testing

- [ ] Response time < 500ms on cache hit (run twice, second call uses cache)
- [ ] Response time < 2s on cache miss (cold start)
