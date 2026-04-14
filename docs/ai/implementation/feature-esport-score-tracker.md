---
phase: implementation
title: FC25 Esport Score Tracker - Implementation Guide
description: Technical implementation patterns and code guidelines
feature: esport-score-tracker
created: 2026-04-14
---

# Implementation Guide

## Development Setup

### Prerequisites
- **Go:** 1.21 or higher
- **Node.js:** 18 or higher
- **PostgreSQL:** 14 or higher (or Docker)
- **Git:** For version control
- **IDE:** VS Code recommended (with Go, Vue, and Tailwind extensions)

### Environment Setup Steps

**1. Clone and Initialize Repository**
```bash
mkdir esport-score-tracker
cd esport-score-tracker
git init
```

**2. Backend Setup**
```bash
# Create Go module
mkdir backend
cd backend
go mod init github.com/yourusername/esport-score-tracker

# Install dependencies
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/joho/godotenv
go get github.com/golang-migrate/migrate/v4

# Create .env file
cat > .env << EOF
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=esport_tracker
DB_SSLMODE=disable
PORT=8080
CORS_ORIGINS=http://localhost:5173
EOF
```

**3. Database Setup**
```bash
# Option 1: Local PostgreSQL
createdb esport_tracker

# Option 2: Docker PostgreSQL
docker run --name esport-postgres \
  -e POSTGRES_PASSWORD=yourpassword \
  -e POSTGRES_DB=esport_tracker \
  -p 5432:5432 \
  -d postgres:14
```

**4. Frontend Setup**
```bash
# From project root
npm create vite@latest frontend -- --template vue-ts
cd frontend

# Install dependencies
npm install
npm install vue-router pinia axios @tanstack/vue-query
npm install element-plus @element-plus/icons-vue
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p

# Create .env file
cat > .env << EOF
VITE_API_BASE_URL=http://localhost:8080/api/v1
EOF
```

**5. Run Development Servers**
```bash
# Terminal 1 - Backend
cd backend
go run cmd/server/main.go

# Terminal 2 - Frontend
cd frontend
npm run dev
```

### Configuration Files

**backend/.gitignore**
```
.env
bin/
*.exe
*.log
```

**frontend/.gitignore**
```
node_modules/
dist/
.env.local
.env.production.local
```

**frontend/tailwind.config.js**
```js
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

---

## Code Structure

### Backend Directory Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
├── internal/
│   ├── api/
│   │   ├── router.go               # Route definitions
│   │   ├── middleware.go           # CORS, logging, error handling
│   │   ├── user_handler.go         # User HTTP handlers
│   │   ├── match_handler.go        # Match HTTP handlers
│   │   ├── settlement_handler.go   # Settlement HTTP handlers
│   │   ├── fund_handler.go         # Fund HTTP handlers
│   │   └── config_handler.go       # Config HTTP handlers
│   ├── service/
│   │   ├── user_service.go         # User business logic
│   │   ├── match_service.go        # Match + scoring logic
│   │   ├── settlement_service.go   # Debt settlement calculations
│   │   ├── fund_service.go         # Fund balance calculations
│   │   └── config_service.go       # Config management
│   ├── repository/
│   │   ├── user_repository.go      # User DB operations
│   │   ├── match_repository.go     # Match DB operations
│   │   ├── settlement_repository.go
│   │   ├── fund_repository.go
│   │   └── config_repository.go
│   ├── model/
│   │   ├── user.go                 # User model
│   │   ├── match.go                # Match models
│   │   ├── settlement.go           # Settlement model
│   │   ├── fund.go                 # Fund transaction model
│   │   └── config.go               # Config model
│   └── database/
│       └── database.go             # DB connection setup
├── migrations/
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── go.mod
├── go.sum
└── .env
```

### Frontend Directory Structure

```
frontend/
├── src/
│   ├── assets/
│   │   └── main.css               # Tailwind imports
│   ├── components/
│   │   ├── layout/
│   │   │   ├── AppLayout.vue      # Main layout with nav
│   │   │   └── AppNavbar.vue      # Navigation bar
│   │   ├── user/
│   │   │   ├── UserTable.vue      # Leaderboard table
│   │   │   └── UserForm.vue       # Add/Edit user form
│   │   ├── match/
│   │   │   ├── MatchForm.vue      # Match entry form
│   │   │   ├── MatchList.vue      # Match history list
│   │   │   └── MatchCard.vue      # Single match display
│   │   ├── settlement/
│   │   │   └── SettlementModal.vue
│   │   ├── fund/
│   │   │   └── FundBalance.vue    # Fund balance widget
│   │   ├── common/
│   │   │   ├── ScoreDisplay.vue   # Score with VND
│   │   │   └── DebtBadge.vue      # Debt warning indicator
│   ├── views/
│   │   ├── DashboardView.vue      # Main dashboard
│   │   ├── UsersView.vue          # User management
│   │   ├── MatchesView.vue        # Match history
│   │   ├── CreateMatchView.vue    # Match creation
│   │   ├── FundView.vue           # Fund management
│   │   └── SettingsView.vue       # Configuration
│   ├── stores/
│   │   ├── userStore.ts           # User state (Pinia)
│   │   ├── matchStore.ts          # Match state
│   │   ├── settlementStore.ts     # Settlement state
│   │   ├── fundStore.ts           # Fund state
│   │   └── configStore.ts         # Config state
│   ├── services/
│   │   ├── api.ts                 # Axios instance
│   │   ├── userService.ts         # User API calls
│   │   ├── matchService.ts        # Match API calls
│   │   ├── settlementService.ts   # Settlement API calls
│   │   ├── fundService.ts         # Fund API calls
│   │   └── configService.ts       # Config API calls
│   ├── types/
│   │   ├── user.ts                # User types
│   │   ├── match.ts               # Match types
│   │   ├── settlement.ts          # Settlement types
│   │   ├── fund.ts                # Fund types
│   │   └── config.ts              # Config types
│   ├── utils/
│   │   ├── currency.ts            # VND formatting
│   │   └── date.ts                # Date formatting
│   ├── router/
│   │   └── index.ts               # Route definitions
│   ├── App.vue                    # Root component
│   └── main.ts                    # Entry point
├── public/
├── index.html
├── vite.config.ts
├── tailwind.config.js
├── tsconfig.json
├── package.json
└── .env
```

### Naming Conventions

**Backend (Go):**
- Files: `snake_case.go`
- Types/Structs: `PascalCase`
- Functions: `PascalCase` (exported), `camelCase` (private)
- Variables: `camelCase`
- Constants: `PascalCase` or `SCREAMING_SNAKE_CASE`
- DB tables: `snake_case` (plural)

**Frontend (TypeScript/Vue):**
- Files: `PascalCase.vue` (components), `camelCase.ts` (utilities)
- Components: `PascalCase`
- Functions: `camelCase`
- Variables: `camelCase`
- Constants: `SCREAMING_SNAKE_CASE`
- CSS classes: `kebab-case` (Tailwind)

---

## Implementation Notes

### Core Features

#### Feature 1: User Management

**Backend Model (model/user.go):**
```go
type User struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Name         string    `gorm:"type:varchar(100);unique;not null"`
    CurrentScore int       `gorm:"default:0"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    IsActive     bool      `gorm:"default:true"`
}
```

**Frontend Type (types/user.ts):**
```typescript
export interface User {
  id: string
  name: string
  current_score: number
  created_at: string
  updated_at: string
  is_active: boolean
}

export interface CreateUserRequest {
  name: string
}

export interface UpdateUserRequest {
  name: string
}
```

**Key Implementation Details:**
- Use soft delete (`is_active = false`) instead of hard delete
- Validate unique names in service layer before DB insert
- Index on `current_score DESC` for fast leaderboard queries

#### Feature 2: Match Recording

**Backend Models (model/match.go):**
```go
type Match struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    MatchType   string    `gorm:"type:varchar(10);not null"` // "1v1" or "2v2"
    WinnerTeam  int       `gorm:"not null"` // 1 or 2
    MatchDate   time.Time `gorm:"default:now()"`
    RecordedBy  string    `gorm:"type:varchar(100)"`
    CreatedAt   time.Time
    IsLocked    bool      `gorm:"default:false"`
    Participants []MatchParticipant `gorm:"foreignKey:MatchID"`
}

type MatchParticipant struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    MatchID     uuid.UUID `gorm:"type:uuid;not null"`
    UserID      uuid.UUID `gorm:"type:uuid;not null"`
    TeamNumber  int       `gorm:"not null"` // 1 or 2
    PointChange int       `gorm:"not null"` // +1 or -1
    User        User      `gorm:"foreignKey:UserID"`
}
```

**Scoring Logic (service/match_service.go):**
```go
func (s *MatchService) CreateMatch(req CreateMatchRequest) (*MatchResponse, error) {
    // Start transaction
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 1. Create match record
    match := &model.Match{
        MatchType:  req.MatchType,
        WinnerTeam: req.WinnerTeam,
        MatchDate:  req.MatchDate,
        RecordedBy: req.RecordedBy,
    }
    if err := tx.Create(match).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    // 2. Create participants and update scores
    var scoreUpdates []ScoreUpdate
    teams := map[int][]uuid.UUID{1: req.Team1, 2: req.Team2}
    
    for teamNum, userIDs := range teams {
        pointChange := -1
        if teamNum == req.WinnerTeam {
            pointChange = 1
        }
        
        for _, userID := range userIDs {
            // Create participant
            participant := &model.MatchParticipant{
                MatchID:     match.ID,
                UserID:      userID,
                TeamNumber:  teamNum,
                PointChange: pointChange,
            }
            if err := tx.Create(participant).Error; err != nil {
                tx.Rollback()
                return nil, err
            }
            
            // Update user score
            if err := tx.Model(&model.User{}).
                Where("id = ?", userID).
                UpdateColumn("current_score", gorm.Expr("current_score + ?", pointChange)).
                Error; err != nil {
                tx.Rollback()
                return nil, err
            }
            
            // Track score change for response
            var user model.User
            tx.First(&user, "id = ?", userID)
            scoreUpdates = append(scoreUpdates, ScoreUpdate{
                UserID:   userID,
                UserName: user.Name,
                OldScore: user.CurrentScore - pointChange,
                NewScore: user.CurrentScore,
                Change:   pointChange,
            })
        }
    }

    // 3. Check for debt settlements
    settlements, err := s.settlementService.CheckAndTriggerSettlements(tx, req.Team1, req.Team2)
    if err != nil {
        tx.Rollback()
        return nil, err
    }

    // 4. Commit transaction
    if err := tx.Commit().Error; err != nil {
        return nil, err
    }

    return &MatchResponse{
        Match:              match,
        ScoreUpdates:       scoreUpdates,
        SettlementsTriggered: settlements,
    }, nil
}
```

**Key Implementation Details:**
- **Atomicity:** Use database transaction for match + score updates + settlements
- **Point Calculation:** Winner team gets +1, loser team gets -1 per player
- **Validation:** Ensure team sizes match match type (1v1: 1 player/team, 2v2: 2 players/team)
- **Locking:** Set `is_locked = true` after debt settlement to prevent edits

#### Feature 3: Debt Settlement

**Backend Model (model/settlement.go):**
```go
type DebtSettlement struct {
    ID                 uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    DebtorID           uuid.UUID       `gorm:"type:uuid;not null"`
    DebtAmount         int             `gorm:"not null"` // Negative value (e.g., -6)
    MoneyAmount        decimal.Decimal `gorm:"type:numeric(12,2);not null"`
    ToFund             decimal.Decimal `gorm:"type:numeric(12,2);not null"`
    ToWinners          decimal.Decimal `gorm:"type:numeric(12,2);not null"`
    WinnerDistribution datatypes.JSON  `gorm:"type:jsonb"`
    SettledAt          time.Time       `gorm:"default:now()"`
    Debtor             User            `gorm:"foreignKey:DebtorID"`
}
```

**Settlement Logic (service/settlement_service.go):**
```go
func (s *SettlementService) CheckAndTriggerSettlements(tx *gorm.DB, team1, team2 []uuid.UUID) ([]*DebtSettlement, error) {
    var settlements []*DebtSettlement
    
    // Get config values
    debtThreshold := s.getDebtThreshold() // e.g., -6
    pointToVND := s.getPointToVND()       // e.g., 22000
    
    // Check all participants
    allPlayers := append(team1, team2...)
    for _, userID := range allPlayers {
        var user model.User
        if err := tx.First(&user, "id = ?", userID).Error; err != nil {
            return nil, err
        }
        
        // Trigger settlement if at or below threshold
        if user.CurrentScore <= debtThreshold {
            settlement, err := s.settleDebt(tx, &user, pointToVND)
            if err != nil {
                return nil, err
            }
            settlements = append(settlements, settlement)
        }
    }
    
    return settlements, nil
}

func (s *SettlementService) settleDebt(tx *gorm.DB, debtor *model.User, pointToVND int) (*DebtSettlement, error) {
    debtAmount := debtor.CurrentScore // e.g., -6
    moneyAmount := decimal.NewFromInt(int64(abs(debtAmount) * pointToVND)) // 6 * 22000 = 132000 VND
    
    // Calculate 50/50 split
    toFund := moneyAmount.Div(decimal.NewFromInt(2)) // 66000 VND
    toWinners := moneyAmount.Div(decimal.NewFromInt(2)) // 66000 VND
    
    // Find recent opponents (winners who caused the debt)
    winners := s.findRecentOpponents(tx, debtor.ID, abs(debtAmount))
    
    // Distribute to winners evenly
    winnerDistribution := s.distributeToWinners(tx, winners, toWinners)
    
    // Reset debtor's score to 0
    newScore := 0
    if err := tx.Model(debtor).Update("current_score", newScore).Error; err != nil {
        return nil, err
    }
    
    // Create settlement record
    settlement := &model.DebtSettlement{
        DebtorID:           debtor.ID,
        DebtAmount:         debtAmount,
        MoneyAmount:        moneyAmount,
        ToFund:             toFund,
        ToWinners:          toWinners,
        WinnerDistribution: winnerDistribution,
    }
    if err := tx.Create(settlement).Error; err != nil {
        return nil, err
    }
    
    // Create fund transaction
    fundTx := &model.FundTransaction{
        Amount:              toFund,
        TransactionType:     "debt_in",
        Description:         fmt.Sprintf("Debt settlement from %s", debtor.Name),
        RelatedSettlementID: &settlement.ID,
    }
    if err := tx.Create(fundTx).Error; err != nil {
        return nil, err
    }
    
    // Lock related matches (last N matches involving debtor)
    s.lockRelatedMatches(tx, debtor.ID)
    
    return settlement, nil
}

func (s *SettlementService) distributeToWinners(tx *gorm.DB, winners []uuid.UUID, totalAmount decimal.Decimal) datatypes.JSON {
    if len(winners) == 0 {
        return datatypes.JSON([]byte("[]"))
    }
    
    perWinner := totalAmount.Div(decimal.NewFromInt(int64(len(winners))))
    distribution := make(map[string]interface{})
    
    for _, winnerID := range winners {
        var winner model.User
        tx.First(&winner, "id = ?", winnerID)
        
        // Calculate points to deduct from winner (reverse conversion)
        pointsToDeduct := int(perWinner.Div(decimal.NewFromInt(int64(s.getPointToVND()))).IntPart())
        
        // Update winner's score (reduce by points paid out)
        tx.Model(&winner).UpdateColumn("current_score", gorm.Expr("current_score - ?", pointsToDeduct))
        
        distribution[winner.Name] = map[string]interface{}{
            "user_id":        winnerID,
            "amount_vnd":     perWinner.String(),
            "points_deducted": pointsToDeduct,
        }
    }
    
    jsonData, _ := json.Marshal(distribution)
    return datatypes.JSON(jsonData)
}
```

**Key Implementation Details:**
- **Threshold Check:** Check all participants after each match
- **50/50 Split:** Use `decimal` library to avoid floating-point errors
- **Winner Identification:** Find recent opponents from match history
- **Score Adjustments:**
  - Debtor: Reset to 0 (add abs(debt) to current score)
  - Winners: Reduce by (money received / point_to_vnd)
- **Match Locking:** Lock matches involving debtor to prevent score manipulation

#### Feature 4: Fund Management

**Backend Model (model/fund.go):**
```go
type FundTransaction struct {
    ID                  uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Amount              decimal.Decimal `gorm:"type:numeric(12,2);not null"`
    TransactionType     string          `gorm:"type:varchar(20);not null"` // "debt_in" or "expense_out"
    Description         string          `gorm:"type:text"`
    RelatedSettlementID *uuid.UUID      `gorm:"type:uuid"`
    CreatedAt           time.Time       `gorm:"default:now()"`
}
```

**Fund Balance Calculation:**
```go
func (s *FundService) GetBalance() (decimal.Decimal, error) {
    var balance decimal.Decimal
    
    // SUM(amount) WHERE type = 'debt_in'
    var income decimal.Decimal
    s.db.Model(&model.FundTransaction{}).
        Where("transaction_type = ?", "debt_in").
        Select("COALESCE(SUM(amount), 0)").
        Scan(&income)
    
    // SUM(amount) WHERE type = 'expense_out'
    var expenses decimal.Decimal
    s.db.Model(&model.FundTransaction{}).
        Where("transaction_type = ?", "expense_out").
        Select("COALESCE(SUM(amount), 0)").
        Scan(&expenses)
    
    balance = income.Sub(expenses)
    return balance, nil
}
```

**Key Implementation Details:**
- **Automatic Income:** Fund transaction created automatically during debt settlement
- **Manual Expenses:** Admin can record expenses (e.g., "Bought new controller")
- **Balance Calculation:** Real-time calculation from transaction history
- **Negative Balance Allowed:** Shows warning but doesn't block operations

---

### Patterns & Best Practices

#### Design Patterns

**1. Repository Pattern**
```go
// Separate data access from business logic
type UserRepository interface {
    GetAll() ([]*model.User, error)
    GetByID(id uuid.UUID) (*model.User, error)
    Create(user *model.User) error
    Update(user *model.User) error
    SoftDelete(id uuid.UUID) error
}
```

**2. Service Layer Pattern**
```go
// Business logic separate from HTTP concerns
type UserService struct {
    repo UserRepository
}

func (s *UserService) CreateUser(name string) (*model.User, error) {
    // Validation
    if err := s.validateName(name); err != nil {
        return nil, err
    }
    
    // Business logic
    user := &model.User{Name: name}
    if err := s.repo.Create(user); err != nil {
        return nil, err
    }
    
    return user, nil
}
```

**3. DTO Pattern (Frontend)**
```typescript
// Separate API types from UI types
export interface UserDTO {
  id: string
  name: string
  current_score: number
}

export interface UserViewModel extends UserDTO {
  scoreInVND: number
  isInDebt: boolean
  rank: number
}
```

#### Code Style Guidelines

**Backend (Go):**
- Use `gofmt` for formatting
- Error handling: Always check errors, wrap with context
  ```go
  if err != nil {
      return nil, fmt.Errorf("failed to create user: %w", err)
  }
  ```
- Use structured logging (e.g., `zerolog` or `logrus`)
- Keep functions small (<50 lines ideally)
- Document exported functions and types

**Frontend (Vue 3):**
- Use Composition API (not Options API)
- Extract reusable logic into composables
  ```typescript
  // composables/useScore.ts
  export function useScore() {
    const config = useConfigStore()
    const toVND = (points: number) => points * config.pointToVND
    return { toVND }
  }
  ```
- Use `<script setup>` syntax for components
- Keep components focused (single responsibility)
- Use TypeScript strictly (no `any` types)

#### Common Utilities/Helpers

**Frontend Currency Formatting (utils/currency.ts):**
```typescript
export function formatVND(amount: number): string {
  return new Intl.NumberFormat('vi-VN', {
    style: 'currency',
    currency: 'VND',
    minimumFractionDigits: 0,
  }).format(amount)
}

export function pointsToVND(points: number, conversionRate: number): number {
  return points * conversionRate
}
```

**Frontend Date Formatting (utils/date.ts):**
```typescript
export function formatDateTime(date: string | Date): string {
  return new Intl.DateTimeFormat('vi-VN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(date))
}
```

**Backend Decimal Helper:**
```go
func abs(n int) int {
    if n < 0 {
        return -n
    }
    return n
}
```

---

## Integration Points

### API Integration

**Axios Configuration (services/api.ts):**
```typescript
import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000,
})

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      // Server error
      const message = error.response.data?.error?.message || 'An error occurred'
      ElMessage.error(message)
    } else if (error.request) {
      // Network error
      ElMessage.error('Network error. Please check your connection.')
    }
    return Promise.reject(error)
  }
)
```

**Service Example (services/userService.ts):**
```typescript
import { api } from './api'
import type { User, CreateUserRequest, UpdateUserRequest } from '@/types/user'

export const userService = {
  async getAll(): Promise<User[]> {
    const response = await api.get<User[]>('/users')
    return response.data
  },

  async create(data: CreateUserRequest): Promise<User> {
    const response = await api.post<User>('/users', data)
    return response.data
  },

  async update(id: string, data: UpdateUserRequest): Promise<User> {
    const response = await api.put<User>(`/users/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/users/${id}`)
  },
}
```

### Database Connection

**Backend Database Setup (internal/database/database.go):**
```go
package database

import (
    "esport-score-tracker/internal/model"
    "fmt"
    "log"
    "os"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func Connect() (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"),
        os.Getenv("DB_SSLMODE"),
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Auto-migrate models (development only, use migrations in production)
    if err := db.AutoMigrate(
        &model.User{},
        &model.Match{},
        &model.MatchParticipant{},
        &model.DebtSettlement{},
        &model.FundTransaction{},
        &model.Config{},
    ); err != nil {
        return nil, fmt.Errorf("failed to migrate database: %w", err)
    }

    log.Println("Database connected successfully")
    return db, nil
}
```

---

## Error Handling

### Backend Error Handling Strategy

**Custom Error Types:**
```go
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
    return e.Message
}

var (
    ErrValidation = &AppError{Code: "VALIDATION_ERROR", Message: "Validation failed"}
    ErrNotFound   = &AppError{Code: "NOT_FOUND", Message: "Resource not found"}
    ErrConflict   = &AppError{Code: "CONFLICT", Message: "Resource conflict"}
    ErrInternal   = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
)
```

**Error Middleware:**
```go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            var appErr *AppError
            if errors.As(err, &appErr) {
                c.JSON(getStatusCode(appErr.Code), gin.H{"error": appErr})
                return
            }

            // Default error
            c.JSON(500, gin.H{"error": ErrInternal})
        }
    }
}
```

### Frontend Error Handling

**Global Error Handler (in api.ts interceptor - shown above)**

**Component-Level Error Handling:**
```typescript
<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const error = ref<string | null>(null)

async function createUser(name: string) {
  loading.value = true
  error.value = null
  
  try {
    await userService.create({ name })
    ElMessage.success('User created successfully')
  } catch (err: any) {
    error.value = err.message || 'Failed to create user'
  } finally {
    loading.value = false
  }
}
</script>
```

### Logging Approach

**Backend:**
```go
import "github.com/gin-gonic/gin"

func LoggerMiddleware() gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        return fmt.Sprintf("[%s] %s %s %d %s\n",
            param.TimeStamp.Format("2006-01-02 15:04:05"),
            param.Method,
            param.Path,
            param.StatusCode,
            param.Latency,
        )
    })
}
```

**Frontend:**
```typescript
// Use console.error for errors, console.warn for warnings
console.error('Failed to load users:', error)
```

---

## Performance Considerations

### Optimization Strategies

**1. Database Indexes (Already in schema)**
- `idx_users_score` on `users(current_score DESC)` - Fast leaderboard
- `idx_matches_date` on `matches(match_date DESC)` - Fast history
- `idx_participants_match` and `idx_participants_user` - Fast joins

**2. Pagination**
```go
// Backend
func (r *MatchRepository) GetAll(page, pageSize int) ([]*model.Match, error) {
    var matches []*model.Match
    offset := (page - 1) * pageSize
    
    err := r.db.
        Preload("Participants.User").
        Order("match_date DESC").
        Limit(pageSize).
        Offset(offset).
        Find(&matches).Error
    
    return matches, err
}
```

```typescript
// Frontend
const currentPage = ref(1)
const pageSize = ref(50)

async function loadMatches() {
  const matches = await matchService.getAll(currentPage.value, pageSize.value)
}
```

**3. Eager Loading (Avoid N+1 queries)**
```go
// Load relationships in single query
db.Preload("Participants.User").Find(&matches)
```

### Caching Approach (Future Enhancement)

**Redis Caching for Leaderboard:**
```go
// Cache leaderboard for 30 seconds
func (s *UserService) GetLeaderboard() ([]*model.User, error) {
    cacheKey := "leaderboard"
    
    // Try cache first
    if cached := s.cache.Get(cacheKey); cached != nil {
        return cached.([]*model.User), nil
    }
    
    // Load from DB
    users, err := s.repo.GetAllOrderedByScore()
    if err != nil {
        return nil, err
    }
    
    // Cache for 30 seconds
    s.cache.Set(cacheKey, users, 30*time.Second)
    
    return users, nil
}
```

### Query Optimization

- Use `SELECT` specific columns instead of `SELECT *` when possible
- Use `COUNT(*)` for pagination total count
- Avoid loading large JSONB fields unless needed

### Resource Management

**Connection Pooling:**
```go
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

---

## Security Notes

### Input Validation

**Backend:**
```go
import "github.com/go-playground/validator/v10"

type CreateUserRequest struct {
    Name string `json:"name" binding:"required,min=2,max=100"`
}

func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // ... proceed with creation
}
```

**Frontend:**
```typescript
// Element Plus form validation
const rules = {
  name: [
    { required: true, message: 'Please enter name', trigger: 'blur' },
    { min: 2, max: 100, message: 'Length should be 2-100 characters', trigger: 'blur' },
  ],
}
```

### SQL Injection Prevention

- **Use ORM (GORM)** - automatically parameterizes queries
- **Never** use string concatenation for SQL

### XSS Prevention

- **Vue automatic escaping** - templates escape HTML by default
- Use `v-html` only with trusted content

### CORS Configuration

**Backend (middleware.go):**
```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        allowedOrigins := os.Getenv("CORS_ORIGINS") // "http://localhost:5173"
        
        c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

### Secrets Management

**Development:**
- Use `.env` files (never commit)
- Use `godotenv` (Go) and `vite` (frontend) to load

**Production:**
- Use environment variables (Railway/Vercel provide UI)
- Rotate database passwords periodically

---

## Production Deployment Notes

### Environment Variables

**Backend (.env.production):**
```
DB_HOST=your-supabase-host.supabase.co
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=***
DB_NAME=postgres
DB_SSLMODE=require
PORT=8080
CORS_ORIGINS=https://your-app.vercel.app
```

**Frontend (.env.production):**
```
VITE_API_BASE_URL=https://your-backend.railway.app/api/v1
```

### Build Commands

**Backend (Railway):**
```bash
go build -o bin/server cmd/server/main.go
```

**Frontend (Vercel):**
```json
{
  "buildCommand": "npm run build",
  "outputDirectory": "dist"
}
```

### Health Checks

**Backend:**
```go
router.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
})
```

### Monitoring & Logging

**Backend:**
- Log all errors with context
- Use structured logging (JSON format in production)

**Frontend:**
- Use Sentry for error tracking (optional)
- Log critical errors to console

---

## Testing Strategy (Future)

### Unit Tests

**Backend:**
```go
func TestUserService_CreateUser(t *testing.T) {
    // Setup
    db := setupTestDB()
    defer db.Close()
    
    service := NewUserService(db)
    
    // Test
    user, err := service.CreateUser("Test User")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Test User", user.Name)
    assert.Equal(t, 0, user.CurrentScore)
}
```

**Frontend:**
```typescript
import { describe, it, expect } from 'vitest'
import { pointsToVND } from '@/utils/currency'

describe('Currency Utils', () => {
  it('converts points to VND correctly', () => {
    expect(pointsToVND(5, 22000)).toBe(110000)
  })
})
```

### Integration Tests

- Test API endpoints with Postman collections
- Test database transactions
- Test debt settlement flow end-to-end

### Manual Testing Checklist

- [ ] Create user with unique name
- [ ] Create user with duplicate name (should fail)
- [ ] Record 1v1 match and verify scores
- [ ] Record 2v2 match and verify scores
- [ ] Trigger debt settlement at -6
- [ ] Verify fund balance updates
- [ ] Record fund expense
- [ ] Update config values
- [ ] View leaderboard on mobile
- [ ] Delete match and verify score reversion

---

## Next Steps After Implementation

1. **Code Review** - Review code for best practices
2. **Security Audit** - Check for common vulnerabilities
3. **Performance Testing** - Load test with expected user count
4. **User Acceptance Testing** - Get feedback from actual users
5. **Backup Strategy** - Set up automated database backups
6. **Monitoring** - Add error tracking and uptime monitoring
7. **Documentation** - Update README with deployment steps
8. **Knowledge Transfer** - Train admins on how to use the system
