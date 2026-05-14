---
phase: testing
title: Dashboard Player Sort Strategy — Testing Strategy
description: Test scope, cases, and validation criteria for the player sort strategy feature
---

# Testing Strategy

## Scope

- **Unit tests:** `sortByStrategy()` utility (all strategies, all edge cases).
- **Component/integration tests:** `Leaderboard.vue` prop wiring, `DashboardView.vue` sort control interaction and localStorage persistence.
- **Manual smoke test:** Dashboard renders correctly with each strategy selected.

## Test Files

| File | Package/Layer | Coverage Target |
|---|---|---|
| `frontend/src/utils/sort.test.ts` | Utility | 100% branch coverage of `sortByStrategy` |
| `frontend/src/components/shared/Leaderboard.test.ts` | Component | `sortStrategy` prop changes produce correct `displayUsers` |

## Unit Tests

### `sortByStrategy()` — `sort.test.ts`

**Happy paths:**
- `debt-first` / `default`: negative group (worst first) → positive group (best first) → zero group (alpha) ✓
- `winners-first`: positive group (best first) → zero group (alpha) → negative group (worst first) ✓

**Within-group tie-break:**
- Two users with same score → sorted alphabetically by name ✓

**Edge cases:**
- All users score = 0 → all strategies produce alphabetical order ✓
- All users score > 0 → `debt-first`/`default`: pos → (no zero) → (no neg); `winners-first`: pos → same ✓
- All users score < 0 → `debt-first`/`default`: neg → (no pos) → (no zero); `winners-first`: (no pos) → (no zero) → neg ✓
- Single user → returned as-is ✓
- Empty array → returns `[]` ✓

**Sample test fixture:**
```ts
const players = [
  { id: '1', name: 'Alice', current_score: 5, ... },
  { id: '2', name: 'Bob',   current_score: -3, ... },
  { id: '3', name: 'Carol', current_score: 0, ... },
  { id: '4', name: 'Dave',  current_score: -10, ... },
  { id: '5', name: 'Eve',   current_score: 12, ... },
]
// debt-first expected: [Dave(-10), Bob(-3), Eve(12), Alice(5), Carol(0)]
// winners-first expected: [Eve(12), Alice(5), Carol(0), Dave(-10), Bob(-3)]
```

## Integration Tests

### `Leaderboard.vue`

- Passing `sort-strategy="debt-first"` renders debt holders at the top.
- Changing `sort-strategy` prop re-renders in the new order.
- `limit` prop still slices after sort.

### `DashboardView.vue`

- Clicking "Debt First" button activates it (primary type) and updates leaderboard order.
- `localStorage.getItem('dashboard-player-sort')` returns the new value after click.
- On mount with `localStorage` pre-set to `'winners-first'`, leaderboard opens in winners-first order.

## Test Data & Environments

- Fixtures: a mix of positive, negative, and zero-score users (see sample above).
- No environment variables needed; no external services called.
- Tests run in jsdom/happy-dom (Vitest).

## Execution

```bash
cd frontend
npm run test           # run all tests
npm run test -- sort   # run only sort utility tests
npm run coverage       # generate coverage report
```

## Coverage & Quality Gates

- `sort.ts`: 100% line + branch coverage required (it is pure logic with no side effects).
- `Leaderboard.vue` sort prop: all three strategy variants must have at least one test.
- No regressions on existing Leaderboard tests (score ordering in non-strategy contexts).

## Risks & Gaps

- No E2E test for localStorage persistence — covered by integration test using `localStorage` stub.
- `vi.json` translation correctness is manual — needs a native speaker review.
