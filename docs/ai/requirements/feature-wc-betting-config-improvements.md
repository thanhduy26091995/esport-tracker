---
phase: requirements
title: WC Betting Config Improvements
description: Configurable min/max bet via admin, handicap display in predict tabs, and consistent bet-type labels system-wide
---

# Requirements & Problem Understanding

## Problem Statement

Three related pain points in the WC betting/prediction UX:

1. **Hardcoded bet limits** — Min (1 điểm) and max (5 điểm) for all prediction types (handicap, exact score, over/under, champion) are scattered across frontend components and backend Gin binding tags. Adjusting them requires a code deploy, which is impractical for a live event.

2. **Missing handicap context in predict history** — On `/world-cup/predict`, the "Đang chờ kết quả" and "Lịch sử" tabs show a user's past predictions but do not show which team gives handicap and by how much. Without that context, users cannot verify or understand their own bets.

3. **Inconsistent bet-type label rendering** — The labels "Kèo Handicap", "Kèo tỉ số", "Kèo tài xỉu" are already defined in `vi.json` i18n keys, but some components render raw type strings (`handicap`, `exact_score`, `over_under`) or use ad-hoc translations instead of the shared keys, creating an inconsistent experience.

**Who is affected:** All WC players and the admin managing the event.

**Current workaround:**
- Admin asks developer to change constants and redeploy.
- Users must navigate to the match detail to find handicap info.
- Label inconsistency is a visual defect with no workaround.

## Goals & Objectives

### Primary Goals
- Allow admin to set `min_points` and `max_points` per prediction (including champion) through the existing admin UI, with immediate effect on all bet forms.
- Display handicap info (who gives handicap, amount) next to each prediction in the pending and history tabs.
- Use consistent i18n-driven bet-type labels (`predictionTypeHandicap`, `predictionTypeExactScore`, `predictionTypeOverUnder`) everywhere a prediction type is shown.

### Secondary Goals
- Backend validation must read limits dynamically from `wc_config` so no code change is needed for limit adjustments.
- Bet-type label component/composable is the single source of truth — no inline label strings.

### Non-Goals
- Per-match or per-user bet limits.
- Adding new bet types (e.g. "Kèo góc") — that ships in a separate feature. However, the label utility built here **must be extensible** so adding a future type requires only one file + one i18n key.
- VND currency limits — the config applies to điểm (points) only, for both predictions and wallet bets.

## User Stories & Use Cases

**Admin:**
- As an admin, I want to raise the max prediction from 5 to 10 điểm mid-tournament so that top players can make higher-stakes predictions, without needing a code deploy.
- As an admin, I want to lower the min prediction to 1 điểm so that casual players can join with a smaller stake.

**Player:**
- As a player on the "Đang chờ kết quả" tab, I want to see "Việt Nam chấp Morocco 0.5 trái" next to my handicap prediction so I can remember what I bet without navigating away.
- As a player, I want to see the same label "Kèo Handicap" whether I'm looking at the bet form, prediction history, or the match detail page.

## Success Criteria

1. Admin sets `min_points = 2`, `max_points = 10` via the admin panel; on next page load the prediction and bet forms enforce those limits with no redeploy.
2. Backend returns 400 for any prediction or wallet bet where stake < `min_points` or stake > `max_points` (read from `wc_config` at request time).
3. "Đang chờ kết quả" and "Lịch sử" tabs display handicap team names and handicap value for all handicap-type predictions.
4. A grep for raw `"handicap"`, `"exact_score"`, `"over_under"` strings in template/render contexts returns zero results — all replaced with `betTypeLabel()` composable calls.
5. Existing pending predictions/bets are unaffected by a limit change — only new submissions are validated against the updated limits.

## Constraints & Assumptions

- `wc_config` is a single-row table (id=1); adding `min_points` / `max_points` columns is simplest.
- `min_points` and `max_points` are **integers** (e.g. 1, 2, 5) — DB column is `INT`. Frontend input uses default step of 1.
- The config applies to **both** `wc_predictions` (points-based) **and** `wc_bets` (wallet stake) — same column, same validation logic in both service paths.
- Champion prediction limits use the same `wc_config` values.
- All frontend forms already use `<el-input-number>` — only `:min`, `:max`, `:step` bindings need to change.
- Backend currently validates `min=1,max=5` via Gin binding struct tags; dynamic validation must move to service layer. Existing pending records are not retroactively voided by a limit change.
- Handicap data (`home_handicap`, `away_handicap`, home/away team names) is already returned by match endpoints; it just needs to be surfaced in the prediction list components.
- Enforcement timing: limits are read from the Pinia config store (fetched on page load). Real-time enforcement is not required — the backend always validates at request time, so a stale frontend limit only causes a backend 400, not silent acceptance.

## Questions & Open Items

- Handicap display format: "Vietnam chấp Morocco 0.5 trái" vs "Vietnam -0.5 / Morocco +0.5"? → Use narrative form ("X chấp Y Z trái") — matches how Vietnamese bettors talk about it. Confirm during design review.
