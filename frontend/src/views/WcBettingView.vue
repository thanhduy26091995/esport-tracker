<template>
  <div class="page-wrapper">
    <div class="page-container">
      <!-- Header -->
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">🏆 World Cup 2026</h1>
          <p class="page-subtitle">{{ t('wc.betting') }}</p>
        </div>
        <div class="wc-user-header">
          <div class="wc-wallet-badge">
            <span class="wc-wallet-label">{{ t('wc.balance') }}</span>
            <span class="wc-wallet-value" :class="walletClass">
              {{ store.wallet?.balance ?? 0 }} {{ t('wc.points') }}
            </span>
          </div>
          <div class="wc-user-info">
            <span class="wc-user-name">{{ authStore.userName }}</span>
            <span v-if="authStore.isAdmin" class="wc-admin-badge">Admin</span>
          </div>
          <el-button size="small" text @click="handleLogout">{{ t('wc.logout') }}</el-button>
        </div>
      </div>

      <!-- Tabs -->
      <el-tabs v-model="activeTab" class="wc-main-tabs">
        <!-- BETTING TAB -->
        <el-tab-pane :label="t('wc.tabBetting')" name="betting">
          <!-- Filter bar -->
          <div class="wc-filter-bar">
            <el-input
              v-model="betSearch"
              placeholder="Tìm đội bóng..."
              clearable
              size="small"
              class="wc-filter-search"
            >
              <template #prefix>🔍</template>
            </el-input>
            <div class="wc-filter-pills">
              <button
                v-for="f in betFilterOptions"
                :key="f.key"
                class="wc-filter-pill"
                :class="{ 'wc-filter-pill--active': betFilter === f.key }"
                @click="betFilter = f.key"
              >
                {{ f.label }}
                <span v-if="f.count > 0" class="wc-filter-count">{{ f.count }}</span>
              </button>
            </div>
          </div>

          <div v-if="betFiltered.length === 0 && !store.loading" class="empty-state">
            <div class="empty-state-title">{{ t('wc.noBettableMatches') }}</div>
          </div>
          <div class="wc-match-bet-list">
            <div v-for="match in betFiltered" :key="match.id">
              <WcMatchCard :match="match" :show-actions="true">
                <template #actions>
                  <el-button
                    v-if="isBettable(match)"
                    type="success"
                    size="small"
                    @click="openBetForm(match)"
                    plain
                  >
                    {{ t('wc.placeBet') }}
                  </el-button>
                  <el-button
                    size="small"
                    text
                    @click="toggleMatchBets(match.id)"
                  >
                    {{ t('wc.allBets') }}
                    <span v-if="matchBetCounts[match.id] !== undefined">
                      ({{ matchBetCounts[match.id] }})
                    </span>
                  </el-button>
                  <el-button
                    size="small"
                    text
                    @click="toggleCustomBets(match.id)"
                  >
                    Kèo phụ
                  </el-button>
                </template>
              </WcMatchCard>
              <div v-if="expandedMatchId === match.id" class="wc-match-bets-panel">
                <WcMatchBetList :bets="store.matchBets" />
              </div>
              <div v-if="expandedCustomBetMatchId === match.id" class="wc-custom-bets-panel">
                <div v-if="!customBetsByMatch[match.id] || customBetsByMatch[match.id].length === 0" class="wc-custom-bets-empty">
                  Chưa có kèo phụ cho trận này.
                </div>
                <WcCustomBetCard
                  v-for="bet in customBetsByMatch[match.id]"
                  :key="bet.id"
                  :bet="bet"
                  @refresh="refreshCustomBets(match.id)"
                />
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- OPEN BETS TAB -->
        <el-tab-pane :label="t('wc.tabOpenBets')" name="open_bets">
          <WcBetHistoryList :bets="openBets" />
          <div v-if="pendingCustomEntries.length > 0" class="wc-custom-history-section">
            <div class="wc-custom-history-title">Kèo phụ đang chờ</div>
            <WcCustomBetHistoryList :entries="pendingCustomEntries" />
          </div>
        </el-tab-pane>

        <!-- HISTORY TAB -->
        <el-tab-pane :label="t('wc.tabHistory')" name="history">
          <WcBetHistoryList :bets="settledBets" />
          <div v-if="settledCustomEntries.length > 0" class="wc-custom-history-section">
            <div class="wc-custom-history-title">Kèo phụ</div>
            <WcCustomBetHistoryList :entries="settledCustomEntries" />
          </div>
        </el-tab-pane>

        <!-- LEADERBOARD TAB -->
        <el-tab-pane :label="t('wc.tabLeaderboard')" name="leaderboard">
          <WcLeaderboard :entries="store.leaderboard" />
        </el-tab-pane>

        <!-- ADMIN TAB -->
        <el-tab-pane :label="t('wc.tabAdmin')" name="admin" :disabled="!authStore.isAdmin">
          <WcAdminPanel v-if="authStore.isAdmin" />
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- Bet Form Dialog -->
    <WcBetForm
      v-model="betFormVisible"
      :match="selectedMatch"
      :score-odds="selectedScoreOdds"
      :existing-bets="selectedMatchBets"
      @bet-placed="onBetPlaced"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useWcStore } from '@/stores/wcStore'
import { useWcAuthStore } from '@/stores/wcAuthStore'
import { useMatchFilter } from '@/composables/useMatchFilter'
import WcMatchCard from '@/components/wc/WcMatchCard.vue'
import WcBetForm from '@/components/wc/WcBetForm.vue'
import WcBetHistoryList from '@/components/wc/WcBetHistoryList.vue'
import WcMatchBetList from '@/components/wc/WcMatchBetList.vue'
import WcLeaderboard from '@/components/wc/WcLeaderboard.vue'
import WcAdminPanel from '@/components/wc/WcAdminPanel.vue'
import WcCustomBetCard from '@/components/wc/WcCustomBetCard.vue'
import WcCustomBetHistoryList from '@/components/wc/WcCustomBetHistoryList.vue'
import type { WcMatchWithOdds, WcScoreOdds, WcMatch, WcCustomBetWithOptions, WcCustomBetEntryHistory } from '@/types/wc'
import { wcService } from '@/services/wcService'

const { t } = useI18n()
const router = useRouter()
const store = useWcStore()
const authStore = useWcAuthStore()

const activeTab = ref('betting')
const betFormVisible = ref(false)
const selectedMatch = ref<WcMatchWithOdds | null>(null)
const selectedScoreOdds = ref<WcScoreOdds[]>([])
const expandedMatchId = ref<string | null>(null)
const matchBetCounts = ref<Record<string, number>>({})
const expandedCustomBetMatchId = ref<string | null>(null)
const customBetsByMatch = ref<Record<string, WcCustomBetWithOptions[]>>({})

const selectedMatchBets = computed(() =>
  selectedMatch.value
    ? store.bets.filter((b) => b.match_id === selectedMatch.value!.id)
    : [],
)

const walletClass = computed(() => {
  const b = store.wallet?.balance ?? 0
  return b >= 0 ? 'wc-wallet-pos' : 'wc-wallet-neg'
})

const storeMatches = computed(() => store.matches)
const { search: betSearch, activeFilter: betFilter, filtered: betFiltered, counts: betCounts } =
  useMatchFilter(storeMatches, 'incoming')

const betFilterOptions = computed(() => [
  { key: 'incoming' as const, label: 'Sắp tới', count: betCounts.value.incoming },
  { key: 'open' as const, label: 'Mở cược', count: betCounts.value.open },
  { key: 'live' as const, label: 'Đang diễn', count: betCounts.value.live },
  { key: 'locked' as const, label: 'Đã khóa', count: betCounts.value.locked },
  { key: 'completed' as const, label: 'Đã kết thúc', count: betCounts.value.completed },
  { key: 'all' as const, label: 'Tất cả', count: betCounts.value.all },
])

function isBettable(m: WcMatch): boolean {
  if (m.status === 'completed' || m.status === 'cancelled') return false
  if (m.bets_locked_at && new Date(m.bets_locked_at) <= new Date()) return false
  return true
}

const openBets = computed(() => store.bets.filter(b => !b.result))
const settledBets = computed(() => store.bets.filter(b => !!b.result))
const customEntries = ref<WcCustomBetEntryHistory[]>([])
const pendingCustomEntries = computed(() => customEntries.value.filter(e => e.status === 'pending'))
const settledCustomEntries = computed(() => customEntries.value.filter(e => e.status !== 'pending'))

async function openBetForm(match: WcMatch) {
  const full = await wcService.getMatch(match.id)
  selectedMatch.value = full
  selectedScoreOdds.value = full.score_odds ?? []
  betFormVisible.value = true
}

async function toggleMatchBets(matchId: string) {
  if (expandedMatchId.value === matchId) {
    expandedMatchId.value = null
    return
  }
  expandedMatchId.value = matchId
  await store.fetchMatchBets(matchId)
  matchBetCounts.value[matchId] = store.matchBets.length
}

async function onBetPlaced() {
  await store.fetchWallet()
  await store.fetchBets()
}

async function toggleCustomBets(matchId: string) {
  if (expandedCustomBetMatchId.value === matchId) {
    expandedCustomBetMatchId.value = null
    return
  }
  expandedCustomBetMatchId.value = matchId
  const bets = await wcService.listCustomBets(matchId)
  customBetsByMatch.value[matchId] = bets
}

async function refreshCustomBets(matchId: string) {
  const bets = await wcService.listCustomBets(matchId)
  customBetsByMatch.value[matchId] = bets
  await store.fetchWallet()
}

function handleLogout() {
  authStore.logout()
  router.push('/world-cup/login')
}

async function fetchCustomEntries() {
  customEntries.value = await wcService.getMyCustomBetEntries()
}

watch(activeTab, async (tab) => {
  if (tab === 'leaderboard') await store.fetchLeaderboard()
  if (tab === 'open_bets' || tab === 'history') {
    await store.fetchBets()
    await fetchCustomEntries()
  }
})

onMounted(async () => {
  await Promise.all([
    store.fetchMatches(),
    store.fetchWallet(),
    store.fetchBets(),
    fetchCustomEntries(),
    store.fetchPublicConfig(),
  ])
})
</script>

<style scoped>
.wc-user-header {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.wc-wallet-badge {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.wc-wallet-label {
  font-size: 10px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.wc-wallet-value {
  font-size: 18px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.wc-wallet-pos { color: #16a34a; }
.wc-wallet-neg { color: #ef4444; }

.wc-user-info {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.wc-user-name {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
}

.wc-admin-badge {
  font-size: 10px;
  font-weight: 700;
  background: rgba(217, 119, 6, 0.12);
  color: #d97706;
  padding: 1px 6px;
  border-radius: 4px;
}

@media (max-width: 540px) {
  .wc-user-header {
    justify-content: space-between;
    flex-shrink: 1;
    width: 100%;
    gap: 8px;
  }

  .wc-wallet-badge {
    flex-direction: row;
    align-items: center;
    gap: 5px;
  }

  .wc-wallet-value {
    font-size: 14px;
  }

  .wc-user-info {
    flex-direction: row;
    align-items: center;
    gap: 5px;
  }
}

.wc-main-tabs :deep(.el-tabs__nav-wrap::after) {
  height: 1px;
}

.wc-match-bet-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wc-match-bets-panel {
  margin-top: 8px;
  margin-bottom: 4px;
}

.wc-custom-bets-panel {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 8px;
  margin-bottom: 4px;
}

.wc-custom-bets-empty {
  font-size: 13px;
  color: var(--text-muted);
  text-align: center;
  padding: 12px 0;
}

.wc-filter-bar {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 16px;
}

.wc-filter-search {
  max-width: 280px;
}

.wc-filter-pills {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.wc-filter-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 12px;
  border-radius: 20px;
  border: 1px solid var(--border-default);
  background: var(--surface-card);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.wc-filter-pill:hover {
  border-color: #16a34a60;
  color: var(--text-primary);
}

.wc-filter-pill--active {
  background: #16a34a;
  border-color: #16a34a;
  color: #fff;
}

.wc-filter-count {
  font-size: 11px;
  font-weight: 700;
  background: rgba(255,255,255,0.25);
  border-radius: 8px;
  padding: 0 5px;
  line-height: 1.4;
}

.wc-filter-pill:not(.wc-filter-pill--active) .wc-filter-count {
  background: var(--surface-page);
  color: var(--text-muted);
}

.wc-custom-history-section {
  margin-top: 20px;
}

.wc-custom-history-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 10px;
}
</style>
