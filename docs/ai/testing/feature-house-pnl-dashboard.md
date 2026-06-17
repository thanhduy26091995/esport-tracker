---
phase: testing
title: House P&L Dashboard — Testing Strategy
description: Test scope cho house P&L feature
---

# Testing Strategy

## Scope

- **Unit:** Repository query logic — đặc biệt là void/pending/settled phân loại đúng
- **Integration:** Endpoint `/admin/house-pnl` với dữ liệu seed
- **Manual:** UI hiển thị đúng màu sắc, số liệu khớp, auto-refresh sau settle

## Test Files

| File | Layer | Coverage Target |
|---|---|---|
| `internal/repository/wc_repository_test.go` | Repository | P&L aggregate accuracy |

## Unit Tests — Repository

Seed các scenarios sau và verify kết quả:

| Scenario | Stake | Payout | Result | Expected |
|---|---|---|---|---|
| Win bet | 2 | 3 | `win` | Settled: stake=2, payout=3 → profit=-1 |
| Lose bet | 2 | 0 | `lose` | Settled: stake=2, payout=0 → profit=+2 |
| Void bet | 2 | 2 | `void` | Void bucket, không tính profit |
| Pending bet | 2 | NULL | NULL | Pending bucket |
| Win half | 2 | 2 | `win_half` | Settled: stake=2, payout=2 → profit=0 |

**Test cases:**
- `TestGetHousePnL_NoData`: DB trống → tất cả về 0
- `TestGetHousePnL_OnlyPending`: Chỉ có pending bets → settled=0, pending>0
- `TestGetHousePnL_MixedResults`: Win + lose + void + pending → đúng từng bucket
- `TestGetHousePnL_MatchBreakdown`: 2 trận, verify profit từng trận đúng

## Integration Tests

- `GET /admin/house-pnl` không có auth → 401
- `GET /admin/house-pnl` với user thường → 403
- `GET /admin/house-pnl` với admin → 200, đúng response shape
- Sau settle 1 trận, gọi lại `/admin/house-pnl` → `house_profit` thay đổi

## Execution

```bash
cd backend
go test ./internal/repository/... -run TestGetHousePnL -v
```

## Risks & Gaps

- `payout` column có thể NULL cho pending bets — `SUM(payout)` cần `FILTER` để không bị lỗi.
- Win half / lose half: `payout` là giá trị thực tế sau tính toán, không phải NULL — cần verify seeding đúng trong test.
