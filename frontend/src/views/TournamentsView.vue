<template>
  <div class="page-wrapper">
    <div class="page-container">
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">Tournaments</h1>
          <p class="page-subtitle">Manage round-robin tournaments</p>
        </div>
        <el-button type="primary" @click="router.push('/tournaments/create')" :icon="Plus" size="large">
          Create Tournament
        </el-button>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
        <StatCard
          title="Total Tournaments"
          :value="store.tournaments.length"
          :icon="Trophy"
          :loading="store.loading"
          type="info"
        />
        <StatCard
          title="Active"
          :value="activeTournaments"
          :icon="TrendCharts"
          :loading="store.loading"
          type="success"
        />
        <StatCard
          title="Completed"
          :value="completedTournaments"
          :icon="CircleCheck"
          :loading="store.loading"
          type="default"
        />
      </div>

      <!-- Table -->
      <div class="card">
        <div class="card-body">
          <el-table
            :data="store.tournaments"
            v-loading="store.loading"
            stripe
            empty-text="No tournaments yet. Create one to get started!"
          >
            <el-table-column prop="name" label="Name" min-width="180" />
            <el-table-column prop="match_type" label="Type" width="80">
              <template #default="{ row }">
                <el-tag size="small">{{ row.match_type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Players" width="90">
              <template #default="{ row }">
                {{ row.participants?.length ?? 0 }}
              </template>
            </el-table-column>
            <el-table-column prop="status" label="Status" width="110">
              <template #default="{ row }">
                <el-tag :type="row.status === 'completed' ? 'success' : 'primary'" size="small">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Affects Score" width="120">
              <template #default="{ row }">
                <el-tag :type="row.affects_score ? 'warning' : 'info'" size="small">
                  {{ row.affects_score ? 'Yes' : 'No' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Created" width="120">
              <template #default="{ row }">
                {{ new Date(row.created_at).toLocaleDateString('vi-VN') }}
              </template>
            </el-table-column>
            <el-table-column label="Actions" width="160" align="right" fixed="right">
              <template #default="{ row }">
                <el-button size="small" text @click="router.push(`/tournaments/${row.id}`)">View</el-button>
                <el-button size="small" text type="danger" @click="handleDelete(row)">Delete</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Trophy, CircleCheck, TrendCharts } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'
import { useTournamentStore } from '@/stores/tournamentStore'
import StatCard from '@/components/shared/StatCard.vue'
import type { Tournament } from '@/types/tournament'

const router = useRouter()
const store = useTournamentStore()

const activeTournaments = computed(() => store.tournaments.filter(t => t.status === 'active').length)
const completedTournaments = computed(() => store.tournaments.filter(t => t.status === 'completed').length)

onMounted(() => store.fetchTournaments())

const handleDelete = (tournament: Tournament) => {
  ElMessageBox.confirm(
    `Are you sure you want to delete "${tournament.name}"? This action cannot be undone.`,
    'Delete Tournament',
    {
      confirmButtonText: 'Delete',
      cancelButtonText: 'Cancel',
      type: 'warning',
      confirmButtonClass: 'el-button--danger',
    }
  )
    .then(() => store.deleteTournament(tournament.id))
    .catch(() => {})
}
</script>
