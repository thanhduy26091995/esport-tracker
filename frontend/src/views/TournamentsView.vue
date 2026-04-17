<template>
  <div class="page-wrapper">
    <div class="page-container">
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">{{ t('tournaments.pageTitle') }}</h1>
          <p class="page-subtitle">{{ t('tournaments.pageSubtitle') }}</p>
        </div>
        <el-button type="primary" class="tournament-create-button" @click="router.push('/tournaments/create')" :icon="Plus" size="large">
          {{ t('tournaments.createButton') }}
        </el-button>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
        <StatCard
          :title="t('tournaments.statTotal')"
          :value="store.tournaments.length"
          :icon="Trophy"
          :loading="store.loading"
          type="info"
        />
        <StatCard
          :title="t('tournaments.statActive')"
          :value="activeTournaments"
          :icon="TrendCharts"
          :loading="store.loading"
          type="success"
        />
        <StatCard
          :title="t('tournaments.statCompleted')"
          :value="completedTournaments"
          :icon="CircleCheck"
          :loading="store.loading"
          type="default"
        />
      </div>

      <!-- Table -->
      <div class="card">
        <div class="card-body">
          <div class="tournament-table-wrap">
            <el-table
              :data="store.tournaments"
              v-loading="store.loading"
              stripe
              class="tournament-table"
              :empty-text="t('tournaments.empty')"
            >
            <el-table-column prop="name" :label="t('tournaments.colName')" min-width="180" />
            <el-table-column prop="match_type" :label="t('tournaments.colType')" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ getMatchTypeLabel(row.match_type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('tournaments.colPlayers')" width="100">
              <template #default="{ row }">
                {{ row.participants?.length ?? 0 }}
              </template>
            </el-table-column>
            <el-table-column prop="status" :label="t('tournaments.colStatus')" width="130">
              <template #default="{ row }">
                <el-tag :type="row.status === 'completed' ? 'success' : 'primary'" size="small">
                  {{ getTournamentStatusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('tournaments.colAffectsScore')" width="130">
              <template #default="{ row }">
                <el-tag :type="row.affects_score ? 'warning' : 'info'" size="small">
                  {{ getTournamentAffectsScoreLabel(row.affects_score) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('tournaments.colCreated')" width="120">
              <template #default="{ row }">
                {{ formatDate(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column :label="t('common.actions')" width="170" align="right">
              <template #default="{ row }">
                <div class="tournament-actions-cell">
                  <el-button size="small" text @click="router.push(`/tournaments/${row.id}`)">{{ t('common.view') }}</el-button>
                  <el-button size="small" text type="danger" @click="handleDelete(row)">{{ t('common.delete') }}</el-button>
                </div>
              </template>
            </el-table-column>
            </el-table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Plus, Trophy, CircleCheck, TrendCharts } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'
import { useTournamentStore } from '@/stores/tournamentStore'
import StatCard from '@/components/shared/StatCard.vue'
import type { Tournament } from '@/types/tournament'
import { formatDate } from '@/utils/date'
import { getMatchTypeLabel, getTournamentAffectsScoreLabel, getTournamentStatusLabel } from '@/utils/tournamentLabels'

const router = useRouter()
const { t } = useI18n()
const store = useTournamentStore()

const activeTournaments = computed(() => store.tournaments.filter(t => t.status === 'active').length)
const completedTournaments = computed(() => store.tournaments.filter(t => t.status === 'completed').length)

onMounted(() => store.fetchTournaments())

const handleDelete = (tournament: Tournament) => {
  ElMessageBox.confirm(
    t('tournaments.deleteConfirm', { name: tournament.name }),
    t('tournaments.deleteTitle'),
    {
      confirmButtonText: t('common.delete'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
      confirmButtonClass: 'el-button--danger',
    }
  )
    .then(() => store.deleteTournament(tournament.id))
    .catch(() => {})
}
</script>

<style scoped>
.tournament-table-wrap {
  width: 100%;
  overflow-x: auto;
}

.tournament-table {
  min-width: 760px;
}

.tournament-actions-cell {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
  flex-wrap: wrap;
}

@media (max-width: 640px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
  }

  .page-header-left {
    width: 100%;
  }

  .page-subtitle {
    white-space: normal;
  }

  .tournament-create-button {
    width: 100%;
  }

  .tournament-table {
    min-width: 700px;
  }

  .tournament-actions-cell {
    justify-content: flex-start;
  }
}
</style>
