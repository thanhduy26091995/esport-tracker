<template>
  <div class="wc-admin-panel">
    <!-- Feature Toggle -->
    <div class="card card-body wc-admin-section">
      <div class="wc-admin-section-title">{{ t('wc.featureToggle') }}</div>
      <div class="wc-feature-toggle-row">
        <span class="wc-feature-status" :class="configEnabled ? 'wc-feat--on' : 'wc-feat--off'">
          {{ configEnabled ? t('wc.featureEnabled') : t('wc.featureDisabled') }}
        </span>
        <el-switch
          v-model="configEnabled"
          :loading="togglingFeature"
          @change="handleFeatureToggle"
        />
      </div>
    </div>

    <!-- Match Management -->
    <div class="card card-body wc-admin-section">
      <div class="wc-admin-section-title">Quản lý trận đấu</div>
      <div class="wc-admin-row">
        <el-button
          type="primary"
          :loading="syncing"
          @click="handleSync"
          :icon="Refresh"
        >
          {{ t('wc.syncMatches') }}
        </el-button>
      </div>

      <!-- Admin match filter bar -->
      <div class="wc-filter-bar">
        <el-input
          v-model="adminSearch"
          placeholder="Tìm đội bóng..."
          clearable
          size="small"
          class="wc-filter-search"
        >
          <template #prefix>🔍</template>
        </el-input>
        <div class="wc-filter-pills">
          <button
            v-for="f in adminFilterOptions"
            :key="f.key"
            class="wc-filter-pill"
            :class="{ 'wc-filter-pill--active': adminFilter === f.key }"
            @click="adminFilter = f.key"
          >
            {{ f.label }}
            <span v-if="f.count > 0" class="wc-filter-count">{{ f.count }}</span>
          </button>
        </div>
      </div>

      <div class="wc-admin-match-list">
        <div v-if="adminFiltered.length === 0" class="wc-admin-empty">Không tìm thấy trận đấu nào.</div>
        <div v-for="match in adminFiltered" :key="match.id" class="wc-admin-match-row">
          <div class="wc-admin-match-name">
            {{ match.home_team }} vs {{ match.away_team }}
            <span class="wc-admin-match-date">{{ formatDate(match.match_date) }}</span>
          </div>
          <div class="wc-admin-match-actions">
            <el-button
              v-if="!isLocked(match)"
              size="small"
              @click="handleLock(match.id)"
              :icon="Lock"
            >
              {{ t('wc.lockMatch') }}
            </el-button>
            <el-button
              v-else
              size="small"
              type="warning"
              @click="handleUnlock(match.id)"
            >
              🔓 Mở cược
            </el-button>
            <el-button
              size="small"
              type="success"
              @click="handleSettle(match.id)"
              :disabled="match.status !== 'completed'"
            >
              {{ t('wc.settleMatch') }}
            </el-button>
            <el-button
              size="small"
              type="warning"
              @click="openScoreOddsDialog(match)"
            >
              {{ t('wc.scoreOdds') }}
            </el-button>
            <el-button
              size="small"
              type="info"
              @click="openHandicapDialog(match)"
            >
              Kèo chấp
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- User & Wallet Management -->
    <div class="card card-body wc-admin-section">
      <div class="wc-admin-section-title">{{ t('wc.userManagement') }}</div>
      <div class="wc-user-table">
        <div v-for="user in store.allUsers" :key="user.id" class="wc-user-row">
          <div class="wc-user-info-col">
            <span class="wc-user-name-col">{{ user.name }}</span>
            <span v-if="user.is_admin" class="wc-admin-tag">Admin</span>
          </div>
          <div class="wc-user-wallet-col">
            <span class="wc-user-balance">
              {{ walletBalance(user.id) }} pts
            </span>
          </div>
          <div class="wc-user-actions-col">
            <el-button
              size="small"
              @click="openTopUpDialog(user)"
            >
              {{ t('wc.topUp') }}
            </el-button>
            <el-button
              size="small"
              :type="user.is_admin ? 'danger' : 'primary'"
              text
              @click="handleRoleToggle(user)"
            >
              {{ user.is_admin ? t('wc.removeAdmin') : t('wc.makeAdmin') }}
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- Settlement Panel -->
    <div class="card card-body wc-admin-section">
      <div class="wc-admin-section-title">{{ t('wc.settlementPanel') }}</div>
      <el-tabs v-model="settlementTab">
        <el-tab-pane :label="t('wc.previewSettlement')" name="preview">
          <WcSettlementPreview />
        </el-tab-pane>
        <el-tab-pane :label="t('wc.settlementHistory')" name="history">
          <WcSettlementHistory :settlements="store.settlements" />
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- Top-Up Dialog -->
    <el-dialog v-model="topUpVisible" :title="t('wc.topUp')" width="360px">
      <div v-if="topUpTarget" class="wc-topup-header">{{ topUpTarget.name }}</div>
      <el-form :model="topUpForm" label-position="top">
        <el-form-item :label="t('wc.topUpDelta')">
          <el-input-number
            v-model="topUpForm.delta"
            style="width: 100%"
            controls-position="right"
          />
        </el-form-item>
        <el-form-item :label="t('wc.topUpNote')">
          <el-input v-model="topUpForm.note" :placeholder="t('wc.topUpNote')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="topUpVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="topping" @click="handleTopUp">
          {{ t('wc.topUp') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Handicap Dialog -->
    <el-dialog v-model="handicapVisible" title="Cấu hình kèo chấp" width="440px">
      <div class="wc-so-match-name" v-if="handicapMatch">
        {{ handicapMatch.home_team }} vs {{ handicapMatch.away_team }}
      </div>
      <el-form :model="handicapForm" label-position="top" class="wc-handicap-config-form">
        <el-form-item label="Đội chấp (Handicap Team)">
          <el-radio-group v-model="handicapForm.handicap_team" style="width: 100%">
            <el-radio-button value="home">{{ handicapMatch?.home_team ?? 'Home' }}</el-radio-button>
            <el-radio-button value="away">{{ handicapMatch?.away_team ?? 'Away' }}</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="Số bàn chấp (Handicap Value)">
          <el-input-number
            v-model="handicapForm.handicap_value"
            :min="0"
            :max="5"
            :step="0.25"
            :precision="2"
            style="width: 100%"
          />
        </el-form-item>
        <div class="wc-handicap-odds-row">
          <el-form-item label="Kèo Home" style="flex: 1">
            <el-input-number
              v-model="handicapForm.odds_handicap_home"
              :min="1.01"
              :step="0.05"
              :precision="2"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="Kèo Away" style="flex: 1">
            <el-input-number
              v-model="handicapForm.odds_handicap_away"
              :min="1.01"
              :step="0.05"
              :precision="2"
              style="width: 100%"
            />
          </el-form-item>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="handicapVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="savingHandicap" @click="handleSaveHandicap">
          Lưu kèo chấp
        </el-button>
      </template>
    </el-dialog>

    <!-- Score Odds Dialog -->
    <el-dialog v-model="scoreOddsVisible" :title="t('wc.scoreOdds')" width="480px">
      <div class="wc-so-match-name" v-if="scoreOddsMatch">
        {{ scoreOddsMatch.home_team }} vs {{ scoreOddsMatch.away_team }}
      </div>
      <div class="wc-so-list">
        <div v-for="so in currentScoreOdds" :key="so.id" class="wc-so-row">
          <span class="wc-so-score">{{ so.home_score }}–{{ so.away_score }}</span>
          <el-input-number
            v-model="so.odds"
            :min="1.01"
            :step="0.05"
            :precision="2"
            size="small"
            style="width: 120px"
            @change="handleUpdateOdds(so.id, so.odds)"
          />
          <el-button size="small" type="danger" text @click="handleDeleteScoreOdds(so.id)">
            {{ t('common.delete') }}
          </el-button>
        </div>
      </div>
      <el-divider />
      <div class="wc-so-add-form">
        <span class="wc-so-add-label">Thêm tỉ số:</span>
        <el-input-number v-model="newSo.homeScore" :min="0" :max="20" size="small" style="width: 80px" />
        <span>–</span>
        <el-input-number v-model="newSo.awayScore" :min="0" :max="20" size="small" style="width: 80px" />
        <el-input-number v-model="newSo.odds" :min="1.01" :step="0.05" :precision="2" size="small" style="width: 100px" />
        <el-button type="primary" size="small" @click="handleAddScoreOdds">
          {{ t('common.create') }}
        </el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Refresh, Lock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useWcStore } from '@/stores/wcStore'
import { wcService } from '@/services/wcService'
import { useMatchFilter } from '@/composables/useMatchFilter'
import type { WcUser, WcMatch, WcScoreOdds } from '@/types/wc'
import WcSettlementPreview from './WcSettlementPreview.vue'
import WcSettlementHistory from './WcSettlementHistory.vue'

const { t } = useI18n()
const store = useWcStore()

const storeMatches = computed(() => store.matches)
const { search: adminSearch, activeFilter: adminFilter, filtered: adminFiltered, counts: adminCounts } =
  useMatchFilter(storeMatches, 'incoming')

const adminFilterOptions = computed(() => [
  { key: 'incoming' as const, label: 'Sắp tới', count: adminCounts.value.incoming },
  { key: 'open' as const, label: 'Mở cược', count: adminCounts.value.open },
  { key: 'live' as const, label: 'Đang diễn', count: adminCounts.value.live },
  { key: 'locked' as const, label: 'Đã khóa', count: adminCounts.value.locked },
  { key: 'completed' as const, label: 'Đã kết thúc', count: adminCounts.value.completed },
  { key: 'all' as const, label: 'Tất cả', count: adminCounts.value.all },
])

const settlementTab = ref('preview')
const syncing = ref(false)
const togglingFeature = ref(false)
const configEnabled = ref(store.config?.is_enabled ?? false)

const topUpVisible = ref(false)
const topUpTarget = ref<WcUser | null>(null)
const topUpForm = ref({ delta: 0, note: '' })
const topping = ref(false)

const scoreOddsVisible = ref(false)
const scoreOddsMatch = ref<WcMatch | null>(null)
const currentScoreOdds = ref<WcScoreOdds[]>([])
const newSo = ref({ homeScore: 0, awayScore: 0, odds: 3.00 })

const handicapVisible = ref(false)
const handicapMatch = ref<WcMatch | null>(null)
const savingHandicap = ref(false)
const handicapForm = ref({
  handicap_team: 'home',
  handicap_value: 0.5,
  odds_handicap_home: 1.90,
  odds_handicap_away: 1.90,
})

const walletMap = computed(() => {
  const m: Record<string, number> = {}
  for (const w of store.allWallets) m[w.wc_user_id] = w.balance
  return m
})

function walletBalance(userId: string) {
  return walletMap.value[userId] ?? 0
}

async function handleSync() {
  syncing.value = true
  try {
    await store.syncMatches()
  } finally {
    syncing.value = false
  }
}

async function handleFeatureToggle(val: boolean) {
  togglingFeature.value = true
  try {
    await store.updateConfig(val)
    if (store.config) store.config.is_enabled = val
  } finally {
    togglingFeature.value = false
  }
}

function isLocked(match: WcMatch) {
  return !!match.bets_locked_at && new Date(match.bets_locked_at) <= new Date()
}

async function handleLock(matchId: string) {
  await store.lockMatch(matchId)
}

async function handleUnlock(matchId: string) {
  await store.unlockMatch(matchId)
}

async function handleSettle(matchId: string) {
  await store.settleMatch(matchId)
}

function openTopUpDialog(user: WcUser) {
  topUpTarget.value = user
  topUpForm.value = { delta: 0, note: '' }
  topUpVisible.value = true
}

async function handleTopUp() {
  if (!topUpTarget.value) return
  topping.value = true
  try {
    await store.topUp(topUpTarget.value.id, topUpForm.value.delta, topUpForm.value.note)
    topUpVisible.value = false
  } finally {
    topping.value = false
  }
}

async function handleRoleToggle(user: WcUser) {
  await store.setUserRole(user.id, !user.is_admin)
}

function openHandicapDialog(match: WcMatch) {
  handicapMatch.value = match
  handicapForm.value = {
    handicap_team: match.handicap_team ?? 'home',
    handicap_value: match.handicap_value ?? 0.5,
    odds_handicap_home: match.odds_handicap_home ?? 1.90,
    odds_handicap_away: match.odds_handicap_away ?? 1.90,
  }
  handicapVisible.value = true
}

async function handleSaveHandicap() {
  if (!handicapMatch.value) return
  savingHandicap.value = true
  try {
    await wcService.updateMatch(handicapMatch.value.id, {
      handicap_team: handicapForm.value.handicap_team,
      handicap_value: handicapForm.value.handicap_value,
      odds_handicap_home: handicapForm.value.odds_handicap_home,
      odds_handicap_away: handicapForm.value.odds_handicap_away,
    })
    ElMessage.success('Đã lưu kèo chấp')
    handicapVisible.value = false
    await store.fetchMatches()
  } finally {
    savingHandicap.value = false
  }
}

async function openScoreOddsDialog(match: WcMatch) {
  scoreOddsMatch.value = match
  const odds = await wcService.getScoreOdds(match.id)
  currentScoreOdds.value = odds
  scoreOddsVisible.value = true
}

async function handleAddScoreOdds() {
  if (!scoreOddsMatch.value) return
  const so = await wcService.addScoreOdds(
    scoreOddsMatch.value.id,
    newSo.value.homeScore,
    newSo.value.awayScore,
    newSo.value.odds,
  )
  currentScoreOdds.value.push(so)
  newSo.value = { homeScore: 0, awayScore: 0, odds: 3.00 }
}

async function handleUpdateOdds(id: string, odds: number) {
  await wcService.updateScoreOdds(id, odds)
}

async function handleDeleteScoreOdds(id: string) {
  await wcService.deleteScoreOdds(id)
  currentScoreOdds.value = currentScoreOdds.value.filter(so => so.id !== id)
}

function formatDate(s: string) {
  return new Date(s).toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })
}

onMounted(async () => {
  await Promise.all([
    store.fetchConfig(),
    store.fetchMatches(),
    store.fetchAllUsers(),
    store.fetchAllWallets(),
    store.fetchSettlements(),
  ])
  configEnabled.value = store.config?.is_enabled ?? false
})
</script>

<style scoped>
.wc-admin-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding-top: 8px;
}

.wc-admin-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wc-admin-section-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-subtle);
}

.wc-feature-toggle-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.wc-feature-status {
  font-size: 13px;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 8px;
}

.wc-feat--on {
  background: rgba(22, 163, 74, 0.12);
  color: #16a34a;
}

.wc-feat--off {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.wc-admin-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.wc-filter-bar {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wc-filter-search {
  max-width: 260px;
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
  padding: 3px 10px;
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

.wc-admin-empty {
  font-size: 13px;
  color: var(--text-muted);
  padding: 16px 0;
  text-align: center;
}

.wc-admin-match-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 480px;
  overflow-y: auto;
}

.wc-admin-match-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  background: var(--surface-page);
  border-radius: 8px;
  flex-wrap: wrap;
}

.wc-admin-match-name {
  flex: 1;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  min-width: 150px;
}

.wc-admin-match-date {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  font-weight: 400;
}

.wc-admin-match-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.wc-user-table {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wc-user-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  background: var(--surface-page);
  border-radius: 8px;
  flex-wrap: wrap;
}

.wc-user-info-col {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 120px;
}

.wc-user-name-col {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.wc-admin-tag {
  font-size: 10px;
  font-weight: 700;
  background: rgba(217, 119, 6, 0.12);
  color: #d97706;
  padding: 1px 6px;
  border-radius: 4px;
}

.wc-user-wallet-col {
  flex-shrink: 0;
}

.wc-user-balance {
  font-size: 14px;
  font-weight: 700;
  tabular-nums: true;
  color: var(--text-primary);
}

.wc-user-actions-col {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.wc-topup-header {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 16px;
}

.wc-so-match-name {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 12px;
  text-align: center;
}

.wc-so-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 240px;
  overflow-y: auto;
}

.wc-so-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.wc-so-score {
  font-size: 16px;
  font-weight: 800;
  color: var(--text-primary);
  width: 50px;
  text-align: center;
}

.wc-so-add-form {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.wc-so-add-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.wc-handicap-config-form {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.wc-handicap-odds-row {
  display: flex;
  gap: 16px;
}
</style>
