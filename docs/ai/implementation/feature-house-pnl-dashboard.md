---
phase: implementation
title: House P&L Dashboard — Implementation Guide
description: Technical notes, patterns, và code snippets
---

# Implementation Guide

## Code Structure

```
backend/internal/
  model/wc_match.go          ← ADD: HousePnLResponse, HousePnLMatch structs
  repository/wc_repository.go ← ADD: GetHousePnL() method
  service/wc_service.go       ← ADD: GetHousePnL() thin wrapper
  api/wc_handler.go           ← ADD: GetHousePnL handler
  cmd/server/main.go          ← ADD: route GET /admin/house-pnl

frontend/src/
  types/wc.ts                 ← ADD: HousePnLResponse, HousePnLMatch
  services/wcService.ts       ← ADD: getHousePnL()
  components/wc/
    WcHousePnL.vue            ← NEW component
    WcAdminPanel.vue          ← EDIT: mount WcHousePnL + wire settle event
```

## Implementation Notes

### Repository — raw SQL qua GORM

Dùng `r.db.Raw(...)` thay vì GORM ORM vì query có `FILTER (WHERE ...)` PostgreSQL-specific:

```go
func (r *WcRepository) GetHousePnL() (*model.HousePnLResponse, error) {
    var summary struct {
        TotalStakeSettled  int `gorm:"column:total_stake_settled"`
        TotalPayoutSettled int `gorm:"column:total_payout_settled"`
        SettledBetCount    int `gorm:"column:settled_bet_count"`
        TotalStakeVoid     int `gorm:"column:total_stake_void"`
        TotalStakePending  int `gorm:"column:total_stake_pending"`
        PendingBetCount    int `gorm:"column:pending_bet_count"`
    }
    err := r.db.Raw(`
        SELECT
            COALESCE(SUM(stake)   FILTER (WHERE result IS NOT NULL AND result != 'void'), 0) AS total_stake_settled,
            COALESCE(SUM(payout)  FILTER (WHERE result IS NOT NULL AND result != 'void'), 0) AS total_payout_settled,
            COALESCE(COUNT(*)     FILTER (WHERE result IS NOT NULL AND result != 'void'), 0) AS settled_bet_count,
            COALESCE(SUM(stake)   FILTER (WHERE result = 'void'), 0)                        AS total_stake_void,
            COALESCE(SUM(stake)   FILTER (WHERE result IS NULL), 0)                         AS total_stake_pending,
            COALESCE(COUNT(*)     FILTER (WHERE result IS NULL), 0)                         AS pending_bet_count
        FROM wc_bets
    `).Scan(&summary).Error
    if err != nil {
        return nil, err
    }

    var matches []model.HousePnLMatch
    err = r.db.Raw(`
        SELECT
            b.match_id::text AS match_id,
            m.home_team, m.away_team,
            m.match_date::text AS match_date,
            m.stage,
            SUM(b.stake)::int   AS stake,
            SUM(b.payout)::int  AS payout,
            COUNT(b.id)::int    AS bet_count
        FROM wc_bets b
        JOIN wc_matches m ON m.id = b.match_id
        WHERE b.result IS NOT NULL AND b.result != 'void'
        GROUP BY b.match_id, m.home_team, m.away_team, m.match_date, m.stage
        ORDER BY (SUM(b.stake) - SUM(b.payout)) ASC
    `).Scan(&matches).Error
    if err != nil {
        return nil, err
    }

    for i := range matches {
        matches[i].Profit = matches[i].Stake - matches[i].Payout
    }

    return &model.HousePnLResponse{
        TotalStakeSettled:  summary.TotalStakeSettled,
        TotalPayoutSettled: summary.TotalPayoutSettled,
        HouseProfit:        summary.TotalStakeSettled - summary.TotalPayoutSettled,
        TotalStakeVoid:     summary.TotalStakeVoid,
        TotalStakePending:  summary.TotalStakePending,
        PendingBetCount:    summary.PendingBetCount,
        SettledBetCount:    summary.SettledBetCount,
        MatchBreakdown:     matches,
        GeneratedAt:        time.Now().Format(time.RFC3339),
    }, nil
}
```

### Frontend — WcHousePnL.vue layout

```
┌─────────────────────────────────────────────┐
│  House P&L                     [Refresh 🔄] │
├──────────────┬──────────────┬───────────────┤
│  Stake thu   │  Payout trả  │  Lời/Lỗ       │
│  1,250 pts   │  1,100 pts   │  +150 pts ✅  │
├─────────────────────────────────────────────┤
│  ⏳ Pending: 80 pts (12 bets chờ settle)    │
├─────────────────────────────────────────────┤
│  [Xem chi tiết theo trận ▼]                 │
│  Brazil vs Argentina  120 | 95  | +25       │
│  France vs Spain      100 | 130 | -30 🔴   │
│  ...                                         │
└─────────────────────────────────────────────┘
```

### Auto-refresh sau settle

Trong `WcAdminPanel.vue`, sau khi `SettleMatch` thành công:
```ts
// Emit hoặc dùng ref để trigger reload trong WcHousePnL
await wcService.settleMatch(matchId)
pnlRef.value?.reload()  // nếu dùng expose/ref
// hoặc dùng một reactive key để remount component
```

## Security Notes

- Route `GET /admin/house-pnl` **bắt buộc** dưới `WcAdminMiddleware`.
- Không expose raw query hay DB details trong response lỗi.
