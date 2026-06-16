---
phase: implementation
title: WC Champion Prediction — Implementation Guide
description: Technical implementation notes, patterns, and code guidelines
---

# Implementation Guide

## Development Setup

Không cần thêm dependency mới. Cần:
- PostgreSQL running với database WC hiện tại
- `WC_JWT_SECRET` env var (đã có)
- Go 1.21+, Vue 3 + Element Plus (đã có)

---

## Code Structure

```
backend/internal/
├── model/
│   └── wc_match.go          ← Thêm WcChampionTeam, WcChampionConfig, WcChampionPrediction structs
├── repository/
│   └── wc_champion_repository.go   ← Mới
├── service/
│   └── wc_champion_service.go      ← Mới
└── api/
    ├── wc_champion_handler.go      ← Mới
    └── router.go                   ← Thêm routes

backend/internal/database/
└── database.go              ← Thêm AutoMigrate + seed

frontend/src/
├── components/wc/
│   └── WcChampionPanel.vue         ← Mới
├── views/
│   └── WcPredictView.vue           ← Thêm WcChampionPanel
├── views/
│   └── WcAdminView.vue (hoặc WcAdminPanel.vue) ← Thêm champion admin section
└── services/
    └── wcService.ts                ← Thêm champion API calls
```

---

## Implementation Notes

### Backend Models (thêm vào `wc_match.go`)

```go
type WcChampionTeam struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Name      string    `gorm:"not null;uniqueIndex"`
    Code      string    `gorm:"not null"`
    FlagEmoji string
    Odds      float64   `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type WcChampionConfig struct {
    ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    IsOpen    bool       `gorm:"not null;default:false"`
    WinnerID  *uuid.UUID `gorm:"type:uuid"`
    Winner    *WcChampionTeam `gorm:"foreignKey:WinnerID"`
    SettledAt *time.Time
    CreatedAt time.Time
    UpdatedAt time.Time
}

type WcChampionPrediction struct {
    ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    WcUserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
    TeamID       uuid.UUID `gorm:"type:uuid;not null"`
    Team         WcChampionTeam `gorm:"foreignKey:TeamID"`
    Points       int       `gorm:"not null"`
    OddsSnapshot float64   `gorm:"not null"`
    Result       *string   // nil | "correct" | "incorrect"
    PointsEarned *int
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### Service — Settle logic

```go
func (s *WcChampionService) SettleChampion(winnerTeamID uuid.UUID) (*SettleSummary, error) {
    cfg, _ := s.repo.GetConfig()
    if cfg.SettledAt != nil {
        return nil, fmt.Errorf("already settled")
    }

    return s.repo.SettleChampion(winnerTeamID) // DB transaction bên trong
}
```

Trong `repo.SettleChampion`:
1. Update `wc_champion_config`: set `winner_id`, `settled_at = NOW()`
2. Load tất cả predictions
3. Với mỗi prediction:
   - Nếu `team_id == winnerTeamID`: `points_earned = floor(points * odds_snapshot)`, `result = "correct"`, `wallet += points_earned - points`
   - Ngược lại: `points_earned = 0`, `result = "incorrect"`, `wallet -= points`
4. Gọi trong 1 DB transaction

### Frontend — WcChampionPanel.vue structure

```
WcChampionPanel
├── Status badge (Đang mở / Đã đóng / Đã có kết quả: [Đội])
├── [Nếu chưa đặt và đang mở]
│   ├── Bảng đội: flag | tên | odds | payout nếu thắng X điểm
│   └── Form: chọn đội (click row) + input điểm + button Đặt cược
├── [Nếu đã đặt]
│   ├── Card: "Bạn đang đặt [Flag] [Tên đội] — X điểm — x[odds] — Thắng: Y điểm"
│   └── [Nếu còn mở] Button Sửa / Hủy
└── [Nếu đã settle]
    ├── Kết quả: ✅ Đúng +Y điểm / ❌ Sai -X điểm
    └── Đội vô địch: [Flag] [Tên]
```

### API calls cần thêm vào `wcService.ts`

```typescript
getChampionTeams(): GET /wc/champion/teams
getChampionConfig(): GET /wc/champion/config
getMyChampionPrediction(): GET /wc/champion/my-prediction
placeChampionPrediction(teamId, points): POST /wc/champion/predict
deleteChampionPrediction(): DELETE /wc/champion/predict
// Admin:
updateChampionConfig(isOpen): PUT /wc/admin/champion/config
updateTeamOdds(teamId, odds): PUT /wc/admin/champion/teams/:id
settleChampion(winnerTeamId): POST /wc/admin/champion/settle
```

---

## Integration Points

- Route đặt trong `wcFeature` group (yêu cầu feature enabled) cho user endpoints
- Route admin đặt trong `wcAdminAlways` group (không cần feature flag)
- Wallet updates dùng `repo.UpdateWalletBalance(tx, wcUserID, delta)` — cùng function với match predictions
- Seed data chạy trong `database.go` cùng với `wc_config` seed hiện tại

---

## Error Handling

| Trường hợp | Response |
|-----------|----------|
| Đặt khi cửa sổ đóng | 400 `"champion predictions are closed"` |
| Points > 5 | 400 `"points must not exceed 5"` |
| Team không tồn tại | 404 `"team not found"` |
| Settle lần 2 | 409 `"already settled"` |
| Sửa sau khi đã settle | 400 `"prediction is already finalized"` |

Thêm friendly messages vào `wcApi.ts` ERROR_MAP:
```typescript
['champion predictions are closed',    'Cửa sổ dự đoán vô địch đã đóng'],
['already settled',                    'Kết quả vô địch đã được công bố'],
['prediction is already finalized',    'Dự đoán đã được chốt, không thể sửa'],
```

---

## Security Notes

- PlacePredict endpoint: yêu cầu JWT (`requiresWcAuth`)
- Admin endpoints: yêu cầu JWT + `isAdmin = true`
- `odds_snapshot` lấy từ DB tại thời điểm đặt — không nhận từ client
- Settle: kiểm tra `winnerTeamID` tồn tại trong `wc_champion_teams` trước khi xử lý
