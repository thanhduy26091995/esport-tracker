# FC25 Esport Score Tracker

Quick scoring system for FC25 matches with automatic debt settlement and shared fund management.

## 🚀 Quick Start

### Prerequisites

- **Backend:**
  - Go 1.21+
  - PostgreSQL 14+ (or Docker)

- **Frontend:**
  - Node.js 18+
  - npm or yarn

### 1. Start PostgreSQL

```bash
# Using Docker (recommended)
docker run --name esport-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=esport_tracker \
  -p 5432:5432 \
  -d postgres:14
```

### 2. Start Backend

```bash
cd backend
go run cmd/server/main.go
```

✅ Backend will be running on http://localhost:8080

### 3. Start Frontend

```bash
cd frontend
npm install  # first time only
npm run dev
```

✅ Frontend will be running on http://localhost:5173

## 📁 Project Structure

```
esport/
├── backend/              # Go/Gin REST API
│   ├── cmd/server/      # Application entry point
│   ├── internal/        # Internal packages
│   │   ├── api/        # HTTP handlers & routing
│   │   ├── service/    # Business logic
│   │   ├── repository/ # Data access layer
│   │   ├── model/      # Database models
│   │   └── database/   # DB connection
│   ├── migrations/      # Database migrations
│   └── .env            # Environment variables
│
├── frontend/            # Vue 3 + TypeScript
│   ├── src/
│   │   ├── components/ # Reusable Vue components
│   │   ├── views/      # Page components
│   │   ├── stores/     # Pinia state management
│   │   ├── services/   # API service layer
│   │   ├── types/      # TypeScript types
│   │   ├── utils/      # Utility functions
│   │   └── router/     # Vue Router configuration
│   └── .env            # Environment variables
│
└── docs/
    └── ai/             # Feature documentation
        ├── requirements/
        ├── design/
        ├── planning/
        └── implementation/
```

## ✨ Features

### Implemented (Phase 1 - Foundation)

✅ **Backend:**
- Go/Gin REST API with CORS support
- PostgreSQL database with GORM ORM
- Database models created (Users, Matches, Participants, Settlements, Fund, Config)
- Config seeding (debt threshold, point conversion, fund split)
- Health check endpoint

✅ **Frontend:**
- Vue 3 + TypeScript + Vite
- Vue Router with page navigation
- Pinia state management setup
- Element Plus UI library
- Tailwind CSS styling
- API service layer configured
- Basic dashboard and placeholder views

### Coming Next (Phase 2 - User & Match Management)

- ⏳ User CRUD operations (backend + frontend)
- ⏳ Leaderboard with real-time scores
- ⏳ Match recording (1v1 and 2v2)
- ⏳ Match history view

### Planned (Phase 3-4 - Debt & Fund)

- ⏳ Automatic debt settlement system
- ⏳ Fund balance tracking
- ⏳ Fund expense recording
- ⏳ Configuration management UI

## 🔧 Tech Stack

### Backend
- **Language:** Go 1.21+
- **Framework:** Gin (HTTP router)
- **Database:** PostgreSQL 14
- **ORM:** GORM
- **Libraries:**
  - CORS: gin-contrib/cors
  - Decimal: shopspring/decimal (for money calculations)
  - UUID: google/uuid
  - Env: joho/godotenv

### Frontend
- **Framework:** Vue 3 + TypeScript
- **Build Tool:** Vite
- **UI Library:** Element Plus
- **Styling:** Tailwind CSS
- **State Management:** Pinia
- **Server State:** TanStack Vue Query
- **Router:** Vue Router
- **HTTP Client:** Axios

### Database
- **Primary:** PostgreSQL 14
- **Hosting (prod):** Supabase (planned)

## 📊 Database Schema

**Core Tables:**
- `users` - Player information and current scores
- `matches` - Match records (1v1, 2v2)
- `match_participants` - Players in each match with point changes
- `debt_settlements` - Debt payment records
- `fund_transactions` - Shared fund income/expenses
- `config` - System configuration (debt threshold, point conversion)

## 🎯 Key Business Logic

### Scoring System
- Winner: +1 point per player
- Loser: -1 point per player
- Draws: 0 points (can be recorded)

### Debt Settlement
- **Trigger:** When player score ≤ configured threshold (default: -6)
- **Calculation:**
  1. Convert debt to VND: `abs(score) × conversion_rate` (e.g., 6 × 22,000 = 132,000 VND)
  2. Split 50/50:
     - 50% → Shared fund
     - 50% → Distributed to recent winning opponents
  3. Reset debtor's score to 0
  4. Deduct points from winners (money paid ÷ conversion rate)
  5. Lock related matches to prevent edits

### Fund Management
- **Income:** 50% of all debt settlements
- **Expenses:** Manual recording (e.g., buying gear, game disks)
- **Balance:** Real-time calculation from transaction history

## 🔗 API Endpoints

### Currently Available

```
GET  /health              - Health check
GET  /api/v1/users        - Placeholder (will list users)
```

### Coming Soon

```
# User Management
POST   /api/v1/users              - Create user
GET    /api/v1/users/:id          - Get user details
PUT    /api/v1/users/:id          - Update user
DELETE /api/v1/users/:id          - Delete user

# Match Management
POST   /api/v1/matches            - Create match & update scores
GET    /api/v1/matches            - List matches (paginated)
GET    /api/v1/matches/:id        - Get match details
PUT    /api/v1/matches/:id        - Update match (if not locked)
DELETE /api/v1/matches/:id        - Delete match & revert scores

# Debt Settlement
GET    /api/v1/settlements        - List all settlements
GET    /api/v1/settlements/:id    - Get settlement details

# Fund
GET    /api/v1/fund               - Get balance & recent transactions
POST   /api/v1/fund/expense       - Record expense

# Config
GET    /api/v1/config             - Get all config values
PUT    /api/v1/config/:key        - Update config value
```

## 🚧 Development

### Backend Development

```bash
cd backend

# Run with auto-reload (optional)
go install github.com/air-verse/air@latest
air

# Or run normally
go run cmd/server/main.go

# Run tests (when added)
go test ./...
```

### Frontend Development

```bash
cd frontend

# Start dev server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Type check
npm run type-check
```

### Environment Variables

**Backend (`.env` in `backend/`):**
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=esport_tracker
DB_SSLMODE=disable
PORT=8080
CORS_ORIGINS=http://localhost:5173
```

**Frontend (`.env` in `frontend/`):**
```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## 📝 Next Steps

1. ✅ **Phase 1 Complete:** Foundation setup (backend + frontend)
2. 🔄 **Phase 2 In Progress:** User & match management
   - Implement user CRUD operations
   - Build leaderboard component
   - Create match entry form
3. **Phase 3:** Debt settlement system
4. **Phase 4:** Fund management
5. **Phase 5:** Polish & deployment

## 📚 Documentation

Detailed documentation available in `docs/ai/`:

- [Requirements](docs/ai/requirements/feature-esport-score-tracker.md) - Problem statement, user stories, success criteria
- [Design](docs/ai/design/feature-esport-score-tracker.md) - Architecture, data models, API design
- [Planning](docs/ai/planning/feature-esport-score-tracker.md) - Task breakdown, timeline estimates
- [Implementation](docs/ai/implementation/feature-esport-score-tracker.md) - Code guidelines, patterns

## 🤝 Contributing

This is a personal project for managing FC25 matches. If you have suggestions:
1. Open an issue for discussion
2. Fork and create a pull request

## 📄 License

Private project - Not licensed for public use yet.

---

**Built with ❤️ for FC25 players**
