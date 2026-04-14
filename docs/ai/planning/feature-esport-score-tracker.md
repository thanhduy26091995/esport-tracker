---
phase: planning
title: FC25 Esport Score Tracker - Project Plan
description: Task breakdown and timeline for 2-3 week MVP development
feature: esport-score-tracker
created: 2026-04-14
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **Milestone 1:** Foundation Setup (Week 1, Days 1-2) - Dev environment + database ready
- [ ] **Milestone 2:** Core MVP (Week 1-2, Days 3-10) - User management + match recording working
- [ ] **Milestone 3:** Debt & Fund (Week 2, Days 11-14) - Auto settlement + fund tracking complete
- [ ] **Milestone 4:** Polish & Deploy (Week 3, Days 15-18) - Bug fixes, testing, production deploy

**Target:** 2.5 weeks (18 working days) for full MVP

## Task Breakdown

### Phase 1: Foundation & Setup (Days 1-2)

**Backend Setup (Day 1 - 4 hours)**
- [ ] **Task 1.1:** Initialize Go project with Gin framework
  - Install Go 1.21+, create module, add Gin dependency
  - Set up project structure (cmd, internal folders)
  - Create basic health check endpoint
  - **Estimate:** 1 hour
  
- [ ] **Task 1.2:** Database setup and migrations
  - Install PostgreSQL locally or set up Supabase account
  - Create database schema (users, matches, participants, settlements, fund, config)
  - Write migration scripts using golang-migrate
  - Seed initial config values
  - **Estimate:** 2 hours
  
- [ ] **Task 1.3:** Set up ORM and database connection
  - Configure GORM with PostgreSQL
  - Create model structs (User, Match, MatchParticipant, etc.)
  - Test database connection and basic CRUD
  - **Estimate:** 1 hour

**Frontend Setup (Day 2 - 4 hours)**
- [ ] **Task 1.4:** Initialize Vue 3 project with Vite + TypeScript
  - Run `npm create vite@latest frontend -- --template vue-ts`
  - Install dependencies: vue-router, pinia, axios, vue-query
  - Install Element Plus and Tailwind CSS
  - Configure Tailwind (tailwind.config.js, main CSS file)
  - **Estimate:** 1.5 hours
  
- [ ] **Task 1.5:** Set up project structure and routing
  - Create folder structure (views, components, stores, services, types)
  - Set up Vue Router with basic routes (Dashboard, Users, Matches, Fund, Settings)
  - Create AppLayout component with navigation
  - **Estimate:** 1.5 hours
  
- [ ] **Task 1.6:** Configure API service layer
  - Create Axios instance with base URL
  - Create service files (userService.ts, matchService.ts, fundService.ts)
  - Add TypeScript types for API responses
  - **Estimate:** 1 hour

**DevOps (Day 2 - 1 hour)**
- [ ] **Task 1.7:** Set up development workflow
  - Create .env files for frontend and backend
  - Set up CORS in backend for local development
  - Test frontend → backend connection
  - Create README with setup instructions
  - **Estimate:** 1 hour

**Total Phase 1: 9 hours (2 days)**

---

### Phase 2: User Management (Days 3-5)

**Backend (Day 3 - 5 hours)**
- [ ] **Task 2.1:** Implement User repository layer
  - Create UserRepository with methods: GetAll, GetByID, Create, Update, SoftDelete
  - Add proper error handling and validation
  - **Estimate:** 2 hours
  
- [ ] **Task 2.2:** Implement User service layer
  - Create UserService with business logic
  - Validate unique names, prevent duplicate users
  - Handle soft delete (is_active = false)
  - **Estimate:** 1.5 hours
  
- [ ] **Task 2.3:** Create User API endpoints
  - Implement handlers: GET /users, POST /users, PUT /users/:id, DELETE /users/:id
  - Add request validation and error responses
  - Test with Postman/curl
  - **Estimate:** 1.5 hours

**Frontend (Days 4-5 - 8 hours)**
- [ ] **Task 2.4:** Create User Pinia store
  - Define user state and actions (fetchUsers, createUser, updateUser, deleteUser)
  - Integrate with API service
  - **Estimate:** 1.5 hours
  
- [ ] **Task 2.5:** Build UsersView page
  - Create UserTable component (display users with current_score)
  - Add "Add User" button → UserForm modal
  - Add edit/delete buttons per user row
  - **Estimate:** 3 hours
  
- [ ] **Task 2.6:** Build UserForm component
  - Create form with name input (Element Plus form)
  - Add validation (required, unique name)
  - Handle create vs edit mode
  - **Estimate:** 2 hours
  
- [ ] **Task 2.7:** Build basic leaderboard on Dashboard
  - Fetch users sorted by current_score DESC
  - Display in table with rank, name, score, VND value
  - Add ScoreDisplay component (shows points + VND conversion)
  - **Estimate:** 1.5 hours

**Testing (Day 5 - 1 hour)**
- [ ] **Task 2.8:** Manual testing of user management
  - Test create, edit, delete flows
  - Verify unique name validation
  - Check leaderboard updates
  - **Estimate:** 1 hour

**Total Phase 2: 14 hours (3 days)**

---

### Phase 3: Match Management (Days 6-8)

**Backend (Days 6-7 - 8 hours)**
- [ ] **Task 3.1:** Implement Match repository layer
  - Create MatchRepository: Create, GetByID, GetAll (paginated), Update, Delete
  - Create MatchParticipantRepository: CreateBatch, GetByMatch, GetByUser
  - **Estimate:** 2.5 hours
  
- [ ] **Task 3.2:** Implement Match service layer with score updates
  - CreateMatch: Save match, create participants, update user scores
  - Validate team sizes (1 or 2 players per team)
  - Use database transaction for atomicity
  - Check if match is locked before update/delete
  - Revert scores on match deletion
  - **Estimate:** 3.5 hours
  
- [ ] **Task 3.3:** Create Match API endpoints
  - POST /matches - Create match with score updates
  - GET /matches - List recent matches (pagination)
  - GET /matches/:id - Match details with participants
  - PUT /matches/:id - Update match (if not locked)
  - DELETE /matches/:id - Delete and revert scores
  - **Estimate:** 2 hours

**Frontend (Days 7-8 - 9 hours)**
- [ ] **Task 3.4:** Create Match Pinia store
  - Define match state and actions
  - Handle create, fetch, update, delete matches
  - **Estimate:** 1.5 hours
  
- [ ] **Task 3.5:** Build CreateMatchView page
  - Create MatchForm component
  - Select match type (1v1 or 2v2)
  - Select players for Team 1 and Team 2 (dropdowns)
  - Select winner team (radio buttons)
  - Optional: match date/time picker
  - Optional: recorder name input
  - Display score preview before save
  - **Estimate:** 4 hours
  
- [ ] **Task 3.6:** Build MatchesView page
  - Display match history in list/table format
  - Show match type, teams, winner, date
  - Add edit/delete buttons (disabled if locked)
  - Add pagination
  - **Estimate:** 2.5 hours
  
- [ ] **Task 3.7:** Update Dashboard with recent matches
  - Add "Recent Matches" widget showing last 5-10 matches
  - **Estimate:** 1 hour

**Testing (Day 8 - 1.5 hours)**
- [ ] **Task 3.8:** Manual testing of match flows
  - Test 1v1 and 2v2 match creation
  - Verify scores update correctly
  - Test match edit and delete with score reversion
  - Edge case: Same player in both teams (should fail)
  - **Estimate:** 1.5 hours

**Total Phase 3: 18.5 hours (3 days)**

---

### Phase 4: Debt Settlement System (Days 9-11)

**Backend (Days 9-10 - 10 hours)**
- [ ] **Task 4.1:** Implement Settlement repository layer
  - Create DebtSettlementRepository: Create, GetAll, GetByUser, GetByID
  - **Estimate:** 1.5 hours
  
- [ ] **Task 4.2:** Implement Settlement service layer
  - CheckAndTriggerSettlement: Check each participant's score after match
  - CalculateSettlement: Get debt amount, convert to VND, split 50/50
  - Distribute to winners: Find recent opponents, distribute evenly
  - Reset debtor's score to 0 (add abs(debt) to current_score)
  - Reduce winners' scores by amount paid to them
  - Lock related matches
  - **Estimate:** 4.5 hours
  
- [ ] **Task 4.3:** Integrate settlement into match creation
  - Call settlement check after match save in MatchService
  - Return settlement info in match creation response
  - Handle transaction: rollback if settlement fails
  - **Estimate:** 2 hours
  
- [ ] **Task 4.4:** Create Settlement API endpoints
  - GET /settlements - List all settlements
  - GET /settlements/:id - Settlement details
  - GET /users/:id/settlements - User's settlement history
  - **Estimate:** 2 hours

**Frontend (Days 10-11 - 6 hours)**
- [ ] **Task 4.5:** Create Settlement store and service
  - Define settlement state and fetch actions
  - **Estimate:** 1 hour
  
- [ ] **Task 4.6:** Build SettlementModal component
  - Show debt details (debtor, amount, money)
  - Show fund allocation (50%)
  - Show winner distribution (50%)
  - Display after match creation if settlement triggered
  - **Estimate:** 2.5 hours
  
- [ ] **Task 4.7:** Add debt indicators to UI
  - Create DebtBadge component (shows warning if score < -3)
  - Add to leaderboard and user table
  - Show settlement history in user detail view
  - **Estimate:** 1.5 hours
  
- [ ] **Task 4.8:** Update Dashboard with settlement summary
  - Show "Recent Settlements" widget
  - **Estimate:** 1 hour

**Testing (Day 11 - 2.5 hours)**
- [ ] **Task 4.9:** Manual testing of debt settlement
  - Test player reaching -6 triggers settlement
  - Verify 50/50 split calculations
  - Verify score resets and winner score reductions
  - Verify fund balance updates
  - Test edge cases (multiple debtors, no recent opponents)
  - **Estimate:** 2.5 hours

**Total Phase 4: 18.5 hours (3 days)**

---

### Phase 5: Fund Management (Days 12-13)

**Backend (Day 12 - 4 hours)**
- [ ] **Task 5.1:** Implement Fund repository and service
  - FundTransactionRepository: Create, GetAll, GetBalance
  - FundService: AddFromSettlement, RecordExpense, GetBalance
  - Calculate balance: SUM(debt_in) - SUM(expense_out)
  - **Estimate:** 2.5 hours
  
- [ ] **Task 5.2:** Integrate fund into settlement
  - Create fund transaction when settlement occurs
  - Link to settlement record
  - **Estimate:** 1 hour
  
- [ ] **Task 5.3:** Create Fund API endpoints
  - GET /fund - Get balance and recent transactions
  - POST /fund/expense - Record expense
  - GET /fund/transactions - List all transactions
  - **Estimate:** 0.5 hours

**Frontend (Day 13 - 4.5 hours)**
- [ ] **Task 5.4:** Create Fund store and service
  - Fetch balance and transactions
  - Record expense action
  - **Estimate:** 1 hour
  
- [ ] **Task 5.5:** Build FundView page
  - Display current fund balance prominently
  - Show transaction history table (date, type, amount, description)
  - Add "Record Expense" button → modal form
  - **Estimate:** 2.5 hours
  
- [ ] **Task 5.6:** Add fund widget to Dashboard
  - FundBalance component showing current balance
  - Link to full fund page
  - **Estimate:** 1 hour

**Testing (Day 13 - 1 hour)**
- [ ] **Task 5.7:** Manual testing of fund tracking
  - Verify fund increases on debt settlement
  - Test expense recording
  - Verify balance calculation
  - **Estimate:** 1 hour

**Total Phase 5: 9.5 hours (2 days)**

---

### Phase 6: Configuration & Settings (Day 14)

**Backend (Day 14 - 2 hours)**
- [ ] **Task 6.1:** Implement Config service
  - GetConfig: Fetch all config key-value pairs
  - UpdateConfig: Update single config value
  - Validate config values (debt_threshold must be negative, point_to_vnd must be positive)
  - **Estimate:** 1.5 hours
  
- [ ] **Task 6.2:** Create Config API endpoints
  - GET /config
  - PUT /config/:key
  - **Estimate:** 0.5 hours

**Frontend (Day 14 - 3.5 hours)**
- [ ] **Task 6.3:** Create Config store
  - Fetch and update config
  - Use in other components (e.g., ScoreDisplay uses point_to_vnd)
  - **Estimate:** 1 hour
  
- [ ] **Task 6.4:** Build SettingsView page
  - Display config values in editable form
  - Fields: Debt Threshold, Point to VND Conversion, Fund Split %
  - Add validation and save button
  - **Estimate:** 2 hours
  
- [ ] **Task 6.5:** Use config in components
  - Update ScoreDisplay to use point_to_vnd from config
  - Update debt settlement logic to use debt_threshold from config
  - **Estimate:** 0.5 hours

**Total Phase 6: 5.5 hours (1 day)**

---

### Phase 7: Polish & Refinement (Days 15-16)

**UI/UX Polish (Day 15 - 5 hours)**
- [ ] **Task 7.1:** Improve Dashboard layout
  - Organize widgets (leaderboard, recent matches, settlements, fund)
  - Add statistics (total matches today, active players)
  - Make responsive for mobile
  - **Estimate:** 2.5 hours
  
- [ ] **Task 7.2:** Add Vietnamese language support
  - Add i18n if needed, or just hardcode Vietnamese labels
  - Fix VND formatting (use toLocaleString)
  - **Estimate:** 1.5 hours
  
- [ ] **Task 7.3:** Error handling and loading states
  - Add loading spinners during API calls
  - Add error messages for failed operations
  - Add success toasts for create/update/delete
  - **Estimate:** 1 hour

**Backend Improvements (Day 15 - 2 hours)**
- [ ] **Task 7.4:** Add validation and error messages
  - Improve validation error responses
  - Add proper HTTP status codes
  - **Estimate:** 1 hour
  
- [ ] **Task 7.5:** Add logging and request/response logging
  - Use Gin middleware for logging
  - **Estimate:** 1 hour

**Testing (Day 16 - 6 hours)**
- [ ] **Task 7.6:** Comprehensive end-to-end testing
  - Test complete user journey: add users → create matches → debt settlement → fund tracking
  - Test all edge cases documented in requirements
  - Test on different browsers (Chrome, Firefox, Safari)
  - Test on mobile devices (responsive)
  - **Estimate:** 4 hours
  
- [ ] **Task 7.7:** Bug fixes from testing
  - Fix issues found during testing
  - **Estimate:** 2 hours

**Total Phase 7: 13 hours (2 days)**

---

### Phase 8: Deployment (Days 17-18)

**Backend Deployment (Day 17 - 4 hours)**
- [ ] **Task 8.1:** Set up production database
  - Create Supabase project (or Railway PostgreSQL)
  - Run migrations on production database
  - Seed config values
  - **Estimate:** 1 hour
  
- [ ] **Task 8.2:** Deploy backend to Railway/Render
  - Create Railway/Render account
  - Connect GitHub repo
  - Configure environment variables
  - Deploy and test health check
  - **Estimate:** 2 hours
  
- [ ] **Task 8.3:** Configure production CORS
  - Update CORS to allow frontend domain
  - Test frontend → backend connection
  - **Estimate:** 1 hour

**Frontend Deployment (Day 17 - 3 hours)**
- [ ] **Task 8.4:** Configure frontend for production
  - Update API base URL to production backend
  - Build production bundle
  - Test build locally
  - **Estimate:** 1 hour
  
- [ ] **Task 8.5:** Deploy frontend to Vercel
  - Create Vercel account
  - Connect GitHub repo
  - Configure build settings
  - Deploy and test
  - **Estimate:** 1.5 hours
  
- [ ] **Task 8.6:** Set up custom domain (optional)
  - Configure DNS if custom domain available
  - **Estimate:** 0.5 hours

**Final Testing & Documentation (Day 18 - 4 hours)**
- [ ] **Task 8.7:** Production smoke testing
  - Test all features on production
  - Verify database persistence
  - Test from multiple devices
  - **Estimate:** 2 hours
  
- [ ] **Task 8.8:** Create user documentation
  - Write simple user guide (how to add users, record matches, etc.)
  - Create admin guide for configuration
  - **Estimate:** 1.5 hours
  
- [ ] **Task 8.9:** Create developer documentation
  - Update README with setup, development, and deployment instructions
  - Document API endpoints
  - Document database schema
  - **Estimate:** 0.5 hours

**Total Phase 8: 11 hours (2 days)**

---

## Dependencies

### Critical Path
```
Foundation → User Management → Match Management → Debt Settlement → Fund Management → Deploy
```

### Parallel Work Opportunities
- Frontend and backend for same feature can be developed in sequence
- UI polish can happen while testing
- Documentation can be written during deployment

### External Dependencies
- **Supabase/Railway account** - Needed by Day 17 (can create anytime)
- **Vercel account** - Needed by Day 17 (can create anytime)
- **Domain name** - Optional, not blocking

### Blockers
- Phase 2 (Users) must complete before Phase 3 (Matches) can start
- Phase 3 (Matches) must complete before Phase 4 (Debt) can start
- Phase 4 (Debt) must complete before Phase 5 (Fund) can fully work

## Timeline & Estimates

### Total Effort Estimate
- Phase 1: Foundation - 9 hours (2 days)
- Phase 2: User Management - 14 hours (3 days)
- Phase 3: Match Management - 18.5 hours (3 days)
- Phase 4: Debt Settlement - 18.5 hours (3 days)
- Phase 5: Fund Management - 9.5 hours (2 days)
- Phase 6: Configuration - 5.5 hours (1 day)
- Phase 7: Polish & Testing - 13 hours (2 days)
- Phase 8: Deployment - 11 hours (2 days)

**Total: 99 hours ≈ 18 working days (assuming 5-6 hours/day productivity)**

### Gantt Chart

```
Week 1:
Day 1-2:   [Foundation Setup]
Day 3-5:   [User Management]
Day 6-8:   [Match Management]

Week 2:
Day 9-11:  [Debt Settlement]
Day 12-13: [Fund Management]
Day 14:    [Configuration]

Week 3:
Day 15-16: [Polish & Testing]
Day 17-18: [Deployment]
```

### Target Milestones
- **End of Week 1:** Match recording working (no debt yet)
- **End of Week 2:** Full MVP features complete (debt + fund)
- **End of Week 3:** Deployed to production

### Buffer
- Built-in 10% buffer in each estimate
- Extra 2-3 days buffer recommended for unforeseen issues

## Risks & Mitigation

### Technical Risks

**Risk 1: Debt settlement logic complexity**
- **Impact:** High - Core feature
- **Probability:** Medium
- **Mitigation:** 
  - Write detailed unit tests for settlement calculations
  - Test with manual calculations first
  - Use database transactions to ensure atomicity

**Risk 2: Floating-point money calculation errors**
- **Impact:** High - Financial accuracy
- **Probability:** Low (using NUMERIC type)
- **Mitigation:**
  - Use PostgreSQL NUMERIC type for all money amounts
  - Test with edge cases (odd numbers, rounding)
  - Consider storing amounts in cents if issues arise

**Risk 3: Database transaction deadlocks**
- **Impact:** Medium
- **Probability:** Low (simple operations)
- **Mitigation:**
  - Keep transactions small and fast
  - Use proper locking (SELECT FOR UPDATE if needed)
  - Add retry logic for deadlock errors

**Risk 4: Frontend-backend API contract mismatch**
- **Impact:** Medium
- **Probability:** Medium
- **Mitigation:**
  - Define TypeScript types for all API responses
  - Test API with Postman before frontend integration
  - Use consistent naming conventions

### Resource Risks

**Risk 5: Single developer - knowledge gaps**
- **Impact:** Medium
- **Probability:** Medium
- **Mitigation:**
  - Use familiar tech stack (Vue 3, Go/Gin)
  - Consult documentation and examples
  - Ask for help in communities if stuck

**Risk 6: Scope creep - additional features requested**
- **Impact:** High - Timeline
- **Probability:** High
- **Mitigation:**
  - Stick to documented requirements
  - Create backlog for future enhancements
  - Defer non-MVP features to v2

### Dependency Risks

**Risk 7: Free tier limitations (Supabase, Vercel, Railway)**
- **Impact:** Low-Medium
- **Probability:** Low
- **Mitigation:**
  - Verify limits upfront (500MB DB, bandwidth limits)
  - Monitor usage during development
  - Have backup providers ready (Render, PlanetScale)

**Risk 8: Database migration issues in production**
- **Impact:** High
- **Probability:** Low
- **Mitigation:**
  - Test migrations on staging database first
  - Keep migration scripts reversible
  - Backup data before migrations

## Resources Needed

### Human Resources
- **1 Full-stack Developer** (you)
  - Backend: Go + PostgreSQL knowledge
  - Frontend: Vue 3 + TypeScript knowledge
  - DevOps: Basic deployment skills

### Tools & Services

**Development:**
- Go 1.21+ installed
- Node.js 18+ installed
- PostgreSQL 14+ (local or Docker)
- VS Code (or preferred IDE)
- Git
- Postman/Insomnia (API testing)

**Production Services (Free Tiers):**
- **Supabase** - PostgreSQL database (500MB free)
- **Railway/Render** - Backend hosting (512MB RAM free)
- **Vercel** - Frontend hosting (unlimited free)

**Optional:**
- **Sentry** - Error tracking (free tier)
- **Cloudflare** - CDN and DDoS protection (free)

### Documentation/Knowledge
- Go Gin documentation
- Vue 3 + TypeScript docs
- Element Plus component docs
- PostgreSQL documentation
- Supabase/Railway deployment guides

### Infrastructure
- **Development:** Local machine (macOS, Windows, or Linux)
- **Production:**
  - Database: Supabase PostgreSQL
  - Backend: Railway (or Render)
  - Frontend: Vercel
  - Total cost: $0/month (free tiers)

### Budget
- **Development:** $0 (using free tools and services)
- **Production:** $0 (using free tiers)
- **Contingency:** $10-20/month if need to upgrade (e.g., Railway paid tier if free tier insufficient)

---

## Next Steps

1. **Review this plan** - Confirm task breakdown makes sense
2. **Review requirements** - Run `/review-requirements` command to validate requirements doc
3. **Review design** - Run `/review-design` command to validate design doc
4. **Start development** - If reviews pass, run `/execute-plan` to begin implementation
5. **Store knowledge** - Save key decisions to AI DevKit memory as we progress

---

## Notes

- Timeline assumes 5-6 productive hours/day
- Each phase builds on previous phase - cannot parallelize much
- Testing is integrated into each phase, not separate
- Deployment is at the end, but can be set up earlier for staging
- Open questions from requirements doc should be resolved before starting Phase 4 (Debt Settlement)
