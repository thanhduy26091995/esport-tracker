---
phase: implementation
title: WC Group Standings — Implementation Guide
description: Technical notes, patterns, and critical details for implementing the standings feature
---

# Implementation Guide

## Code Structure

New files:
- `frontend/src/components/wc/WcGroupStandings.vue` (new component)

Modified files:
- `backend/internal/model/wc_match.go` (add 3 structs)
- `backend/internal/repository/wc_repository.go` (add `GetGroupStandings`)
- `backend/internal/service/wc_service.go` (add `GetGroupStandings`)
- `backend/internal/api/wc_handler.go` (add `GetGroupStandings` handler method)
- `backend/internal/api/router.go` (wire route)
- `frontend/src/types/wc.ts` (add 3 types)
- `frontend/src/services/wcPublicApi.ts` (add `getStandings`)
- `frontend/src/views/WcScheduleView.vue` (fetch + render)
- `frontend/src/locales/vi.json` (add i18n keys)
- `frontend/src/locales/en.json` (add i18n keys)

## Implementation Notes

### Repository: GetGroupStandings

The computation is done in Go, not SQL, for clarity. SQL fetches the raw completed group matches; Go builds the standings map.

```go
func (r *WcRepository) GetGroupStandings() ([]model.WcGroupStanding, error) {
    var matches []model.WcMatch
    // Fetch ALL group-stage matches (any status) ordered by date.
    // Team roster is built from all matches; stats only from completed ones.
    err := r.db.Where("stage = ?", model.WcStageGroup).
        Order("match_date ASC").
        Find(&matches).Error
    if err != nil {
        return nil, err
    }

    // map[groupName]map[teamName]*teamAccumulator
    type acc struct {
        code          string
        played, won, drawn, lost int
        gf, ga        int
        recentResults []struct{ date time.Time; result string }
    }
    groups := map[string]map[string]*acc{}

    for _, m := range matches {
        g := m.GroupName
        if g == "" {
            continue
        }
        if groups[g] == nil {
            groups[g] = map[string]*acc{}
        }
        // Always register the team in the roster (even if match not yet played)
        if groups[g][m.HomeTeam] == nil {
            groups[g][m.HomeTeam] = &acc{code: m.HomeTeamCode}
        }
        if groups[g][m.AwayTeam] == nil {
            groups[g][m.AwayTeam] = &acc{code: m.AwayTeamCode}
        }
        // Only count stats for completed matches with valid scores
        if m.Status != model.WcStatusCompleted || m.HomeScore == nil || m.AwayScore == nil {
            continue
        }
        h := groups[g][m.HomeTeam]
        a := groups[g][m.AwayTeam]
        h.played++; a.played++
        h.gf += hs; h.ga += as_
        a.gf += as_; a.ga += hs
        switch {
        case hs > as_:
            h.won++; a.lost++
            h.recentResults = append(h.recentResults, struct{date time.Time; result string}{m.MatchDate, "W"})
            a.recentResults = append(a.recentResults, struct{date time.Time; result string}{m.MatchDate, "L"})
        case hs < as_:
            a.won++; h.lost++
            h.recentResults = append(h.recentResults, struct{date time.Time; result string}{m.MatchDate, "L"})
            a.recentResults = append(a.recentResults, struct{date time.Time; result string}{m.MatchDate, "W"})
        default:
            h.drawn++; a.drawn++
            h.recentResults = append(h.recentResults, struct{date time.Time; result string}{m.MatchDate, "D"})
            a.recentResults = append(a.recentResults, struct{date time.Time; result string}{m.MatchDate, "D"})
        }
    }

    // Build model.WcGroupStanding slice
    // Sort groups alphabetically, sort teams within each group by points/GD/GF/name
    // Form = last 5 results (already sorted by date ASC since matches were ordered ASC)
    ...
}
```

**Important**: matches are queried `ORDER BY match_date ASC`, so `recentResults` slice is already time-ordered. Take the last 5 entries (tail), then extract just the result strings for the `Form` field.

### Service: sorting

```go
sort.Slice(teams, func(i, j int) bool {
    a, b := teams[i], teams[j]
    if a.Points != b.Points { return a.Points > b.Points }
    if a.GoalDifference != b.GoalDifference { return a.GoalDifference > b.GoalDifference }
    if a.GoalsFor != b.GoalsFor { return a.GoalsFor > b.GoalsFor }
    return a.TeamName < b.TeamName
})
```

### Route wiring (router.go)

```go
// In the top-level wc group — same level as /matches and /config,
// OUTSIDE wcFeature (which requires WcFeatureMiddleware).
wc.GET("/standings", wcHandler.GetGroupStandings)
```

This mirrors the `/matches` route which is always accessible regardless of the feature flag.

### Frontend: getStandings (wcPublicApi.ts)

```ts
getStandings: () =>
  wcPublicHttp.get<WcStandingsResponse>('/standings').then(r => r.data),
```

Use `wcPublicApi` (no-auth axios instance), not `wcApi` (auth axios instance).

### Frontend: WcScheduleView integration

```ts
const standingsData = ref<WcStandingsResponse | null>(null)

// true when filter is '' or a knockout stage — show full 12-group grid
const showAllGroupsStandings = computed(() =>
  !selectedFilter.value.startsWith('group_')
)

// the matching single group when a 'group_X' filter is active
const selectedGroupStanding = computed<WcGroupStanding | null>(() => {
  if (!selectedFilter.value.startsWith('group_')) return null
  const groupName = selectedFilter.value.replace('group_', 'Group ')
  return standingsData.value?.groups.find(g => g.group_name === groupName) ?? null
})

// In onMounted (after store.fetchMatches):
try {
  standingsData.value = await getStandings()
} catch { /* standings are non-critical, fail silently */ }
```

**Template** (between `WcGroupFilter` and the match list):
```html
<!-- All-groups 2-column grid (default + knockout stage views) -->
<div v-if="standingsData && showAllGroupsStandings && !standingsData.groups.every(g => g.teams.length === 0)"
     class="wc-standings-grid">
  <WcGroupStandings
    v-for="group in standingsData.groups"
    :key="group.group_name"
    :standing="group"
  />
</div>

<!-- Single group (when 'group_X' filter active) -->
<WcGroupStandings
  v-else-if="selectedGroupStanding"
  :standing="selectedGroupStanding"
/>
```

Note: Knockout stage views (`r16`, `qf`, etc.) still show the 12-group grid as context above the knockout matches. If this becomes too much visual noise, it can be hidden for knockout stages later.

### Frontend: Flag emoji utility

Map ISO 3166-1 alpha-3 codes to emoji flags. Regional indicator letters use a Unicode offset:

```ts
// Convert 2-letter ISO country code to flag emoji
function isoAlpha2ToFlag(alpha2: string): string {
  return [...alpha2.toUpperCase()]
    .map(c => String.fromCodePoint(0x1F1E6 + c.charCodeAt(0) - 65))
    .join('')
}

// WC2026 alpha-3 → alpha-2 mapping (48 teams)
const TEAM_CODE_TO_ALPHA2: Record<string, string> = {
  ARG: 'AR', BRA: 'BR', FRA: 'FR', ESP: 'ES', ENG: 'GB',
  GER: 'DE', ITA: 'IT', POR: 'PT', NED: 'NL', BEL: 'BE',
  URU: 'UY', COL: 'CO', MEX: 'MX', USA: 'US', CAN: 'CA',
  MAR: 'MA', SEN: 'SN', NGA: 'NG', CIV: 'CI', CMR: 'CM',
  EGY: 'EG', GHA: 'GH', MLI: 'ML', TUN: 'TN', ALG: 'DZ',
  JPN: 'JP', KOR: 'KR', AUS: 'AU', IRN: 'IR', SAU: 'SA',
  QAT: 'QA', CHN: 'CN', NZL: 'NZ', ECU: 'EC', CHL: 'CL',
  PER: 'PE', BOL: 'BO', PAR: 'PY', VEN: 'VE', SUI: 'CH',
  AUT: 'AT', CRO: 'HR', CZE: 'CZ', POL: 'PL', UKR: 'UA',
  GRE: 'GR', SCO: 'GB', WAL: 'GB', SVK: 'SK',
  // add remaining 48 teams as needed
}

export function teamCodeToFlag(code: string): string {
  const alpha2 = TEAM_CODE_TO_ALPHA2[code.toUpperCase()]
  if (!alpha2) return '🏳️'
  return isoAlpha2ToFlag(alpha2)
}
```

**Note**: Scotland and Wales share 'GB' alpha-2 with England, so they'll show the same UK flag emoji. This is a known limitation — acceptable for a friend group app.

## Integration Points

### Public route group in router.go

Look for the section where `wcPublic` routes are declared (no auth middleware). Example pattern already in use:
```go
wcPublic := wc.Group("")
wcPublic.GET("/config", wcHandler.GetConfig)
wcPublic.GET("/matches", wcHandler.ListMatches)
wcPublic.GET("/schedule", wcHandler.GetSchedule)
// Add here:
wcPublic.GET("/standings", wcHandler.GetGroupStandings)
```

### wcPublicApi.ts base URL

Confirm the axios instance — it should use `VITE_API_BASE_URL + '/wc'` without an auth interceptor. Check the existing `getStandings` call pattern used by `WcUpcomingWidget`.

## Error Handling

- **Backend**: Return 500 with `{"error": "failed to fetch standings"}` on DB error
- **Frontend**: Wrap the standings fetch in try/catch; set `standingsData.value = null` on failure. The schedule page works fine without standings — they're supplementary.

## Performance Considerations

- WC2026 group stage = 12 groups × 6 matches each = 72 matches max, all completed by the end of the group stage. Query is fast (full table scan on a tiny table).
- No caching needed. The `/standings` response is fetched once per page load.
- Frontend: `standingsData` is reactive; switching group filters uses the already-fetched data (no re-fetch).
