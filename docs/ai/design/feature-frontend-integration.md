---
feature: frontend-integration
phase: design
status: draft
created: 2026-04-14
---

# Frontend Integration Design

## Architecture Overview

### System Architecture
```
┌─────────────────────────────────────────────────────────┐
│                    Browser (SPA)                        │
│  ┌───────────────────────────────────────────────────┐  │
│  │           Vue 3 Application                       │  │
│  │  ┌─────────────────────────────────────────────┐  │  │
│  │  │  Views (Pages)                              │  │  │
│  │  │  - DashboardView                            │  │  │
│  │  │  - UsersView                                │  │  │
│  │  │  - MatchesView                              │  │  │
│  │  │  - SettlementsView                          │  │  │
│  │  │  - FundView                                 │  │  │
│  │  │  - ConfigView                               │  │  │
│  │  └────────────┬────────────────────────────────┘  │  │
│  │               │                                    │  │
│  │  ┌────────────▼──────────────────────────────┐    │  │
│  │  │  Components                               │    │  │
│  │  │  - UserTable, UserForm                    │    │  │
│  │  │  - MatchList, MatchForm                   │    │  │
│  │  │  - SettlementList                         │    │  │
│  │  │  - FundTransactionList, FundForm          │    │  │
│  │  │  - StatCard, Leaderboard                  │    │  │
│  │  └────────────┬──────────────────────────────┘    │  │
│  │               │                                    │  │
│  │  ┌────────────▼──────────────────────────────┐    │  │
│  │  │  Pinia Stores (Global State)              │    │  │
│  │  │  - useUserStore                           │    │  │
│  │  │  - useMatchStore                          │    │  │
│  │  │  - useSettlementStore                     │    │  │
│  │  │  - useFundStore                           │    │  │
│  │  │  - useConfigStore                         │    │  │
│  │  └────────────┬──────────────────────────────┘    │  │
│  │               │                                    │  │
│  │  ┌────────────▼──────────────────────────────┐    │  │
│  │  │  API Service Layer (Axios)                │    │  │
│  │  │  - userService.ts                         │    │  │
│  │  │  - matchService.ts                        │    │  │
│  │  │  - settlementService.ts                   │    │  │
│  │  │  - fundService.ts                         │    │  │
│  │  │  - configService.ts                       │    │  │
│  │  └────────────┬──────────────────────────────┘    │  │
│  └───────────────┼───────────────────────────────────┘  │
└────────────────┼──────────────────────────────────────┘
                 │ HTTP/REST
                 │
┌────────────────▼──────────────────────────────────────┐
│            Go/Gin Backend API                         │
│            http://localhost:8080/api/v1               │
│  ┌────────────────────────────────────────────────┐   │
│  │  27 REST Endpoints (Already Implemented ✅)    │   │
│  └──────────────────┬─────────────────────────────┘   │
└─────────────────────┼─────────────────────────────────┘
                      │
┌─────────────────────▼─────────────────────────────────┐
│              PostgreSQL Database                      │
│  ┌────────────────────────────────────────────────┐   │
│  │  Tables: users, matches, settlements, etc.     │   │
│  └────────────────────────────────────────────────┘   │
└───────────────────────────────────────────────────────┘
```

## Data Models (TypeScript)

### Core Types

```typescript
// types/user.ts
export interface User {
  id: string;
  name: string;
  current_score: number;
  created_at: string;
  updated_at: string;
  is_active: boolean;
}

export interface CreateUserRequest {
  name: string;
}

export interface UpdateUserRequest {
  name: string;
}

// types/match.ts
export interface Match {
  id: string;
  match_type: '1v1' | '2v2';
  winner_team: 1 | 2;
  match_date: string;
  recorded_by: string;
  created_at: string;
  is_locked: boolean;
  participants: MatchParticipant[];
}

export interface MatchParticipant {
  id: string;
  match_id: string;
  user_id: string;
  team_number: 1 | 2;
  point_change: number;
  user: User;
}

export interface CreateMatchRequest {
  match_type: '1v1' | '2v2';
  team1: string[]; // User IDs
  team2: string[]; // User IDs
  winner_team: 1 | 2;
  match_date?: string;
}

// types/settlement.ts
export interface DebtSettlement {
  id: string;
  debtor_id: string;
  debt_amount: number;
  money_amount: number;
  fund_amount: number;
  winner_distribution: number;
  original_debt_points: number;
  settlement_date: string;
  created_at: string;
  debtor: User;
  winners: SettlementWinner[];
}

export interface SettlementWinner {
  id: string;
  settlement_id: string;
  winner_id: string;
  money_amount: number;
  points_deducted: number;
  winner: User;
}

// types/fund.ts
export interface FundTransaction {
  id: string;
  amount: number;
  transaction_type: 'deposit' | 'withdrawal';
  description: string;
  related_settlement_id?: string;
  transaction_date: string;
  created_at: string;
}

export interface CreateDepositRequest {
  amount: number;
  description: string;
  date?: string;
}

export interface CreateWithdrawalRequest {
  amount: number;
  description: string;
  date?: string;
}

// types/config.ts
export interface Config {
  key: string;
  value: string;
  description: string;
}

export interface UpdateConfigRequest {
  value: string;
}
```

## API Service Layer

### Service Pattern
```typescript
// services/api.ts
import axios from 'axios';

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Intercept responses to handle errors globally
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const message = error.response?.data?.message || 'An error occurred';
   const code = error.response?.data?.code || 'UNKNOWN_ERROR';
    
    // Can add toast notification here
    console.error(`API Error [${code}]:`, message);
    
    return Promise.reject(error);
  }
);
```

### Individual Services
```typescript
// services/userService.ts
import { apiClient } from './api';
import type { User, CreateUserRequest, UpdateUserRequest } from '@/types/user';

export const userService = {
  async getAll(): Promise<User[]> {
    const { data } = await apiClient.get('/users');
    return data;
  },
  
  async getById(id: string): Promise<User> {
    const { data } = await apiClient.get(`/users/${id}`);
    return data;
  },
  
  async create(request: CreateUserRequest): Promise<User> {
    const { data } = await apiClient.post('/users', request);
    return data;
  },
  
  async update(id: string, request: UpdateUserRequest): Promise<User> {
    const { data } = await apiClient.put(`/users/${id}`, request);
    return data;
  },
  
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/users/${id}`);
  },
  
  async getLeaderboard(limit?: number): Promise<User[]> {
    const { data } = await apiClient.get('/users/leaderboard', {
      params: { limit },
    });
    return data;
  },
};
```

## State Management (Pinia Stores)

### Store Pattern
```typescript
// stores/userStore.ts
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { userService } from '@/services/userService';
import type { User } from '@/types/user';
import { ElMessage } from 'element-plus';

export const useUserStore = defineStore('user', () => {
  // State
  const users = ref<User[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  
  // Getters
  const activeUsers = computed(() => 
    users.value.filter(u => u.is_active)
  );
  
  const usersInDebt = computed(() =>
    activeUsers.value.filter(u => u.current_score < 0)
  );
  
  const topUser = computed(() =>
    activeUsers.value.length > 0
      ? activeUsers.value.reduce((prev, current) =>
          prev.current_score > current.current_score ? prev : current
        )
      : null
  );
  
  // Actions
  async function fetchUsers() {
    loading.value = true;
    error.value = null;
    try {
      users.value = await userService.getAll();
    } catch (e: any) {
      error.value = e.message;
      ElMessage.error('Failed to load users');
    } finally {
      loading.value = false;
    }
  }
  
  async function createUser(name: string) {
    const trimmed = name.trim();
    if (trimmed.length < 2 || trimmed.length > 100) {
      ElMessage.error('Name must be 2-100 characters');
      return;
    }
    
    try {
      const user = await userService.create({ name: trimmed });
      users.value.push(user);
      ElMessage.success(`Created user: ${user.name}`);
    } catch (e: any) {
      const message = e.response?.data?.message || 'Failed to create user';
      ElMessage.error(message);
      throw e;
    }
  }
  
  return {
    users,
    loading,
    error,
    activeUsers,
    usersInDebt,
    topUser,
    fetchUsers,
    createUser,
  };
});
```

## Component Architecture

### View Components (Pages)

#### Dashboard View
```typescript
// views/DashboardView.vue
<script setup lang="ts">
import { onMounted, computed } from 'vue';
import { useUserStore } from '@/stores/userStore';
import { useMatchStore } from '@/stores/matchStore';
import { useFundStore } from '@/stores/fundStore';
import StatCard from '@/components/shared/StatCard.vue';
import Leaderboard from '@/components/shared/Leaderboard.vue';
import RecentMatches from '@/components/match/RecentMatches.vue';

const userStore = useUserStore();
const matchStore = useMatchStore();
const fundStore = useFundStore();

const stats = computed(() => ({
  totalPlayers: userStore.activeUsers.length,
  todayMatches: matchStore.todayCount,
  fundBalance: fundStore.balance,
  playersInDebt: userStore.usersInDebt.length,
}));

onMounted(async () => {
  await Promise.all([
    userStore.fetchUsers(),
    matchStore.fetchStats(),
    fundStore.fetchBalance(),
  ]);
});
</script>
```

### Feature Components

#### Match Form Component
```typescript
// components/match/MatchForm.vue
<script setup lang="ts">
import { ref, computed } from 'vue';
import { useUserStore } from '@/stores/userStore';
import { useMatchStore } from '@/stores/matchStore';
import type { CreateMatchRequest } from '@/types/match';

const props = defineProps<{
  visible: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
  (e: 'created'): void;
}>();

const userStore = useUserStore();
const matchStore = useMatchStore();

const form = ref<CreateMatchRequest>({
  match_type: '1v1',
  team1: [],
  team2: [],
  winner_team: 1,
});

const teamSize = computed(() => form.value.match_type === '1v1' ? 1 : 2);

const isValid = computed(() => {
  return (
    form.value.team1.length === teamSize.value &&
    form.value.team2.length === teamSize.value &&
    !hasDuplicatePlayers.value
  );
});

const hasDuplicatePlayers = computed(() => {
  const allPlayers = [...form.value.team1, ...form.value.team2];
  return new Set(allPlayers).size !== allPlayers.length;
});

async function handleSubmit() {
  if (!isValid.value) return;
  
  await matchStore.createMatch(form.value);
  emit('created');
  emit('update:visible', false);
  resetForm();
}

function resetForm() {
  form.value = {
    match_type: '1v1',
    team1: [],
    team2: [],
    winner_team: 1,
  };
}
</script>
```

## Routing Structure

```typescript
// router/index.ts
import { createRouter, createWebHistory } from 'vue-router';

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: () => import('@/views/DashboardView.vue'),
    },
    {
      path: '/users',
      name: 'users',
      component: () => import('@/views/UsersView.vue'),
    },
    {
      path: '/matches',
      name: 'matches',
      component: () => import('@/views/MatchesView.vue'),
    },
    {
      path: '/settlements',
      name: 'settlements',
      component: () => import('@/views/SettlementsView.vue'),
    },
    {
      path: '/fund',
      name: 'fund',
      component: () => import('@/views/FundView.vue'),
    },
    {
      path: '/config',
      name: 'config',
      component: () => import('@/views/ConfigView.vue'),
    },
  ],
});
```

## UI/UX Design

### Navigation
- **Layout:** Sidebar navigation + top bar
- **Menu Items:**
  - Dashboard (home icon)
  - Users (people icon)
  - Matches (game controller icon)
  - Settlements (money icon)
  - Fund (wallet icon)
  - Config (settings icon)

### Color Scheme
- **Primary:** Element Plus blue (#409EFF)
- **Success:** Green for positive scores
- **Danger:** Red for negative scores / debt
- **Warning:** Orange for approaching debt threshold
- **Info:** Gray for neutral information

### Typography
- **Font:** System fonts (San Francisco, Segoe UI, Roboto)
- **Sizes:** 
  - Heading 1: 24px
  - Heading 2: 20px
  - Body: 14px
  - Small: 12px

## Performance Considerations

### Optimization Strategies
1. **Code Splitting:** Lazy load views with dynamic imports
2. **API Caching:** Vue Query with 30s stale time for leaderboard
3. **Debouncing:** Search inputs debounced to 300ms
4. **Virtual Scrolling:** For large lists (100+ items)
5. **Optimistic Updates:** Immediate UI feedback before API confirmation

### Bundle Size
- Target: < 300KB gzipped (initial load)
- Strategy: Tree-shaking, code splitting, lazy loading

## Security Considerations

### Client-Side Validation
- Validate all inputs before sending to API
- Match backend validation rules exactly
- Prevent XSS with proper escaping (Vue handles automatically)

### API Communication
- Use HTTPS in production
- Validate API responses
- Handle errors gracefully
- No sensitive data in localStorage (for future auth)

## Error Handling Strategy

### API Errors
```typescript
// Global error handler in api.ts
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const code = error.response?.data?.code;
    const message = error.response?.data?.message;
    
    // Map backend error codes to user-friendly messages
    const userMessage = ERROR_MESSAGES[code] || message || 'An error occurred';
    
    ElMessage.error(userMessage);
    return Promise.reject(error);
  }
);

const ERROR_MESSAGES: Record<string, string> = {
  VALIDATION_ERROR: 'Invalid input. Please check your data.',
  NOT_FOUND: 'The requested item was not found.',
  CONFLICT: 'A user with this name already exists.',
  INSUFFICIENT_BALANCE: 'Insufficient fund balance for withdrawal.',
  MATCH_LOCKED: 'Cannot modify a locked match.',
  // ... more mappings
};
```

### Form Validation
```typescript
// Use Element Plus validation rules
const rules = {
  name: [
    { required: true, message: 'Name is required', trigger: 'blur' },
    { min: 2, max: 100, message: 'Name must be 2-100 characters', trigger: 'blur' },
  ],
  amount: [
    { required: true, message: 'Amount is required', trigger: 'blur' },
    { type: 'number', min: 1, message: 'Amount must be positive', trigger: 'blur' },
  ],
};
```

## Testing Strategy

### Unit Tests (Vitest)
- Test Pinia stores (actions, getters)
- Test utility functions (formatVND, etc.)
- Test form validation logic

### Component Tests (Vue Test Utils)
- Test component rendering
- Test user interactions (clicks, form submissions)
- Test prop passing and event emission

### Integration Tests
- Test full user flows (create user → record match → view leaderboard)
- Test API service layer with mock responses

### E2E Tests (Optional - Cypress/Playwright)
- Test complete workflows
- Test across different browsers

## Accessibility

### WCAG AA Compliance
- Color contrast ratios ≥ 4.5:1
- Keyboard navigation for all interactive elements
- ARIA labels for screen readers
- Focus indicators visible
- Form inputs have associated labels

### Semantic HTML
- Use proper heading hierarchy (h1 → h2 → h3)
- Use `<button>` for actions, `<a>` for navigation
- Use `<table>` for tabular data with proper headers

## Deployment Considerations

### Environment Variables
```env
# .env.development
VITE_API_URL=http://localhost:8080/api/v1

# .env.production
VITE_API_URL=https://api.yourdomain.com/api/v1
```

### Build Process
```bash
# Development
npm run dev

# Production build
npm run build

# Preview production build
npm run preview
```

### Deployment Targets
- **Development:** Local (localhost:5173)
- **Production:** Vercel (planned)
  - Automatic deployments from `main` branch
  - Environment variables configured in Vercel dashboard
  - API proxy if needed for CORS

## Design Decisions

### Decision 1: Vue Query vs Pure Pinia
**Chosen:** Hybrid approach (Pinia + Vue Query)
**Rationale:**
- Pinia for local state (forms, UI state)
- Vue Query for server state (automatic caching, refetching)
- Best of both worlds

**Alternative Considered:** Pure Pinia
- Would require manual cache management
- More boilerplate code

### Decision 2: Modal vs Page for Match Recording
**Chosen:** Modal
**Rationale:**
- Quick access from any page
- Maintains context (user stays on current page)
- Faster workflow

**Alternative Considered:** Dedicated page
- More space for complex forms
- But requires navigation away from current context

### Decision 3: Real-time Updates vs Manual Refresh
**Chosen:** Manual refresh (Phase 1), Real-time later (Phase 2)
**Rationale:**
- Simpler implementation for MVP
- Single-user system doesn't require real-time urgently
- Can add WebSocket later

**Alternative Considered:** WebSocket from start
- Would require backend changes
- Adds complexity for limited benefit initially

## Next Steps

1. Create implementation plan with detailed task breakdown
2. Set up project structure (directories, base components)
3. Implement in phases:
   - Core (types, services, stores)
   - User management UI
   - Match recording UI
   - Settlement & Fund UI
   - Configuration UI
   - Dashboard & polish
