<template>
  <div class="page-wrapper">
    <div class="page-container">
      <!-- Header -->
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">{{ t('matches.pageTitle') }}</h1>
          <p class="page-subtitle">{{ t('matches.pageSubtitle') }}</p>
        </div>
        <el-button
          type="primary"
          @click="handleRecordMatch"
          :icon="Plus"
          size="large"
          :disabled="!hasEnoughPlayers"
        >
          {{ t('matches.recordMatch') }}
        </el-button>
      </div>

      <!-- Stats (hidden when player filter is active) -->
      <div v-if="!selectedPlayerID" class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
        <StatCard
          :title="t('matches.statTotal')"
          :value="matchStore.stats.total"
          :icon="Trophy"
          :loading="matchStore.loading"
          type="info"
        />
        <StatCard
          :title="t('matches.statToday')"
          :value="matchStore.stats.today"
          :icon="Calendar"
          :loading="matchStore.loading"
          type="success"
        />
        <StatCard
          :title="t('matches.statLocked')"
          :value="lockedMatchCount"
          :icon="Lock"
          :loading="matchStore.loading"
          :type="lockedMatchCount > 0 ? 'warning' : 'default'"
        />
      </div>

      <!-- Warning -->
      <el-alert
        v-if="!hasEnoughPlayers"
        type="warning"
        :closable="false"
        show-icon
        class="mb-6"
      >
        <template #title>
          {{ t('matches.needMorePlayers') }}
          <router-link to="/users" class="text-blue-600 hover:underline ml-1">{{ t('users.addPlayers') }}</router-link>
        </template>
      </el-alert>

      <!-- Player filter -->
      <div class="matches-filter-bar mb-4">
        <el-select
          v-model="selectedPlayerID"
          :placeholder="t('matches.filterPlayer')"
          clearable
          size="default"
          class="matches-player-select"
          @change="handlePlayerFilterChange"
          @clear="handlePlayerFilterChange"
        >
          <el-option
            v-for="player in activePlayers"
            :key="player.id"
            :label="player.name"
            :value="player.id"
          />
        </el-select>
      </div>

      <!-- Match List -->
      <div class="card">
        <div class="card-body">
          <MatchList
            :matches="matchStore.matches"
            :loading="matchStore.loading"
            @delete="handleDeleteConfirm"
          />
        </div>
      </div>

      <!-- Match Form Dialog -->
      <MatchForm
        v-model="showMatchForm"
        :users="userStore.users"
        :loading="matchStore.loading"
        :debt-threshold="configStore.debtThreshold"
        :points-per-win="configStore.pointsPerWin"
        @submit="handleSubmitMatch"
        @cancel="handleCancelMatch"
        @request-users-refresh="handleRefreshUsers"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
import { Plus, Trophy, Calendar, Lock } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'
import { useMatchStore } from '@/stores/matchStore'
import { useUserStore } from '@/stores/userStore'
import { useConfigStore } from '@/stores/configStore'
import MatchForm from '@/components/match/MatchForm.vue'
import MatchList from '@/components/match/MatchList.vue'
import StatCard from '@/components/shared/StatCard.vue'
import type { CreateMatchRequest, MatchFeedItem } from '@/types/match'

const matchStore = useMatchStore()
const userStore = useUserStore()
const configStore = useConfigStore()
const showMatchForm = ref(false)
const selectedPlayerID = ref<string | null>(null)

const activePlayers = computed(() => userStore.users.filter(u => u.is_active))

const hasEnoughPlayers = computed(() => activePlayers.value.length >= 2)

const lockedMatchCount = computed(() =>
  matchStore.matches.filter(m => m.is_locked).length
)

onMounted(async () => {
  await Promise.all([
    matchStore.fetchMatches(),
    matchStore.fetchStats(),
    userStore.fetchUsers(),
    configStore.fetchConfigs()
  ])
})

const handlePlayerFilterChange = async () => {
  await matchStore.fetchMatches(
    selectedPlayerID.value ? { player_id: selectedPlayerID.value } : undefined
  )
}

const handleRecordMatch = () => {
  if (hasEnoughPlayers.value) showMatchForm.value = true
}

const handleSubmitMatch = async (data: CreateMatchRequest) => {
  try {
    await matchStore.createMatch(data)
    showMatchForm.value = false
    await userStore.fetchUsers()
  } catch {}
}

const handleRefreshUsers = async () => {
  await userStore.fetchUsers()
}

const handleCancelMatch = () => { showMatchForm.value = false }

const handleDeleteConfirm = (item: MatchFeedItem) => {
  const isBonus = item.type === 'bonus'
  ElMessageBox.confirm(
    isBonus ? t('matches.bonus.deleteConfirm') : t('matches.deleteConfirm'),
    isBonus ? t('matches.bonus.deleteTitle') : t('matches.deleteTitle'),
    {
      confirmButtonText: t('common.delete'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
      confirmButtonClass: 'el-button--danger'
    }
  )
    .then(async () => {
      if (isBonus) {
        await matchStore.deleteBonus(item.id)
      } else {
        await matchStore.deleteMatch(item.id)
      }
      await userStore.fetchUsers()
    })
    .catch(() => {})
}
</script>
