---
feature: frontend-integration
phase: planning
status: ready
created: 2026-04-14
---

# Frontend Integration Implementation Plan

## Overview

**Goal:** Build complete Vue 3 frontend integrating with the 27 existing backend APIs

**Timeline:** 2-3 days (16-20 hours)

**Approach:** Incremental delivery in 6 phases, each building on the previous

## Implementation Phases

### Phase 1: Foundation Layer ⚙️
**Duration:** 2-3 hours

**Goal:** Set up type-safe API integration foundation

**Tasks:**

1. **TypeScript Types** (30 min)
   - [ ] Create `src/types/user.ts` with User, CreateUserRequest, UpdateUserRequest
   - [ ] Create `src/types/match.ts` with Match, MatchParticipant, CreateMatchRequest
   - [ ] Create `src/types/settlement.ts` with DebtSettlement, SettlementWinner
   - [ ] Create `src/types/fund.ts` with FundTransaction, CreateDepositRequest, CreateWithdrawalRequest
   - [ ] Create `src/types/config.ts` with Config, UpdateConfigRequest
   - [ ] Create `src/types/api.ts` with PaginatedResponse, ApiError, ApiResponse

2. **API Service Layer** (60 min)
   - [ ] Set up `src/services/api.ts` with Axios client + interceptors
   - [ ] Create `src/services/userService.ts` (7 endpoints)
   - [ ] Create `src/services/matchService.ts` (6 endpoints)
   - [ ] Create `src/services/settlementService.ts` (5 endpoints)
   - [ ] Create `src/services/fundService.ts` (5 endpoints)
   - [ ] Create `src/services/configService.ts` (3 endpoints)
   - [ ] Add error code mapping (VALIDATION_ERROR, NOT_FOUND, etc.)

3. **Pinia Stores** (60 min)
   - [ ] Create `src/stores/userStore.ts` (state, actions, getters)
   - [ ] Create `src/stores/matchStore.ts`
   - [ ] Create `src/stores/settlementStore.ts`
   - [ ] Create `src/stores/fundStore.ts`
   - [ ] Create `src/stores/configStore.ts`
   - [ ] Add loading states and error handling

4. **Utility Functions** (30 min)
   - [ ] Create `src/utils/formatters.ts` (formatVND, formatDate, formatScore)
   - [ ] Create `src/utils/validators.ts` (validateName, validateAmount)
   - [ ] Create `src/utils/constants.ts` (DEBT_THRESHOLD, API_BASE_URL)

**Deliverables:**
- Type-safe API layer
- Pinia stores with all CRUD operations
- Utility functions for formatting

**Success Criteria:**
- Can import types without errors
- API services compile without TS errors
- Stores can be imported and used in components

---

### Phase 2: User Management UI 👥
**Duration:** 2-3 hours

**Goal:** Complete user CRUD interface

**Tasks:**

1. **User Table Component** (60 min)
   - [ ] Create `src/components/user/UserTable.vue`
   - [ ] Display columns: Name, Score, VND Value, Status, Actions
   - [ ] Implement sorting by score
   - [ ] Add search by name filter
   - [ ] Add Create User button
   - [ ] Add Edit/Delete actions per row
   - [ ] Color code negative scores (red)

2. **User Form Component** (45 min)
   - [ ] Create `src/components/user/UserForm.vue`
   - [ ] Support both Create and Edit modes
   - [ ] Name input with validation (2-100 chars)
   - [ ] Form submission with loading state
   - [ ] Error message display
   - [ ] Success toast notification

3. **User View Page** (30 min)
   - [ ] Create `src/views/UsersView.vue`
   - [ ] Integrate UserTable component
   - [ ] Integrate UserForm dialog
   - [ ] Handle create/edit/delete operations
   - [ ] Add page title and description

4. **Testing** (15 min)
   - [ ] Test create user flow
   - [ ] Test edit user flow
   - [ ] Test delete user (with confirmation)
   - [ ] Test validation errors
   - [ ] Test duplicate name handling

**Deliverables:**
- Working user management interface
- CRUD operations functional
- Validation matching backend

**Success Criteria:**
- Can create/edit/delete users
- Validation errors show properly
- Score updates reflect immediately

---

### Phase 3: Match Recording UI 🎮
**Duration:** 3-4 hours

**Goal:** Quick match recording with team selection

**Tasks:**

1. **Match Form Modal** (90 min)
   - [ ] Create `src/components/match/MatchForm.vue`
   - [ ] Match type selector (1v1 / 2v2 radio buttons)
   - [ ] Team 1 player selection (multi-select, dynamic count)
   - [ ] Team 2 player selection
   - [ ] Winner selection (Team 1 / Team 2 radio)
   - [ ] Match date picker (optional, defaults to now)
   - [ ] Duplicate player validation
   - [ ] Show warning if player already in debt
   - [ ] Submit button with loading state

2. **Match List Component** (60 min)
   - [ ] Create `src/components/match/MatchList.vue`
   - [ ] Display: Date, Type, Teams, Winner, Point Changes
   - [ ] Color code winners (green) and losers (red)
   - [ ] Show locked matches (with lock icon)
   - [ ] Pagination support (20 per page)
   - [ ] Date filter (today, this week, this month)

3. **Recent Matches Widget** (30 min)
   - [ ] Create `src/components/match/RecentMatches.vue`
   - [ ] Show last 5 matches
   - [ ] Compact display format
   - [ ] Link to full match history

4. **Match View Page** (30 min)
   - [ ] Create `src/views/MatchesView.vue`
   - [ ] Integrate MatchList component
   - [ ] Add "Record Match" button (opens modal)
   - [ ] Show match stats (total, today)

5. **Testing** (30 min)
   - [ ] Test 1v1 match recording
   - [ ] Test 2v2 match recording
   - [ ] Test duplicate player validation
   - [ ] Test winner selection
   - [ ] Test locked match display
   - [ ] Verify score updates after match creation

**Deliverables:**
- Modal form for quick match recording
- Match history list with filters
- Recent matches widget for dashboard

**Success Criteria:**
- Can record 1v1 and 2v2 matches in <10 seconds
- Validation prevents invalid team compositions
- Match list shows all details correctly
- Locked matches clearly indicated

---

### Phase 4: Settlement & Leaderboard UI 💰
**Duration:** 3-4 hours

**Goal:** Display settlements, debt tracking, and leaderboard

**Tasks:**

1. **Settlement List Component** (60 min)
   - [ ] Create `src/components/settlement/SettlementList.vue`
   - [ ] Display: Date, Debtor, Debt Points, Money Amount, Fund Amount, Winner Distribution
   - [ ] Click to expand settlement details
   - [ ] Show winner breakdown in expanded view
   - [ ] Pagination support
   - [ ] Date filter

2. **Settlement Details Modal** (45 min)
   - [ ] Create `src/components/settlement/SettlementDetails.vue`
   - [ ] Show full settlement information
   - [ ] Winners table with money amounts and points deducted
   - [ ] Original debt points vs settlement amount
   - [ ] Related matches list (locked matches)

3. **Leaderboard Component** (60 min)
   - [ ] Create `src/components/shared/Leaderboard.vue`
   - [ ] Display: Rank, Name, Score, VND Value
   - [ ] Color code: Top 3 (gold/silver/bronze), Debt (red)
   - [ ] Show debt threshold indicator (-6 line)
   - [ ] Compact + full modes

4. **Settlement View Page** (30 min)
   - [ ] Create `src/views/SettlementsView.vue`
   - [ ] Integrate SettlementList
   - [ ] Show settlement stats (total, today)
   - [ ] Show current debtors count

5. **Testing** (45 min)
   - [ ] Test settlement display after auto-trigger
   - [ ] Test settlement details modal
   - [ ] Test winner distribution breakdown
   - [ ] Verify leaderboard sorting
   - [ ] Check debt threshold highlighting

**Deliverables:**
- Settlement history view
- Settlement details modal
- Leaderboard component (reusable)

**Success Criteria:**
- Settlements show complete information
- Winner distribution adds up correctly
- Leaderboard updates after matches
- Debt threshold clearly visible

---

### Phase 5: Fund Management UI 💵
**Duration:** 2-3 hours

**Goal:** Fund balance, transactions, deposits, withdrawals

**Tasks:**

1. **Fund Stats Card** (30 min)
   - [ ] Create `src/components/fund/FundStats.vue`
   - [ ] Show current balance (large, prominent)
   - [ ] Show total deposits
   - [ ] Show total withdrawals
   - [ ] Show settlement deposits count

2. **Transaction List Component** (60 min)
   - [ ] Create `src/components/fund/TransactionList.vue`
   - [ ] Display: Date, Type, Amount, Description, Related Settlement
   - [ ] Color code: Deposits (green), Withdrawals (red)
   - [ ] Support filtering by type
   - [ ] Pagination support
   - [ ] Click settlement link to view details

3. **Deposit/Withdrawal Forms** (60 min)
   - [ ] Create `src/components/fund/FundForm.vue`
   - [ ] Support both deposit and withdrawal modes
   - [ ] Amount input (positive numbers only, min: 1000 VND)
   - [ ] Description input
   - [ ] Date picker (optional)
   - [ ] Show current balance
   - [ ] Validate withdrawal doesn't exceed balance

4. **Fund View Page** (30 min)
   - [ ] Create `src/views/FundView.vue`
   - [ ] Show FundStats at top
   - [ ] Show TransactionList below
   - [ ] Add Deposit/Withdrawal buttons
   - [ ] Integrate FundForm dialog

5. **Testing** (30 min)
   - [ ] Test deposit creation
   - [ ] Test withdrawal (with insufficient balance check)
   - [ ] Test transaction list filtering
   - [ ] Test settlement deposit display
   - [ ] Verify balance updates

**Deliverables:**
- Fund balance dashboard
- Transaction history
- Deposit/withdrawal forms

**Success Criteria:**
- Balance displays correctly
- Cannot withdraw more than available
- Transactions show proper type indicators
- Settlement deposits are trackable

---

### Phase 6: Dashboard & Configuration UI 📊
**Duration:** 3-4 hours

**Goal:** Main dashboard and system configuration

**Tasks:**

1. **Stat Cards Component** (45 min)
   - [ ] Create `src/components/shared/StatCard.vue`
   - [ ] Support icon, title, value, trend
   - [ ] Color coding per stat type
   - [ ] Loading skeleton
   - [ ] Responsive grid layout

2. **Dashboard View** (90 min)
   - [ ] Create `src/views/DashboardView.vue`
   - [ ] Stat cards: Total Players, Today Matches, Fund Balance, Debtors
   - [ ] Leaderboard widget (top 10)
   - [ ] Recent matches widget (last 5)
   - [ ] Quick actions: Record Match, Add User
   - [ ] Refresh all data on mount

3. **Configuration View** (60 min)
   - [ ] Create `src/views/ConfigView.vue`
   - [ ] Form for debt_threshold (negative integer)
   - [ ] Form for point_to_vnd (positive integer)
   - [ ] Form for fund_split_percent (0-100 integer)
   - [ ] Show current values prominently
   - [ ] Validation matching backend rules
   - [ ] Save confirmation dialog

4. **Layout & Navigation** (60 min)
   - [ ] Create `src/layouts/MainLayout.vue`
   - [ ] Sidebar navigation with icons
   - [ ] Top bar with title
   - [ ] Mobile responsive (drawer on mobile)
   - [ ] Active route highlighting

5. **Polish & Testing** (45 min)
   - [ ] Add loading states everywhere
   - [ ] Add empty states (no users, no matches)
   - [ ] Add error states (API down)
   - [ ] Test full user journey: Create user → Record match → View settlement → Check fund
   - [ ] Test on mobile viewport
   - [ ] Fix any UI/UX issues

**Deliverables:**
- Complete dashboard with all widgets
- System configuration management
- Navigation layout
- Polish and error handling

**Success Criteria:**
- Dashboard loads all data in <3 seconds
- Config changes save successfully
- Navigation works on mobile
- All empty/error states look good

---

## Dependencies & Ordering

```
Phase 1 (Foundation)
    ↓
    ├─→ Phase 2 (Users) ────────┐
    ├─→ Phase 3 (Matches) ──────┤
    ├─→ Phase 4 (Settlements) ──┼─→ Phase 6 (Dashboard + Config)
    └─→ Phase 5 (Fund) ─────────┘
```

**Critical Path:** Phase 1 → Phase 2 → Phase 3 → Phase 6

**Parallel Work:** Phases 2, 3, 4, 5 can be developed in parallel after Phase 1

## Resource Requirements

### NPM Packages Needed
```bash
npm install pinia vue-router axios element-plus @element-plus/icons-vue
npm install -D tailwindcss postcss autoprefixer
npm install -D @types/node vitest @vue/test-utils
```

### Configuration Files
- `tailwind.config.js` - Configure Tailwind utilities
- `vite.config.ts` - Configure build, proxy if needed
- `.env.development` - Local API URL
- `.env.production` - Production API URL

## Testing Strategy

### Manual Testing Checklist
After each phase:
- [ ] All API calls work correctly
- [ ] Loading states display
- [ ] Error states display
- [ ] Form validation works
- [ ] Success notifications appear

### Integration Testing
After Phase 6:
- [ ] Create user → Record match → Check leaderboard update
- [ ] Record multiple matches → Trigger settlement → Check fund increase
- [ ] Make withdrawal → Check balance update
- [ ] Change config → Record match → Verify new rules apply

## Risk Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| API response format doesn't match types | Low | High | Test all endpoints in Phase 1 |
| Element Plus version incompatibility | Low | Medium | Lock versions in package.json |
| CORS issues in production | Medium | High | Configure backend CORS correctly |
| Performance issues with large datasets | Medium | Medium | Implement pagination early |
| Mobile layout broken | Low | Medium | Test mobile viewport in each phase |

## Success Metrics

### Functional Metrics
- ✅ All 27 backend endpoints integrated
- ✅ All CRUD operations working
- ✅ All validations matching backend
- ✅ All features implemented per requirements

### UX Metrics
- ✅ Match recording < 10 seconds
- ✅ Page load < 3 seconds
- ✅ No UI jank or flicker
- ✅ Mobile-responsive on all screens

### Code Quality Metrics
- ✅ No TypeScript errors
- ✅ No console errors
- ✅ Consistent component structure
- ✅ Proper error handling everywhere

## Rollout Plan

### Development Environment
1. Complete Phase 1
2. Test all API services
3. Complete Phases 2-5
4. Complete Phase 6
5. Full integration testing

### Production Deployment
1. Build production bundle
2. Deploy to Vercel
3. Configure environment variables
4. Test full flow in production
5. Monitor errors

## Post-Launch Tasks

### Immediate (Next sprint)
- [ ] Add unit tests for stores
- [ ] Add component tests
- [ ] Set up error tracking (Sentry)
- [ ] Monitor performance metrics

### Future Enhancements
- [ ] Real-time updates (WebSocket)
- [ ] User authentication
- [ ] Match history export (CSV)
- [ ] Data visualization charts
- [ ] Dark mode

## Open Questions

1. **Should we add optimistic UI updates?**
   - Pro: Better UX, feels faster
   - Con: More complex state management
   - **Decision:** Yes, for create operations only

2. **Should we cache API responses?**
   - Pro: Faster page loads, less API calls
   - Con: Stale data risk
   - **Decision:** Yes, with 30s TTL for leaderboard, no cache for transactions

3. **Should we add keyboard shortcuts?**
   - Pro: Power user efficiency (e.g., Cmd+M to record match)
   - Con: More development time
   - **Decision:** Add in post-launch if requested

## Next Step

**Action:** Start Phase 1 - Foundation Layer

**Command:**
```bash
cd frontend
npm install pinia vue-router axios element-plus @element-plus/icons-vue
npm install -D tailwindcss postcss autoprefixer @types/node
npx tailwindcss init -p
```

Then begin creating TypeScript types in `src/types/` directory.
