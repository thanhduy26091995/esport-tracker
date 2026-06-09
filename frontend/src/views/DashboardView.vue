<template>
  <div class="page-wrapper">
    <div class="page-container">

      <!-- Header -->
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">{{ t('dashboard.pageTitle') }}</h1>
          <p class="page-subtitle">{{ t('dashboard.pageSubtitle') }}</p>
        </div>
        <div class="page-header-actions">
          <el-button
            type="warning"
            plain
            :icon="Star"
            @click="handleAddBonus"
            :disabled="userStore.users.length === 0"
          >
            {{ t('matches.bonus.dialogTitle') }}
          </el-button>
          <el-button
            type="primary"
            :icon="Trophy"
            @click="handleRecordMatch"
            :disabled="userStore.users.length < 2"
          >
            {{ t('dashboard.recordMatch') }}
          </el-button>
        </div>
      </div>

      <div v-if="userStore.users.length < 2" class="info-banner info-banner--warning mb-6">
        <el-icon><Warning /></el-icon>
        {{ t('dashboard.needMorePlayers') }}
      </div>

      <!-- Unified dashboard grid: 2 cols mobile / 4 cols desktop, gap 16px -->
      <!-- Stat cards: 1 col each · Champion banner: all cols · Content cards: 2 cols each -->
      <div class="dashboard-grid">

        <StatCard :title="t('dashboard.statTotalPlayers')"   :value="userStore.users.length"       :icon="User"       :loading="userStore.loading"                          type="info"    />
        <StatCard :title="t('dashboard.statTodayMatches')" :value="getTodayMatchesCount()"        :icon="Trophy"     :loading="matchStore.loading"                         type="success" />
        <StatCard :title="t('dashboard.statFundBalance')"    :value="formatVND(fundStore.balance)"  :icon="Wallet"     :loading="fundStore.loading"                          type="warning" />
        <StatCard :title="t('dashboard.statPlayersInDebt')" :value="getDebtorsCount()"             :icon="TrendCharts" :loading="userStore.loading || configStore.loading"  type="danger"  />

        <!-- Champion banner: spans all columns -->
        <div v-if="championClub" class="champion-banner">
          <div class="champion-bg-text" aria-hidden="true">{{ championClub.name }}</div>
          <div class="champion-left">
            <div class="champion-avatar-wrap">
              <UserAvatar :avatar-url="champion?.avatar_url" :name="champion?.name || ''" size="lg" />
            </div>
            <div class="champion-info">
              <span class="champion-league">{{ championClub.league }}</span>
              <span class="champion-club">{{ championClub.name }}</span>
              <span class="champion-player">👑 {{ champion?.name }}</span>
            </div>
          </div>
          <div class="champion-score">
            <span class="champion-score-value">{{ (champion?.current_score ?? 0) > 0 ? '+' : '' }}{{ champion?.current_score ?? 0 }}</span>
            <span class="champion-score-label">điểm</span>
          </div>
        </div>

        <!-- Content cards: each spans 2 columns -->
        <div class="card content-card">
          <div class="card-header">
            <span class="card-title">{{ t('dashboard.topPlayers') }}</span>
            <router-link to="/users" class="view-all-link">{{ t('dashboard.viewAll') }}</router-link>
          </div>
          <div class="card-body">
            <UserTable
              :users="leaderboardUsers.slice(0, 10)"
              :loading="userStore.loading"
              :conversion-rate="configStore.pointToVnd"
              :debt-threshold="configStore.debtThreshold"
              :min-matches-for-tier="configStore.minMatchesForTier"
              :show-filter-bar="false"
              :show-actions="false"
            />
          </div>
        </div>

        <div class="card content-card">
          <div class="card-header">
            <span class="card-title">{{ t('dashboard.recentMatches') }}</span>
            <router-link to="/matches" class="view-all-link">{{ t('dashboard.viewAll') }}</router-link>
          </div>
          <div class="card-body">
            <RecentMatches :matches="matchStore.matches.slice(0, 10)" :users="userStore.users" />
          </div>
        </div>

        <div class="card content-card">
          <div class="card-header">
            <span class="card-title">{{ t('dashboard.recentSettlements') }}</span>
            <router-link to="/settlements" class="view-all-link">{{ t('dashboard.viewAll') }}</router-link>
          </div>
          <div class="card-body">
            <div v-if="settlementStore.loading" class="loading-center">
              <el-icon class="animate-spin" :size="24" style="color:var(--text-muted)"><Loading /></el-icon>
            </div>
            <div v-else-if="recentSettlements.length === 0" class="empty-state">
              <el-icon :size="36" class="empty-state-icon"><Document /></el-icon>
              <p class="empty-state-title">{{ t('dashboard.noSettlements') }}</p>
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

        <div class="card content-card">
          <div class="card-header">
            <span class="card-title">{{ t('dashboard.recentFundActivity') }}</span>
            <router-link to="/fund" class="view-all-link">{{ t('dashboard.viewAll') }}</router-link>
          </div>
          <div class="card-body">
            <div v-if="fundStore.loading" class="loading-center">
              <el-icon class="animate-spin" :size="24" style="color:var(--text-muted)"><Loading /></el-icon>
            </div>
            <div v-else-if="recentTransactions.length === 0" class="empty-state">
              <el-icon :size="36" class="empty-state-icon"><Wallet /></el-icon>
              <p class="empty-state-title">{{ t('dashboard.noTransactions') }}</p>
            </div>
            <div v-else class="item-list">
              <div
                v-for="tx in recentTransactions" :key="tx.id"
                class="item-row"
                @click="$router.push('/fund')"
              >
                <div class="item-avatar" :class="tx.transaction_type === 'deposit' ? 'item-avatar--green' : 'item-avatar--red'">
                  <el-icon :size="14"><component :is="tx.transaction_type === 'deposit' ? Plus : Minus" /></el-icon>
                </div>
                <div class="item-info">
                  <p class="item-title">{{ tx.transaction_type === 'deposit' ? t('dashboard.deposit') : t('dashboard.withdrawal') }}</p>
                  <p class="item-sub">{{ tx.description }}</p>
                </div>
                <div class="item-amount" :class="tx.transaction_type === 'deposit' ? 'item-amount--green' : 'item-amount--red'">
                  {{ tx.transaction_type === 'deposit' ? '+' : '-' }}{{ formatVND(tx.amount) }}
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="card content-card">
          <div class="card-header">
            <span class="card-title">{{ t('dashboard.fundContributors') }}</span>
          </div>
          <div class="card-body">
            <FundContributors
              :contributors="settlementStore.fundContributors"
              :loading="settlementStore.loading"
              :point-to-vnd="configStore.pointToVnd"
            />
          </div>
        </div>

        <div class="card content-card">
          <div class="card-header">
            <span class="card-title">{{ t('dashboard.winnerContributors') }}</span>
          </div>
          <div class="card-body">
            <WinnerContributors
              :contributors="settlementStore.winnerContributors"
              :loading="settlementStore.loading"
            />
          </div>
        </div>

      </div>
    </div>

    <MatchForm
      v-model="showMatchForm"
      :users="userStore.users"
      :debt-threshold="configStore.debtThreshold"
      :points-per-win="configStore.pointsPerWin"
      :loading="matchStore.loading"
      @submit="handleSubmitMatch"
      @cancel="() => showMatchForm = false"
      @request-users-refresh="handleRefreshUsers"
    />
    <ScoreBonusForm v-model="showBonusForm" :users="activeUsers" :loading="matchStore.loading"
      @submit="handleSubmitBonus" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
import { Trophy, User, Plus, Minus, Warning, Wallet, TrendCharts, Document, Loading, Star } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/userStore'
import { useMatchStore } from '@/stores/matchStore'
import { useSettlementStore } from '@/stores/settlementStore'
import { useFundStore } from '@/stores/fundStore'
import { useConfigStore } from '@/stores/configStore'
import StatCard from '@/components/shared/StatCard.vue'
import UserTable from '@/components/user/UserTable.vue'
import RecentMatches from '@/components/match/RecentMatches.vue'
import MatchForm from '@/components/match/MatchForm.vue'
import ScoreBonusForm from '@/components/match/ScoreBonusForm.vue'
import FundContributors from '@/components/settlement/FundContributors.vue'
import WinnerContributors from '@/components/settlement/WinnerContributors.vue'
import { formatVND } from '@/utils/formatters'
import { formatDate } from '@/utils/date'
import type { CreateMatchRequest } from '@/types/match'
import type { CreateScoreBonusRequest } from '@/types/scoreBonus'
import { sortByStrategy } from '@/utils/sort'
import { CLUBS } from '@/config/clubs'
import UserAvatar from '@/components/shared/UserAvatar.vue'

const userStore = useUserStore()
const matchStore = useMatchStore()
const settlementStore = useSettlementStore()
const fundStore = useFundStore()
const configStore = useConfigStore()
const showMatchForm = ref(false)
const showBonusForm = ref(false)
const activeUsers = computed(() => userStore.users.filter(u => u.is_active))

const leaderboardUsers = computed(() => sortByStrategy(userStore.users, 'default'))

const champion = computed(() =>
  [...userStore.users].sort((a, b) => b.current_score - a.current_score)[0]
)
const championClub = computed(() => {
  const slug = champion.value?.favorite_club
  if (!slug || slug === 'none') return null
  return CLUBS.find(c => c.slug === slug) ?? null
})

onMounted(async () => {
  await Promise.all([
    userStore.fetchUsers(),
    matchStore.fetchMatches(),
    settlementStore.fetchSettlements(),
    settlementStore.fetchFundContributors(),
    settlementStore.fetchWinnerContributors(),
    fundStore.fetchStats(),
    fundStore.fetchTransactions(),
    configStore.fetchConfigs()
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
const getUserName = (id: string) => userStore.users.find(u => u.id === id)?.name || t('common.unknown')

const handleRecordMatch = () => { showMatchForm.value = true }
const handleAddBonus = () => { showBonusForm.value = true }
const handleSubmitBonus = async (req: CreateScoreBonusRequest) => {
  try { await matchStore.createBonus(req); showBonusForm.value = false; await userStore.fetchUsers() } catch {}
}
const handleSubmitMatch = async (data: CreateMatchRequest) => {
  try { await matchStore.createMatch(data); showMatchForm.value = false; await userStore.fetchUsers() } catch {}
}
const handleRefreshUsers = async () => {
  await userStore.fetchUsers()
}
</script>

<style scoped>
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}
@media (min-width: 1024px) {
  .dashboard-grid { grid-template-columns: repeat(4, 1fr); }
}

/* Champion banner + content cards span 2 cols on mobile (full width) */
/* and 2 of 4 cols on desktop — same track widths as stat cards */
.champion-banner,
.content-card {
  grid-column: span 2;
}

/* Champion banner always fills all 4 columns */
.champion-banner {
  grid-column: 1 / -1;
}

.view-all-link {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-primary);
  text-decoration: none;
  flex-shrink: 0;
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

/* ── Champion banner ── */
.champion-banner {
  position: relative;
  overflow: hidden;
  border-radius: 16px;
  padding: 20px 24px;
  background: var(--theme-gradient);
  box-shadow: 0 8px 32px var(--theme-glow);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  transition: background 0.5s ease, box-shadow 0.5s ease;
}

.champion-bg-text {
  position: absolute;
  right: -12px;
  bottom: -8px;
  font-size: 80px;
  font-weight: 900;
  color: var(--theme-text-on-primary);
  opacity: 0.07;
  text-transform: uppercase;
  letter-spacing: -3px;
  line-height: 1;
  user-select: none;
  pointer-events: none;
  white-space: nowrap;
}

.champion-left {
  display: flex;
  align-items: center;
  gap: 16px;
  position: relative;
  z-index: 1;
}

.champion-avatar-wrap {
  flex-shrink: 0;
  filter: drop-shadow(0 4px 12px rgba(0,0,0,0.3));
}

.champion-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.champion-league {
  font-size: 10px;
  font-weight: 700;
  color: var(--theme-text-on-primary);
  opacity: 0.65;
  text-transform: uppercase;
  letter-spacing: 0.12em;
}

.champion-club {
  font-size: 24px;
  font-weight: 800;
  color: var(--theme-text-on-primary);
  line-height: 1.1;
  text-shadow: 0 2px 8px rgba(0,0,0,0.2);
}

.champion-player {
  font-size: 13px;
  font-weight: 600;
  color: var(--theme-text-on-primary);
  opacity: 0.85;
  margin-top: 2px;
}

.champion-score {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  position: relative;
  z-index: 1;
  flex-shrink: 0;
}

.champion-score-value {
  font-size: 36px;
  font-weight: 900;
  color: var(--theme-text-on-primary);
  line-height: 1;
  font-variant-numeric: tabular-nums;
  text-shadow: 0 2px 8px rgba(0,0,0,0.2);
}

.champion-score-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--theme-text-on-primary);
  opacity: 0.65;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

@media (max-width: 480px) {
  .champion-club { font-size: 18px; }
  .champion-score-value { font-size: 28px; }
  .champion-bg-text { font-size: 56px; }
}
</style>
