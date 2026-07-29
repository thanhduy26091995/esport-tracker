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

## Caching

All services receive a `cache.CacheStore` injected via their constructor. Two implementations exist:

- **`RedisCache`** — production backend (`redis/go-redis/v9`)
- **`GoCacheStore`** — dev/test fallback (`patrickmn/go-cache`), activated when `REDIS_URL` is unset

Soft startup in `router.go`: if `REDIS_URL` is set but Redis is unreachable, the server logs a warning and falls back to `GoCacheStore` rather than crashing.

### Cache-Aside pattern

Use `cache.GetOrFetch[T]` for read-through caching with stampede prevention:

```go
func (s *UserService) GetLeaderboard() ([]*model.User, error) {
    return cache.GetOrFetch(s.cache, &s.group, "esport:users:leaderboard", 5*time.Minute, func() ([]*model.User, error) {
        return s.repo.GetLeaderboard()
    })
}
```

`GetOrFetch` uses `singleflight.Group` so concurrent cache misses on the same key result in exactly one DB query — all callers share the result.

### Write-invalidate

Every mutating service method calls `s.cache.Delete(key)` after a successful DB commit. Never delete before the commit (breaks atomicity).

```go
func (s *UserService) CreateUser(req *CreateUserRequest) (*model.User, error) {
    user, err := s.repo.Create(...)
    if err != nil { return nil, err }
    s.invalidateUserCaches()   // delete after successful write
    return user, nil
}
```

### Key naming conventions

| Prefix | Scope | Example |
|--------|-------|---------|
| `esport:users:*` | Core user scores | `esport:users:leaderboard` |
| `esport:matches:*` | Core matches (versioned) | `esport:matches:v3:20:0:` |
| `esport:config` | App config | — |
| `esport:fund:totals` | Fund balance | — |
| `wc:leaderboard:{tt}` | WC leaderboard per tournament | `wc:leaderboard:world_cup` |
| `wc:matches:all:{tt}` | WC match list per tournament | `wc:matches:all:asean_cup` |
| `wc:config:{tt}` | WC betting config per tournament | — |

WC keys are scoped by `tournament_type` so ASEAN Cup and World Cup caches are independent.

### Versioned match keys

Match list pages use a version counter instead of pattern-delete:

```go
version, _ := s.cache.GetInt("esport:matches:version")
key := fmt.Sprintf("esport:matches:v%d:%d:%d:%s", version, limit, offset, pid)
```

On any match write, `s.cache.Incr("esport:matches:version")` makes all old page keys unreachable orphans that expire via TTL — no `SCAN` needed.

### NoopCache

Tests that don't exercise caching logic should pass `&cache.NoopCache{}` to service constructors to avoid importing miniredis.

## Test Patterns

- Unit tests use injectable dependencies (e.g., `googleVerifier` function field in `WcAuthService`)
- Integration tests (`*_integration_test.go`) may use a real test DB
- Middleware tests (`wc_auth_test.go`) create a minimal Gin engine with the middleware under test
