---
phase: testing
title: Bug Fix Batch — 21 June 2026 — Testing Strategy
description: Manual and automated test coverage for all 8 bug fixes
---

# Testing Strategy

## Scope

Manual smoke testing is the primary validation for the frontend-only bugs (1, 2, 3, 5). Backend logic bugs (4, 7, 8) can be tested with existing Go test infrastructure. Bug 6 (multi-pick champion) requires end-to-end testing across DB + backend + frontend.

---

## Test Files

| File | Package/Layer | Covers |
|------|---------------|--------|
| `backend/internal/service/wc_service_test.go` (existing or new) | Go service | Bug 4 (buildPreviewResult HouseSummary guard) |
| `backend/internal/service/wc_champion_service_test.go` (existing or new) | Go service | Bug 6 (multi-pick place/delete/settle) |
| `backend/internal/repository/wc_champion_repository_test.go` | Go repository | Bug 6 (DB constraint, GetMyPredictions) |
| Manual browser testing | Frontend | Bugs 1, 2, 3, 5, 6 (UI) |

---

## Unit Tests

### Bug 4 — `buildPreviewResult` HouseSummary guard

```go
// Test: already-settled match should not contribute to HouseSummary
func TestBuildPreviewResult_ExcludesAlreadySettledFromSummary(t *testing.T) {
    settledAt := time.Now()
    matches := []*model.WcMatch{
        {HomeScore: ptr(1), AwayScore: ptr(0), SettledAt: &settledAt},  // already settled
        {HomeScore: ptr(2), AwayScore: ptr(1), SettledAt: nil},          // new
    }
    // getPredictions returns 1 prediction per match (points = 5)
    result, err := buildPreviewResult(matches, mockPredictions, mockUserName)
    assert.NoError(t, err)
    // HouseSummary should only count the unsettled match
    assert.Equal(t, 1, result.HouseSummary.MatchCount)
    assert.Equal(t, float64(5), result.HouseSummary.TotalStaked)
}
```

### Bug 6 — Champion multi-pick

```go
// Test: placing two predictions on different teams succeeds
func TestPlaceChampionPrediction_MultiPick(t *testing.T) { ... }

// Test: placing prediction on same team twice returns error
func TestPlaceChampionPrediction_DuplicateTeamReturnsError(t *testing.T) { ... }

// Test: delete by ID with wrong userID returns error
func TestDeleteChampionPrediction_WrongOwnerReturnsError(t *testing.T) { ... }

// Test: SettleChampion pays out all matching predictions (two for same user)
func TestSettleChampion_MultipleCorrectPredictionsPerUser(t *testing.T) { ... }
```

### Bug 8 — JSON tag

No unit test needed — this is a struct tag change. Verified by running the existing integration test that calls `GET /admin/settlements/:id` and checking the JSON response contains `user_name` not `name`.

---

## Manual Test Checklist

### Bug 1 — Auto-redirect

- [ ] Log in as a WC user; navigate to `/world-cup` → should immediately redirect to `/world-cup/predict`
- [ ] Log out; navigate to `/world-cup` with feature enabled → should redirect to `/world-cup/login`
- [ ] Disable feature flag; navigate to `/world-cup` while logged in → should NOT redirect (show schedule)
- [ ] Admin user + feature disabled → stays on schedule, sees admin link

### Bug 2 — Scroll to next match

- [ ] Navigate to `/world-cup` as anonymous user; page should auto-scroll so that the first upcoming match date group is visible in the viewport
- [ ] Verify with past-only matches (all completed): page stays at top (no future match to scroll to)

### Bug 3 — Per-match collapse

- [ ] On predict page, expand match A's "Tất cả dự đoán" → predictions panel opens
- [ ] Expand match B's "Tất cả dự đoán" → match A should remain expanded (both visible)
- [ ] Click match A again → match A collapses; match B stays open
- [ ] All matches can be independently toggled

### Bug 4 — P&L preview only for new matches

- [ ] Finalize one match (admin: "Tính điểm" for a single match) → house P&L shows correct values for that one match
- [ ] Then run "Tính lại toàn bộ" → in the preview dialog, the `HouseSummary` at the bottom should only reflect non-already-settled matches (those without the "đã tính" tag)
- [ ] For "Tính điểm toàn bộ" with all unfinalized matches → HouseSummary shows totals for all listed matches

### Bug 5 — Champion tab responsive

- [ ] Open Chrome DevTools, set viewport to 375px wide; navigate to Vô địch tab
- [ ] Layout should be single column; no horizontal scrollbar
- [ ] Predictions table should scroll horizontally within its wrapper, not cause the page to overflow

### Bug 6 — Multi-pick champion

- [ ] Place prediction on Team A (5 pts) → prediction card appears
- [ ] Place prediction on Team B (3 pts) → second prediction card appears; both visible
- [ ] Attempt to place prediction on Team A again → error toast "already predicted this team"
- [ ] Delete Team A prediction by clicking its Xóa button → only Team A card disappears; Team B remains
- [ ] Admin settles champion with Team B as winner → Team B prediction is paid out correctly
- [ ] Public "Tất cả dự đoán" table shows all user rows (including multiple rows per user)

### Bug 7 — Smart cron

- [ ] Set a match to `status = 'live'` in DB; restart backend → check logs: cron should fire every 5 min
- [ ] Set match back to `status = 'scheduled'` (future date); check logs: cron returns to idle interval
- [ ] No live matches, next match within 2 hours → logs show 10-min interval

### Bug 8 — Settlement user name

- [ ] Admin creates a settlement
- [ ] Open settlement details → each row should show the player's name (not empty)
- [ ] Settlement preview table should also show `user_name` column filled

---

## Test Data & Environments

- **Bug 4, 7**: Use the existing local dev environment with seeded WC matches.
- **Bug 6**: Requires at least 2 champion teams in `wc_champion_teams` and an open champion config.
- **Bug 8**: Requires at least one completed settlement in the DB.

---

## Execution

```bash
# Backend tests
cd backend && go test ./internal/service/... -v -run TestBuildPreviewResult
cd backend && go test ./internal/service/... -v -run TestPlaceChampionPrediction
cd backend && go test ./internal/service/... -v -run TestSettleChampion

# Frontend type check (catches type errors from Bug 6 TS changes)
cd frontend && npm run type-check

# Backend build check
cd backend && go build ./...
```

---

## Coverage & Quality Gates

- All existing tests continue to pass after changes (no regressions).
- Type-check passes with zero errors on frontend.
- Backend compiles without warnings.
- Manual checklist above fully green for each bug before considering the milestone done.

---

## Risks & Gaps

- **Bug 7 cron**: Integration test for smart interval would require mocking time or DB state; this is deferred — covered by manual log observation.
- **Bug 2 scroll**: In some browsers/OS combinations, `scrollIntoView` with `behavior: 'smooth'` may not work (e.g., older iOS Safari); acceptable fallback is instant scroll.
- **Bug 6 settlement**: Multi-pick settlement is only tested manually; automated settlement test would require a full integration test with DB. Deferred to follow-up.
