# Backend Patterns

## Dependency Injection Chain

Every request goes through: Repository → Service → Handler.  
All wiring happens in `router.go` — no global singletons.

```go
// Pattern in router.go
userRepo := repository.NewWcUserRepository(db)
authSvc  := service.NewWcAuthService(userRepo, jwtSecret, gcpClientID)
authHandler := api.NewWcAuthHandler(authSvc)
```

Never skip a layer — handlers must not call repositories directly.

## DB Models

- All models define `TableName() string` if the GORM default would differ
- WC models live in `model/wc_*.go` files
- UUID primary keys via `gorm:"type:uuid;default:gen_random_uuid()"`
- Use `*string` / `*time.Time` for nullable fields (pointer = nullable)

```go
type WcUser struct {
    ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Name         string     `gorm:"not null;uniqueIndex"`
    PasswordHash *string    // nil for Google-only accounts
    GoogleID     *string    `gorm:"uniqueIndex"`
    AvatarURL    *string
    IsAdmin      bool       `gorm:"default:false"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
func (WcUser) TableName() string { return "wc_users" }
```

## Repository Pattern

- One file per domain entity or tightly related group: `wc_repository.go`, `wc_user_repository.go`
- Use GORM's `.Where()` chaining for optional filters
- Upsert external data with `ON CONFLICT` clause:

```go
db.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "external_id"}},
    DoUpdates: clause.AssignmentColumns([]string{"status", "home_score", ...}),
}).Create(&matches)
```

## Error Handling

- Define sentinel errors at the service layer (`var ErrXxx = errors.New("...")`)
- Handlers switch on sentinel errors to set correct HTTP status codes
- Never leak DB errors directly to API responses

## Middleware Context Keys

Always use typed constants, not bare strings:

```go
const (
    WcUserIDKey   = "wc_user_id"
    WcUserNameKey = "wc_user_name"
    WcIsAdminKey  = "wc_is_admin"
)
// In handler:
userID := c.GetString(WcUserIDKey)
```

## Cron Jobs

Background jobs registered in `router.go` after route wiring:

```go
// Runs every 30 minutes
go service.StartWcSyncCron(wcRepo, statsApiSvc, 30*time.Minute)
```

## Decimal Math

Always use `shopspring/decimal` for money and points:

```go
amount := decimal.NewFromFloat(22000).Mul(decimal.NewFromInt(absScore))
```

Never use `float64` for financial calculations.

## Test Patterns

- Unit tests use injectable dependencies (e.g., `googleVerifier` function field in `WcAuthService`)
- Integration tests (`*_integration_test.go`) may use a real test DB
- Middleware tests (`wc_auth_test.go`) create a minimal Gin engine with the middleware under test
