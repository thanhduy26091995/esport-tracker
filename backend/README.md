# FC25 Esport Score Tracker - Backend

Go backend API for the FC25 Esport Score Tracker.

## Tech Stack

- **Language:** Go 1.21+
- **Framework:** Gin (HTTP router)
- **Database:** PostgreSQL 14+
- **ORM:** GORM

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher (or Docker)

## Setup

### 1. Database Setup

**Option A: Docker (Recommended)**
```bash
docker run --name esport-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=esport_tracker \
  -p 5432:5432 \
  -d postgres:14
```

**Option B: Local PostgreSQL**
```bash
createdb esport_tracker
```

### 2. Environment Configuration

Copy `.env` file and adjust if needed:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=esport_tracker
DB_SSLMODE=disable
PORT=8080
CORS_ORIGINS=http://localhost:5173
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run Server

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Health Check
- `GET /health` - Check if server is running

### API v1 (Base: `/api/v1`)
- Coming soon: User management, match recording, debt settlement, fund management

## Development

### Project Structure

```
backend/
├── cmd/server/         # Application entry point
├── internal/
│   ├── api/           # HTTP handlers and routing
│   ├── service/       # Business logic layer
│   ├── repository/    # Data access layer
│   ├── model/         # Database models
│   └── database/      # Database connection
├── migrations/        # Database migrations
├── .env              # Environment variables
└── go.mod            # Go module dependencies
```

### Run in Development Mode

```bash
# With auto-reload (install air)
go install github.com/air-verse/air@latest
air
```

## Database Models

- **users** - Player information and current scores
- **matches** - Match records (1v1, 2v2)
- **match_participants** - Players in each match
- **debt_settlements** - Debt payment records
- **fund_transactions** - Shared fund transactions
- **config** - System configuration

## Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Build

```bash
# Build for production
go build -o bin/server cmd/server/main.go

# Run binary
./bin/server
```

## License

Private project
