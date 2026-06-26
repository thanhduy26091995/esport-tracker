---
phase: requirements
title: API Performance Optimization
description: Audit and optimize backend API to reduce response latency and resource usage under concurrent user load
---

# Requirements & Problem Understanding

## Problem Statement

**What problem are we solving?**

The system currently has multiple concurrent users, and API performance is degrading. Slow response times are noticeable in the UI, particularly on pages that rely on complex aggregation queries (leaderboard, analytics, match feed). The backend fetches all data from the database on every request with no caching, no connection pool tuning, and one critical endpoint (GET /matches) performs application-level pagination — loading all records into memory before slicing. As user count grows, these patterns compound.

- **Who is affected?** All authenticated WC users hitting the leaderboard, match feed, and analytics pages; admin users running settlements.
- **Current workaround:** None — users experience slow loads and occasional timeouts.

## Goals & Objectives

**Primary goals:**
- Reduce P95 API response time on the most-used endpoints (leaderboard, match feed, WC matches) by ≥50%
- Eliminate the application-level pagination anti-pattern in GET /matches
- Introduce a caching layer for heavy read endpoints that change infrequently
- Configure the DB connection pool to handle concurrent load properly

**Secondary goals:**
- Add DB indexes for frequently filtered/joined columns
- Reduce SQL verbosity in production logs (logger.Info → logger.Warn)
- Batch-insert custom bet options instead of looping
- Add response compression middleware (gzip)

**Non-goals:**
- Horizontal scaling / load balancing infrastructure changes
- Frontend bundle optimization or SSR
- Full Redis cluster setup — a simple local Redis instance or in-memory cache is sufficient
- Authentication flow changes

## User Stories & Use Cases

- **As a WC user**, I want the leaderboard to load in under 1s so I can quickly check standings after a match result.
- **As a WC user**, I want the match schedule and bet history pages to load instantly without perceived lag.
- **As an admin**, I want settlement and P&L queries to complete quickly even with hundreds of bets.
- **As any user**, I want the app to remain responsive when multiple friends are online simultaneously.

## Success Criteria

- GET /wc/leaderboard: response time ≤ 300ms (cached), ≤ 1s cold (currently ~800ms–2s under load)
- GET /matches: response time does not increase with total record count — DB-level pagination enforced
- GET /wc/matches: sub-200ms with caching
- No "fetch all" patterns in any paginated endpoint
- DB connection pool configured; no connection timeout errors under 20 concurrent users
- Load test: 20 concurrent requests to /wc/leaderboard complete without errors

## Constraints & Assumptions

- **Tech stack**: Go/Gin/GORM/PostgreSQL — no infrastructure changes, no new DBs
- **Caching**: Can use either go-cache (in-memory, zero infra) or Redis; Redis preferred if already available, otherwise go-cache for simplicity
- **No breaking API contract changes** — existing frontend clients must continue to work
- **Backward compatible DB migrations** — only ADD indexes, no column removals
- **Assumption**: Primary bottleneck is DB query cost + memory pressure, not network or app logic

## Questions & Open Items

- Is Redis already running in the deployment environment? (determines cache choice: Redis vs go-cache)
- What is the current p95 latency baseline? (need to measure before optimizing)
- Is there an existing APM or logging dashboard, or do we rely on the request_logger middleware?
- Should leaderboard cache invalidation be time-based (TTL) or event-driven (on score change)?
- Are there any DB index migrations already applied to production? (avoid duplicates)
