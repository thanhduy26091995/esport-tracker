---
feature: frontend-integration
phase: requirements
status: draft
created: 2026-04-14
---

# Frontend Integration Requirements

## Problem Statement

**Current State:** Backend REST API is complete with 27 endpoints covering all business logic (users, matches, settlements, fund management, configuration). The API is tested and ready for integration.

**Problem:** Users have no way to interact with the system. All functionality exists only as API endpoints. We need a web-based user interface to enable score tracking, match recording, and fund management.

**Who It Impacts:**
- **Primary Users:** FC25 esport players who need to quickly record match results and track scores
- **Administrators:** Need to manage users, view leaderboard, monitor fund balance, handle settlements

## Goals

### Must Have (P0)
- ✅ User can view leaderboard with VND conversion (1 point = 22,000 VND)
- ✅ User can add/edit/delete players
- ✅ User can record match results (1v1 and 2v2)
- ✅ User can view match history
- ✅ System automatically triggers debt settlement at -6 points
- ✅ User can view settlement history
- ✅ User can manage fund (deposits/withdrawals)
- ✅ User can view fund balance and statistics

### Nice to Have (P1)
- Real-time updates when matches are recorded
- Mobile-responsive design for on-the-go score tracking
- Match statistics and analytics
- Export data (leaderboard, matches, settlements)

### Non-Goals
- Authentication/authorization (single-user system for now)
- Multi-tenancy (single tournament/group)
- Historical data migration from other systems
- Advanced analytics/charts (Phase 2 feature)

## User Stories

### Epic 1: Player Management
**As a tournament organizer,**
- I want to add new players to the system so they can participate in matches
- I want to edit player names to fix typos or handle name changes
- I want to remove players who are no longer participating
- I want to see all players sorted by their current score

### Epic 2: Match Recording
**As a player,**
- I want to quickly record a 1v1 match result so scores update automatically
- I want to record a 2v2 match result with team selection
- I want to see match history to verify recorded games
- I want to delete a mistakenly recorded match and have scores reverted

### Epic 3: Leaderboard & Scores
**As a player,**
- I want to view the leaderboard showing everyone's score in points
- I want to see VND equivalents next to point scores (22,000 VND per point)
- I want to see who is in debt (negative scores highlighted)
- I want to see my own match history and score changes

### Epic 4: Debt Settlement
**As a tournament organizer,**
- I want to see when a settlement is triggered automatically (at -6 points)
- I want to view settlement details (debtor, amount, distribution to winners and fund)
- I want to manually trigger a settlement if needed
- I want to see settlement history for auditing

### Epic 5: Fund Management
**As a tournament organizer,**
- I want to see the current fund balance
- I want to record deposits (e.g., initial fund, external contributions)
- I want to record withdrawals (e.g., equipment purchases, venue costs)
- I want to see transaction history with descriptions

### Epic 6: Configuration
**As a tournament organizer,**
- I want to adjust the debt threshold (currently -6 points)
- I want to change the point-to-VND conversion rate
- I want to modify the fund split percentage (currently 50/50)

## Success Criteria

### Functional Requirements
- [x] Backend API complete (27 endpoints)
- [ ] All views render correctly on desktop (1920x1080)
- [ ] All views are mobile-responsive (375px min width)
- [ ] Form validation matches backend validation
- [ ] Error messages from API are displayed to users
- [ ] Success messages confirm user actions

### User Experience
- [ ] Match recording takes < 10 seconds (3 clicks max)
- [ ] Leaderboard loads in < 1 second
- [ ] Vietnamese language support (UI can display Vietnamese names correctly)
- [ ] No page reloads required (SPA behavior)

### Performance
- [ ] Initial page load < 3 seconds
- [ ] API calls complete within 500ms (local network)
- [ ] Smooth transitions and animations (60fps)

### Accessibility
- [ ] Keyboard navigation works for all forms
- [ ] Color contrast meets WCAG AA standards
- [ ] Screen reader friendly (ARIA labels)

## Constraints & Assumptions

### Technical Constraints
- **Frontend Stack:** Vue 3 + TypeScript + Vite (already decided)
- **UI Framework:** Element Plus + Tailwind CSS (already decided)
- **State Management:** Pinia for global state, Vue Query for server state
- **Backend:** Go/Gin API at `http://localhost:8080/api/v1` (read-only, no changes)
- **Database:** PostgreSQL (managed by backend)

### Business Constraints
- **Timeline:** Complete frontend in 2-3 days of focused work
- **Team:** Single developer
- **Deployment:** Local development first, production later (Vercel planned)

### Assumptions
- Users have basic computer literacy
- Users understand point-based scoring systems
- Internet connection is stable (local network)
- Single concurrent user (no conflict resolution needed)
- Data is truth (no undo beyond match deletion)

## Open Questions

### Design Questions
- ❓ Should we have a dedicated "Dashboard" view or start with the leaderboard?
  - **Recommendation:** Start with Dashboard showing key stats (total players, recent matches, fund balance) then link to detail views
  
- ❓ Should settlement details show full match history or just summary?
  - **Recommendation:** Summary by default, expandable to full details

- ❓ How to handle Vietnamese player names in the UI?
  - **Recommendation:** Full Unicode support, test with Vietnamese diacritics

### UX Questions
- ❓ Should match recording be a modal or dedicated page?
  - **Recommendation:** Modal for quick access from any page

- ❓ Should we confirm before deleting a match (since it reverts scores)?
  - **Recommendation:** Yes, show confirmation dialog with impact preview

- ❓ Should configuration changes require confirmation?
  - **Recommendation:** Yes, especially for debt_threshold changes

### Technical Questions
- ❓ Should we cache API responses or always fetch fresh data?
  - **Recommendation:** Use Vue Query with 30-second stale time for leaderboard, fresh for match recording

- ❓ How to handle auto-settlement notifications to users?
  - **Recommendation:** Phase 1: Show in settlement history. Phase 2: Toast notifications

## Dependencies

### Backend (Complete ✅)
- All 27 API endpoints implemented and tested
- CORS enabled for frontend access
- Database seeded with default config

### Frontend (To Implement)
- Vue 3 project structure (✅ Already initialized)
- API service layer (TypeScript types + Axios)
- Pinia stores for each domain (user, match, settlement, fund, config)
- View components (Dashboard, Users, Matches, Settlements, Fund, Config)
- Form components (UserForm, MatchForm, FundForm)
- Shared components (Leaderboard, StatCard, MatchHistoryCard)

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| API contract mismatch (types) | High | Medium | Generate TypeScript types from Go models |
| State synchronization issues | Medium | Medium | Use Vue Query for automatic refetching |
| Form validation differs from backend | Medium | Low | Copy backend validation rules exactly |
| Settlement auto-trigger not visible to user | Low | High | Add settlement history view + notifications |
| Poor mobile experience | Medium | Medium | Mobile-first CSS approach |

## Next Steps

1. ✅ **Completed:** Backend API verification document created
2. **Next:** Create design document (`docs/ai/design/feature-frontend-integration.md`)
3. Then create implementation plan with task breakdown
4. Execute implementation in phases:
   - Phase 1: Core structure (types, services, stores)
   - Phase 2: User management UI
   - Phase 3: Match recording UI
   - Phase 4: Settlement & Fund UI
   - Phase 5: Configuration UI
   - Phase 6: Dashboard & Polish
