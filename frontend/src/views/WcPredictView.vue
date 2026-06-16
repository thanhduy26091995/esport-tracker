<template>
  <div class="page-wrapper">
    <div class="page-container">
      <!-- Header -->
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">🏆 World Cup 2026</h1>
          <p class="page-subtitle">{{ t('wc.predicting') }}</p>
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
        <!-- PREDICTIONS TAB -->
        <el-tab-pane :label="t('wc.tabPredictions')" name="predictions">
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
            <div class="empty-state-title">{{ t('wc.noOpenPredictions') }}</div>
          </div>
          <div class="wc-match-bet-list">
            <div v-for="match in betFiltered" :key="match.id">
              <WcMatchCard :match="match" :show-actions="true">
                <template #actions>
                  <el-button
                    v-if="isPredictable(match)"
                    type="success"
                    size="small"
                    @click="openPredictionForm(match)"
                    plain
                  >
                    {{ t('wc.submitPrediction') }}
                  </el-button>
                  <el-button
                    size="small"
                    text
                    @click="toggleMatchPredictions(match.id)"
                  >
                    {{ t('wc.allPredictions') }}
                    <span v-if="matchPredictionCounts[match.id] !== undefined">
                      ({{ matchPredictionCounts[match.id] }})
                    </span>
                  </el-button>
                </template>
              </WcMatchCard>
              <div v-if="expandedMatchId === match.id" class="wc-match-bets-panel">
                <WcMatchPredictionList :predictions="store.matchPredictions" />
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- OPEN PREDICTIONS TAB -->
        <el-tab-pane :label="t('wc.tabOpenPredictions')" name="open_predictions">
          <WcPredictionHistoryList :predictions="openPredictions" />
        </el-tab-pane>

        <!-- HISTORY TAB -->
        <el-tab-pane :label="t('wc.tabHistory')" name="history">
          <WcPredictionHistoryList :predictions="settledPredictions" />
        </el-tab-pane>

        <!-- LEADERBOARD TAB -->
        <el-tab-pane :label="t('wc.tabLeaderboard')" name="leaderboard">
          <WcLeaderboard :entries="store.leaderboard" />
        </el-tab-pane>

      </el-tabs>
    </div>

    <!-- Prediction Form Dialog -->
    <WcPredictionForm
      v-model="predictionFormVisible"
      :match="selectedMatch"
      :score-multipliers="selectedScoreMultipliers"
      :existing-predictions="selectedMatchPredictions"
      @prediction-placed="onPredictionPlaced"
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
import WcPredictionForm from '@/components/wc/WcPredictionForm.vue'
import WcPredictionHistoryList from '@/components/wc/WcPredictionHistoryList.vue'
import WcMatchPredictionList from '@/components/wc/WcMatchPredictionList.vue'
import WcLeaderboard from '@/components/wc/WcLeaderboard.vue'
import type { WcMatchWithOdds, WcScoreMultiplier, WcMatch } from '@/types/wc'
import { wcService } from '@/services/wcService'

const { t } = useI18n()
const router = useRouter()
const store = useWcStore()
const authStore = useWcAuthStore()

const activeTab = ref('predictions')
const predictionFormVisible = ref(false)
const selectedMatch = ref<WcMatchWithOdds | null>(null)
const selectedScoreMultipliers = ref<WcScoreMultiplier[]>([])
const expandedMatchId = ref<string | null>(null)
const matchPredictionCounts = ref<Record<string, number>>({})

const selectedMatchPredictions = computed(() =>
  selectedMatch.value
    ? store.predictions.filter((p) => p.match_id === selectedMatch.value!.id)
    : [],
)

const walletClass = computed(() => {
  const b = store.wallet?.balance ?? 0
  return b >= 0 ? 'wc-wallet-pos' : 'wc-wallet-neg'
})

const storeMatches = computed(() => store.matches)
const { search: betSearch, activeFilter: betFilter, filtered: betFiltered, counts: betCounts } =
  useMatchFilter(storeMatches, 'open')

const betFilterOptions = computed(() => [
  { key: 'incoming' as const, label: 'Sắp tới', count: betCounts.value.incoming },
  { key: 'open' as const, label: 'Mở dự đoán', count: betCounts.value.open },
  { key: 'live' as const, label: 'Đang diễn', count: betCounts.value.live },
  { key: 'locked' as const, label: 'Đã khóa', count: betCounts.value.locked },
  { key: 'completed' as const, label: 'Đã kết thúc', count: betCounts.value.completed },
  { key: 'all' as const, label: 'Tất cả', count: betCounts.value.all },
])

function isPredictable(m: WcMatch): boolean {
  if (m.status === 'completed' || m.status === 'cancelled') return false
  if (!m.predictions_open) return false
  if (m.predictions_locked_at && new Date(m.predictions_locked_at) <= new Date()) return false
  return true
}

const openPredictions = computed(() => store.predictions.filter(p => !p.result))
const settledPredictions = computed(() => store.predictions.filter(p => !!p.result))

async function openPredictionForm(match: WcMatch) {
  const full = await wcService.getMatch(match.id)
  selectedMatch.value = full
  selectedScoreMultipliers.value = full.score_multipliers ?? []
  predictionFormVisible.value = true
}

async function toggleMatchPredictions(matchId: string) {
  if (expandedMatchId.value === matchId) {
    expandedMatchId.value = null
    return
  }
  expandedMatchId.value = matchId
  await store.fetchMatchPredictions(matchId)
  matchPredictionCounts.value[matchId] = store.matchPredictions.length
}

async function onPredictionPlaced() {
  await store.fetchWallet()
  await store.fetchPredictions()
}

function handleLogout() {
  authStore.logout()
  router.push('/world-cup/login')
}

watch(activeTab, async (tab) => {
  if (tab === 'leaderboard') await store.fetchLeaderboard()
  if (tab === 'open_predictions' || tab === 'history') await store.fetchPredictions()
})

onMounted(async () => {
  await Promise.all([
    store.fetchMatches(),
    store.fetchWallet(),
    store.fetchPredictions(),
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
</style>
