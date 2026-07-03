# Feature: Show O/U (Tài/Xỉu) + Custom-Bet Count on Match Cards & Admin List

**Date:** 2026-07-03
**Status:** Design approved (pending spec review)

## Problem

The player-facing match card (`WcMatchCard.vue`) currently displays only the
handicap line (which team gives, and the handicap value — no odds). It does not
surface Over/Under (Tài/Xỉu) odds or any indication that a match has custom bets
(kèo phụ). The admin match list (`WcAdminPanel.vue`) shows only sync-time chips
(HDP / O/U / Poisson), not the actual odds values or custom-bet presence.

Goal: on **both** surfaces, additionally show the O/U line + both odds, and a
count badge for open custom bets — while making the handicap line consistent by
adding its odds too.

## Data Availability

- **Handicap** (`handicap_value`, `handicap_team`, `odds_handicap_home`,
  `odds_handicap_away`) and **O/U** (`ou_line`, `odds_over`, `odds_under`) are
  already present on every `WcMatch` returned by `GET /wc/matches`
  (`WcService.ListMatches` → `WcRepository.ListMatches`). **No backend change
  needed** for handicap/O/U display on either surface.
- **Custom bets** are not in the list payload — they are fetched per-match lazily
  inside `WcPredictionForm.vue` via `GET /matches/:id/custom-bets`. A count badge
  therefore requires the matches-list response to carry a count field.

## Decisions

1. **Handicap line gets `@odds`** on the card, for visual parity with O/U.
2. **Custom-bet count scope = open only** (`wc_custom_bets.status = 'open'`) — the
   badge highlights bets the user can still join.

## Design

### 1. Backend — add `custom_bet_count` to matches list

**`backend/internal/model/wc_custom_bet.go`** — no change (status constants already
exist: `WcCustomBetStatusOpen = "open"`).

**`backend/internal/model/wc_match.go` (`WcMatch`)** — add a non-persisted response field:

```go
CustomBetCount int `gorm:"-" json:"custom_bet_count"`
```

**`backend/internal/repository/wc_custom_bet_repository.go`** — new batch counter
(mirrors existing `CountUnsettledForMatch` style):

```go
// CountOpenByMatchIDs returns the number of open custom bets per match,
// keyed by match ID. Matches with zero open bets are omitted from the map.
func (r *WcCustomBetRepository) CountOpenByMatchIDs(ids []uuid.UUID) (map[uuid.UUID]int, error) {
    if len(ids) == 0 {
        return map[uuid.UUID]int{}, nil
    }
    type row struct {
        MatchID uuid.UUID
        Count   int
    }
    var rows []row
    err := r.db.Model(&model.WcCustomBet{}).
        Select("match_id, COUNT(*) AS count").
        Where("match_id IN ? AND status = ?", ids, model.WcCustomBetStatusOpen).
        Group("match_id").
        Scan(&rows).Error
    if err != nil {
        return nil, err
    }
    out := make(map[uuid.UUID]int, len(rows))
    for _, r := range rows {
        out[r.MatchID] = r.Count
    }
    return out, nil
}
```

**`backend/internal/service/wc_service.go` (`ListMatches`)** — enrich after fetch:

```go
func (s *WcService) ListMatches(f repository.MatchFilter) ([]*model.WcMatch, error) {
    matches, err := s.repo.ListMatches(f)
    if err != nil {
        return nil, err
    }
    if s.customBetRepo != nil && len(matches) > 0 {
        ids := make([]uuid.UUID, len(matches))
        for i, m := range matches {
            ids[i] = m.ID
        }
        counts, err := s.customBetRepo.CountOpenByMatchIDs(ids)
        if err == nil { // best-effort enrichment; count defaults to 0 on error
            for _, m := range matches {
                m.CustomBetCount = counts[m.ID]
            }
        }
    }
    return matches, nil
}
```

Cost: +1 grouped query per matches-list call. Serves both player and admin
surfaces (both hit `GET /wc/matches`).

### 2. Frontend types

**`frontend/src/types/wc.ts` (`WcMatch`)** — add:

```ts
custom_bet_count?: number
```

### 3. Player card — `WcMatchCard.vue`

Under the score center, render a compact odds block (only the parts that have data):

- **Handicap** (existing label enhanced): keep "`{handicapTeamName} {gives} {handicap_value}`",
  then append odds `@{odds_handicap_home} / @{odds_handicap_away}` (formatted `toFixed(2)`,
  guarded when present).
- **O/U:** shown only when `match.ou_line` is set —
  `{ouShort} {ou_line} · {over} @{odds_over} / {under} @{odds_under}`.
- **Kèo phụ:** shown only when `custom_bet_count > 0` — a small badge
  `{custom_bet_count} {customBets}` reusing the existing `.wc-hc-label` chip style.

New computed helpers on the card for the odds strings; no new props (all fields on
`match`). No behavior change to actions slot.

### 4. Admin list — `WcAdminPanel.vue`

In the match-row info area (next to the existing sync chips at lines ~173–184), add
value chips (`.wc-sync-chip` style) shown only when data present:

- **O/U chip:** `T/X {ou_line} @{odds_over}/{odds_under}`.
- **Handicap chip:** `HDP {handicapSide} {handicap_value} @{home}/{away}` (parity).
- **Kèo phụ chip:** `{custom_bet_count} kèo phụ` (only when `> 0`).

These are read-only informational chips; the existing config buttons
(Chấp điểm / Tài Xỉu / Kèo phụ) are unchanged.

### 5. i18n

All new player-card labels go through `vue-i18n` (VI + EN) in the `wc.*` namespace:

- `wc.ouShort` → "T/X" / "O/U"
- `wc.over` → "Tài" / "Over"
- `wc.under` → "Xỉu" / "Under"
- `wc.customBets` → "kèo phụ" / "custom bets"

Admin panel strings in this codebase are hardcoded Vietnamese (existing pattern in
`WcAdminPanel.vue`), so the admin chips follow that file's existing convention.

## Out of Scope

- No change to how custom bets are placed/settled.
- No new per-match custom-bet fetch on the card (count comes from the list payload).
- No change to the prediction/betting dialog (`WcPredictionForm.vue`).

## Testing

- **Backend:** unit test `CountOpenByMatchIDs` (open counted; closed/settled/void
  excluded; empty ID slice → empty map; matches with zero omitted). Service test
  that `ListMatches` populates `CustomBetCount`.
- **Frontend:** `WcMatchCard` renders O/U block only when `ou_line` present, handicap
  odds appended, and badge only when `custom_bet_count > 0`. Verify in the running
  app on Lịch thi đấu / Dự đoán / Cược pages and the admin list.

## Files Touched

| Layer | File |
|-------|------|
| Model | `backend/internal/model/wc_match.go` (`WcMatch.CustomBetCount`) |
| Repository | `backend/internal/repository/wc_custom_bet_repository.go` (`CountOpenByMatchIDs`) |
| Service | `backend/internal/service/wc_service.go` (`ListMatches` enrichment) |
| Types | `frontend/src/types/wc.ts` (`WcMatch.custom_bet_count`) |
| Player card | `frontend/src/components/wc/WcMatchCard.vue` |
| Admin list | `frontend/src/components/wc/WcAdminPanel.vue` |
| i18n | `frontend/src/locales/vi.json`, `frontend/src/locales/en.json` (`wc.ouShort`, `wc.over`, `wc.under`, `wc.customBets`) |
