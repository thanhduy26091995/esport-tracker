---
phase: implementation
title: "External Bet Bonus Points — Implementation Guide"
description: File-by-file instructions for implementing the external bet bonus feature
---

# Implementation Guide

## New Files to Create

| File | Purpose |
|---|---|
| `backend/internal/model/score_bonus.go` | ScoreBonus struct |
| `backend/internal/repository/score_bonus_repository.go` | DB operations |
| `backend/internal/service/score_bonus_service.go` | Business logic |
| `backend/internal/api/score_bonus_handler.go` | HTTP handler |
| `frontend/src/types/scoreBonus.ts` | TS types |
| `frontend/src/services/scoreBonusService.ts` | API calls |
| `frontend/src/components/ScoreBonusForm.vue` | Dialog form |

## Files to Modify

| File | Change |
|---|---|
| `backend/internal/api/router.go` | Register routes + handler |
| `backend/internal/database/database.go` | AutoMigrate ScoreBonus |
| `frontend/src/views/FundView.vue` | Add "Cộng điểm cá cược" button + dialog |
| `frontend/src/locales/vi.json` + `en.json` | i18n keys |

## Implementation Notes

### `score_bonus.go` model

```go
package model

import (
    "time"
    "github.com/google/uuid"
)

type ScoreBonus struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
    Points      int       `gorm:"not null" json:"points"`
    FundAmount  int       `gorm:"not null" json:"fund_amount"`
    Description string    `gorm:"type:text" json:"description"`
    RecordedBy  string    `gorm:"type:varchar(100)" json:"recorded_by"`
    BonusDate   time.Time `gorm:"default:now()" json:"bonus_date"`
    CreatedAt   time.Time `json:"created_at"`
    User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ScoreBonus) TableName() string { return "score_bonuses" }
```

### `score_bonus_service.go` — `CreateBonus` transaction outline

```go
func (s *ScoreBonusService) CreateBonus(req *CreateScoreBonusRequest) (*model.ScoreBonus, error) {
    if req.Points <= 0 {
        return nil, errors.New("points must be positive")
    }
    user, err := s.userRepo.GetByID(req.UserID)
    if err != nil { return nil, fmt.Errorf("user with ID %s not found", req.UserID) }
    _ = user

    pointToVND, _ := s.configService.GetPointToVND()
    fundAmount := req.Points * pointToVND

    bonusDate := time.Now()
    if req.BonusDate != nil { bonusDate = *req.BonusDate }

    tx := s.db.Begin()
    defer func() { if r := recover(); r != nil { tx.Rollback() } }()

    bonus := &model.ScoreBonus{
        UserID:      req.UserID,
        Points:      req.Points,
        Description: req.Description,
        BonusDate:   bonusDate,
    }
    if err := tx.Create(bonus).Error; err != nil { tx.Rollback(); return nil, err }

    if err := tx.Model(&model.User{}).Where("id = ?", req.UserID).
        Update("current_score", gorm.Expr("current_score + ?", req.Points)).Error; err != nil {
        tx.Rollback(); return nil, err
    }

    if err := tx.Commit().Error; err != nil { return nil, err }

    _ = s.tierService.RecalculateForUsers([]uuid.UUID{req.UserID})

    return s.repo.GetByID(bonus.ID)
}
```

### `DeleteBonus` — revert pattern

1. Get bonus by ID
2. Begin tx
3. `UPDATE users SET current_score = current_score - bonus.Points WHERE id = bonus.UserID`
4. Delete bonus record
5. Commit, tier recalc

### Router registration (`router.go`)

```go
bonusRepo := repository.NewScoreBonusRepository(db)
bonusService := service.NewScoreBonusService(bonusRepo, userRepo, tierService, db)
bonusHandler := api.NewScoreBonusHandler(bonusService)

// matchHandler gets bonusService injected so GetAll can merge
matchHandler := api.NewMatchHandler(matchService, bonusService)

bonuses := v1.Group("/score-bonuses")
{
    bonuses.POST("", bonusHandler.Create)    // no GET here — reading via /matches
    bonuses.DELETE("/:id", bonusHandler.Delete)
}
```

### `MatchHandler.GetAll` merge logic (`match_handler.go`)

```go
type matchFeedItem struct {
    Type string `json:"type"`
    *model.Match
}
type bonusFeedItem struct {
    Type string `json:"type"`
    *model.ScoreBonus
}

func (h *MatchHandler) GetAll(c *gin.Context) {
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

    matches, _ := h.matchService.GetAllMatches(limit, offset)
    bonuses, _ := h.bonusService.GetAll(0, 0) // fetch all, merge before paginating

    items := make([]interface{}, 0, len(matches)+len(bonuses))
    for _, m := range matches {
        items = append(items, matchFeedItem{Type: "match", Match: m})
    }
    for _, b := range bonuses {
        items = append(items, bonusFeedItem{Type: "bonus", ScoreBonus: b})
    }

    sort.Slice(items, func(i, j int) bool {
        return getItemDate(items[i]).After(getItemDate(items[j]))
    })

    // Apply pagination after merge
    start := offset
    if start > len(items) { start = len(items) }
    end := start + limit
    if end > len(items) { end = len(items) }

    c.JSON(http.StatusOK, items[start:end])
}
```

> Note: `getItemDate` is a small helper that extracts `MatchDate` or `BonusDate` depending on item type.

### AutoMigrate (`database.go`)

Add `&model.ScoreBonus{}` to the existing `AutoMigrate(...)` call.

### Frontend — `ScoreBonusForm.vue` key props/emits

```ts
defineProps<{ modelValue: boolean; loading: boolean }>()
defineEmits<{ (e: 'update:modelValue', v: boolean): void; (e: 'submit', req: CreateScoreBonusRequest): void }>()
```

### Frontend — `FundView.vue` addition

Add button next to Deposit/Withdraw buttons:
```html
<el-button type="warning" plain @click="showBonusForm = true" :icon="Trophy">
  {{ t('fund.addBetBonus') }}
</el-button>
```

Wire `ScoreBonusForm` and on submit: call `scoreBonusService.create(req)` then `fundStore.fetchAll()`.

## Integration Points

- `tierService.RecalculateForUsers` → called post-commit (non-fatal)
- `userStore` / leaderboard in frontend → refresh after bonus creation to reflect updated score
- `matchStore` in frontend → refresh after bonus create/delete so MatchesView stays in sync (existing refresh mechanism)
