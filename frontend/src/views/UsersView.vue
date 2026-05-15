<template>
  <div class="page-wrapper">
    <div class="page-container">
      <!-- Header -->
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">{{ t('users.pageTitle') }}</h1>
          <p class="page-subtitle">{{ t('users.pageSubtitle') }}</p>
        </div>
        <el-button type="primary" @click="handleAdd" :icon="Plus" size="large">
          {{ t('users.addPlayer') }}
        </el-button>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
        <StatCard
          :title="t('users.statTotalPlayers')"
          :value="userStore.users.length"
          :icon="UserIcon"
          :loading="userStore.loading"
          type="info"
        />
        <StatCard
          :title="t('users.statTopScore')"
          :value="topScore >= 0 ? `+${topScore}` : topScore"
          :icon="Trophy"
          :loading="userStore.loading"
          type="success"
        />
        <StatCard
          :title="t('users.statPlayersInDebt')"
          :value="playersInDebt"
          :icon="Warning"
          :loading="userStore.loading"
          :type="playersInDebt > 0 ? 'danger' : 'default'"
        />
      </div>

      <!-- Player Table -->
      <div class="card">
        <div class="card-header users-card-header">
          <span class="card-title">{{ t('users.pageTitle') }}</span>
          <div class="sort-tabs">
            <button
              v-for="s in sortOptions" :key="s.value"
              class="sort-tab"
              :class="{ 'sort-tab--active': sortStrategy === s.value }"
              @click="onSortChange(s.value)"
            >{{ s.label }}</button>
          </div>
        </div>
        <div class="card-body">
          <UserTable
            :users="sortedUsers"
            :loading="userStore.loading"
            :conversion-rate="configStore.pointToVnd"
            :debt-threshold="configStore.debtThreshold"
            :show-total-paid="sortStrategy === 'debt-first'"
            :min-matches-for-tier="configStore.minMatchesForTier"
            @edit="handleEdit"
            @delete="handleDeleteConfirm"
            @trigger-settlement="handleTriggerSettlement"
          />
        </div>
      </div>

      <!-- User Form Dialog -->
      <UserForm
        v-model="showDialog"
        :user="selectedUser"
        :loading="userStore.loading"
        @submit="handleSubmit"
        @cancel="handleCancel"
      />

      <!-- Settlement Trigger Dialog -->
      <SettlementTriggerDialog
        v-model="showSettlementDialog"
        :debtor="settlementDebtor"
        :users="userStore.users"
        :point-to-vnd="configStore.pointToVnd"
        :fund-split-percent="configStore.fundSplitPercent"
        :loading="settlementStore.loading"
        @confirm="handleSettlementConfirm"
        @cancel="handleSettlementCancel"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
import { Plus, User as UserIcon, Trophy, Warning } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/userStore'
import { useConfigStore } from '@/stores/configStore'
import { useSettlementStore } from '@/stores/settlementStore'
import UserTable from '@/components/user/UserTable.vue'
import UserForm from '@/components/user/UserForm.vue'
import StatCard from '@/components/shared/StatCard.vue'
import SettlementTriggerDialog from '@/components/settlement/SettlementTriggerDialog.vue'
import type { UserWithStats } from '@/types/user'
import type { WinnerInput } from '@/types/settlement'
import { sortByStrategy, type PlayerSortStrategy } from '@/utils/sort'

const userStore = useUserStore()
const configStore = useConfigStore()
const settlementStore = useSettlementStore()
const showDialog = ref(false)
const selectedUser = ref<UserWithStats | null>(null)
const showSettlementDialog = ref(false)
const settlementDebtor = ref<UserWithStats | null>(null)

const USERS_SORT_KEY = 'users-player-sort'

function readSort(key: string, fallback: PlayerSortStrategy): PlayerSortStrategy {
  try { return (localStorage.getItem(key) as PlayerSortStrategy) ?? fallback } catch { return fallback }
}

const sortStrategy = ref<PlayerSortStrategy>(readSort(USERS_SORT_KEY, 'debt-first'))

function onSortChange(s: PlayerSortStrategy) {
  sortStrategy.value = s
  try { localStorage.setItem(USERS_SORT_KEY, s) } catch {}
}

const sortedUsers = computed<UserWithStats[]>(() => {
  if (sortStrategy.value === 'debt-first') return userStore.paymentRankingUsers as unknown as UserWithStats[]
  return sortByStrategy(userStore.users, sortStrategy.value)
})

const sortOptions = computed(() => [
  { value: 'debt-first' as PlayerSortStrategy,    label: t('dashboard.sortDebtFirst') },
  { value: 'default' as PlayerSortStrategy,       label: t('dashboard.sortDefault') },
  { value: 'winners-first' as PlayerSortStrategy, label: t('dashboard.sortWinnersFirst') },
])

const topScore = computed(() => {
  if (userStore.users.length === 0) return 0
  return Math.max(...userStore.users.map(u => u.current_score))
})

const playersInDebt = computed(() =>
  userStore.users.filter(u => u.current_score < 0).length
)

onMounted(async () => {
  await Promise.all([userStore.fetchUsers(), userStore.fetchPaymentRanking(), configStore.fetchConfigs()])
})

const handleAdd = () => {
  selectedUser.value = null
  showDialog.value = true
}

const handleEdit = (user: UserWithStats) => {
  selectedUser.value = user
  showDialog.value = true
}

const handleSubmit = async (data: { name: string; tier: string; handicap_rate: number }) => {
  try {
    if (selectedUser.value) {
      await userStore.updateUser(selectedUser.value.id, data.name, data.tier, data.handicap_rate)
    } else {
      await userStore.createUser(data.name, data.tier, data.handicap_rate)
    }
    showDialog.value = false
  } catch {}
}

const handleCancel = () => {
  selectedUser.value = null
  showDialog.value = false
}

const handleDeleteConfirm = (user: UserWithStats) => {
  ElMessageBox.confirm(
    t('users.deleteConfirm', { name: user.name }),
    t('users.deleteTitle'),
    {
      confirmButtonText: t('common.delete'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
      confirmButtonClass: 'el-button--danger'
    }
  )
    .then(() => { userStore.deleteUser(user.id) })
    .catch(() => {})
}

const handleTriggerSettlement = (user: UserWithStats) => {
  settlementDebtor.value = user
  showSettlementDialog.value = true
}

const handleSettlementConfirm = async (winners: WinnerInput[]) => {
  if (!settlementDebtor.value) return

  try {
    await settlementStore.triggerSettlement(settlementDebtor.value.id, winners)
    await userStore.fetchUsers() // Refresh user data
    ElMessage.success(t('users.settlementSuccess', { name: settlementDebtor.value.name }))
    showSettlementDialog.value = false
    settlementDebtor.value = null
  } catch (error) {
    // Error already handled by store
  }
}

const handleSettlementCancel = () => {
  settlementDebtor.value = null
  showSettlementDialog.value = false
}
</script>

<style scoped>
.users-card-header {
  flex-wrap: wrap;
  row-gap: 8px;
}

.users-card-header .sort-tabs {
  flex-shrink: 0;
}
</style>
