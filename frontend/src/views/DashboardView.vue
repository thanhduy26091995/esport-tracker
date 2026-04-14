<template>
  <div class="page-wrapper">
    <div class="page-container">

      <!-- Header -->
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">Dashboard</h1>
          <p class="page-subtitle">FC25 Esport Score Tracker Overview</p>
        </div>
        <el-button
          type="primary"
          :icon="Trophy"
          @click="handleRecordMatch"
          :disabled="userStore.users.length < 2"
          size="large"
        >
          Record Match
        </el-button>
      </div>

      <div v-if="userStore.users.length < 2" class="info-banner info-banner--warning mb-6">
        <el-icon><Warning /></el-icon>
        Need at least 2 players to record a match.
      </div>

      <!-- Stats grid -->
      <div class="stats-grid">
        <StatCard title="Total Players"   :value="userStore.users.length"       :icon="User"       :loading="userStore.loading"                          type="info"    />
        <StatCard title="Today's Matches" :value="getTodayMatchesCount()"        :icon="Trophy"     :loading="matchStore.loading"                         type="success" />
        <StatCard title="Fund Balance"    :value="formatVND(fundStore.balance)"  :icon="Wallet"     :loading="fundStore.loading"                          type="warning" />
        <StatCard title="Players in Debt" :value="getDebtorsCount()"             :icon="TrendCharts" :loading="userStore.loading || configStore.loading"  type="danger"  />
      </div>

      <!-- Content grid -->
      <div class="content-grid">

        <!-- Leaderboard -->
        <div class="card">
          <div class="card-header">
            <span class="card-title">Top Players</span>
            <router-link to="/users" class="view-all-link">View all →</router-link>
          </div>
          <div class="card-body">
            <Leaderboard :users="userStore.users" :debt-threshold="configStore.debtThreshold" :limit="10" compact />
          </div>
        </div>

        <!-- Recent Matches -->
        <div class="card">
          <div class="card-header">
            <span class="card-title">Recent Matches</span>
            <router-link to="/matches" class="view-all-link">View all →</router-link>
          </div>
          <div class="card-body">
            <RecentMatches :matches="matchStore.matches.slice(0, 10)" :users="userStore.users" />
          </div>
        </div>

        <!-- Recent Settlements -->
        <div class="card">
          <div class="card-header">
            <span class="card-title">Recent Settlements</span>
            <router-link to="/settlements" class="view-all-link">View all →</router-link>
          </div>
          <div class="card-body">
            <div v-if="settlementStore.loading" class="loading-center">
              <el-icon class="animate-spin" :size="24" style="color:var(--text-muted)"><Loading /></el-icon>
            </div>
            <div v-else-if="recentSettlements.length === 0" class="empty-state">
              <el-icon :size="36" class="empty-state-icon"><Document /></el-icon>
              <p class="empty-state-title">No settlements yet</p>
            </div>
            <div v-else class="item-list">
              <div
                v-for="s in recentSettlements" :key="s.id"
                class="item-row"
                @click="$router.push('/settlements')"
              >
                <div class="item-avatar item-avatar--red">
                  <el-icon :size="14"><Document /></el-icon>
                </div>
                <div class="item-info">
                  <p class="item-title">{{ getUserName(s.debtor_id) }}</p>
                  <p class="item-sub">{{ formatDate(s.created_at) }}</p>
                </div>
                <div class="item-amount item-amount--red">-{{ formatVND(s.money_amount) }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Fund Activity -->
        <div class="card">
          <div class="card-header">
            <span class="card-title">Recent Fund Activity</span>
            <router-link to="/fund" class="view-all-link">View all →</router-link>
          </div>
          <div class="card-body">
            <div v-if="fundStore.loading" class="loading-center">
              <el-icon class="animate-spin" :size="24" style="color:var(--text-muted)"><Loading /></el-icon>
            </div>
            <div v-else-if="recentTransactions.length === 0" class="empty-state">
              <el-icon :size="36" class="empty-state-icon"><Wallet /></el-icon>
              <p class="empty-state-title">No transactions yet</p>
            </div>
            <div v-else class="item-list">
              <div
                v-for="t in recentTransactions" :key="t.id"
                class="item-row"
                @click="$router.push('/fund')"
              >
                <div class="item-avatar" :class="t.transaction_type === 'deposit' ? 'item-avatar--green' : 'item-avatar--red'">
                  <el-icon :size="14"><component :is="t.transaction_type === 'deposit' ? Plus : Minus" /></el-icon>
                </div>
                <div class="item-info">
                  <p class="item-title capitalize">{{ t.transaction_type }}</p>
                  <p class="item-sub">{{ t.description }}</p>
                </div>
                <div class="item-amount" :class="t.transaction_type === 'deposit' ? 'item-amount--green' : 'item-amount--red'">
                  {{ t.transaction_type === 'deposit' ? '+' : '-' }}{{ formatVND(t.amount) }}
                </div>
              </div>
            </div>
          </div>
        </div>

      </div>
    </div>

    <MatchForm
      v-model="showMatchForm"
      :users="userStore.users"
      :debt-threshold="configStore.debtThreshold"
      :loading="matchStore.loading"
      @submit="handleSubmitMatch"
      @cancel="() => showMatchForm = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Trophy, User, Plus, Minus, Warning, Wallet, TrendCharts, Document, Loading } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/userStore'
import { useMatchStore } from '@/stores/matchStore'
import { useSettlementStore } from '@/stores/settlementStore'
import { useFundStore } from '@/stores/fundStore'
import { useConfigStore } from '@/stores/configStore'
import StatCard from '@/components/shared/StatCard.vue'
import Leaderboard from '@/components/shared/Leaderboard.vue'
import RecentMatches from '@/components/match/RecentMatches.vue'
import MatchForm from '@/components/match/MatchForm.vue'
import { formatVND } from '@/utils/formatters'
import { formatDate } from '@/utils/date'
import type { CreateMatchRequest } from '@/types/match'

const userStore = useUserStore()
const matchStore = useMatchStore()
const settlementStore = useSettlementStore()
const fundStore = useFundStore()
const configStore = useConfigStore()
const showMatchForm = ref(false)

onMounted(async () => {
  await Promise.all([
    userStore.fetchUsers(), matchStore.fetchMatches(),
    settlementStore.fetchSettlements(), fundStore.fetchStats(),
    fundStore.fetchTransactions(), configStore.fetchConfigs()
  ])
})

const recentSettlements = computed(() => settlementStore.settlements.slice(0, 5))
const recentTransactions = computed(() => fundStore.transactions.slice(0, 5))

const getTodayMatchesCount = () => {
  const today = new Date(); today.setHours(0, 0, 0, 0)
  return matchStore.matches.filter(m => {
    const d = new Date(m.match_date || m.created_at); d.setHours(0, 0, 0, 0)
    return d.getTime() === today.getTime()
  }).length
}

const getDebtorsCount = () => userStore.users.filter(u => u.current_score < configStore.debtThreshold).length
const getUserName = (id: string) => userStore.users.find(u => u.id === id)?.name || 'Unknown'

const handleRecordMatch = () => { showMatchForm.value = true }
const handleSubmitMatch = async (data: CreateMatchRequest) => {
  try { await matchStore.createMatch(data); showMatchForm.value = false; await userStore.fetchUsers() } catch {}
}
</script>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
  margin-bottom: 20px;
}
@media (min-width: 1024px) {
  .stats-grid { grid-template-columns: repeat(4, 1fr); }
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
}
@media (min-width: 1024px) {
  .content-grid { grid-template-columns: repeat(2, 1fr); }
}

.view-all-link {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-primary);
  text-decoration: none;
}
.view-all-link:hover { color: var(--color-primary-dark); }

.info-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
}
.info-banner--warning {
  background: var(--color-warning-bg);
  color: var(--color-warning);
  border: 1px solid var(--color-warning-border);
}

.loading-center { display: flex; justify-content: center; padding: 40px 0; }

.item-list { display: flex; flex-direction: column; gap: 2px; }

.item-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 9px 12px;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.15s;
}
.item-row:hover { background: var(--surface-page); }

.item-avatar {
  width: 32px; height: 32px;
  border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.item-avatar--green { background: var(--color-success-bg); color: var(--color-success); }
.item-avatar--red   { background: var(--color-danger-bg);  color: var(--color-danger);  }
.item-avatar--blue  { background: var(--color-info-bg);    color: var(--color-info);    }

.item-info { flex: 1; min-width: 0; }
.item-title { font-size: 13px; font-weight: 600; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.item-sub   { font-size: 11px; color: var(--text-muted); margin-top: 1px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.item-amount { font-size: 13px; font-weight: 700; flex-shrink: 0; }
.item-amount--green { color: var(--color-success); }
.item-amount--red   { color: var(--color-danger);  }
</style>
