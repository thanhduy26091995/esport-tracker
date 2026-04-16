<template>
  <div>
    <div class="filter-bar">
      <el-input v-model="searchQuery" placeholder="Search players..." clearable class="w-64" :prefix-icon="Search" />
      <el-select v-model="scoreFilter" placeholder="Filter by score" clearable class="w-44">
        <el-option label="All" value="" />
        <el-option label="Positive" value="positive" />
        <el-option label="In Debt" value="negative" />
        <el-option label="Zero" value="zero" />
      </el-select>
      <span class="filter-count">{{ filteredUsers.length }} of {{ users.length }} players</span>
    </div>

    <div v-if="!loading && filteredUsers.length === 0" class="empty-state">
      <svg class="empty-state-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
      </svg>
      <p class="empty-state-title">{{ searchQuery ? 'No players found' : 'No players yet' }}</p>
      <p class="empty-state-desc">{{ searchQuery ? 'Try a different search term' : 'Add your first player to get started.' }}</p>
    </div>

    <el-table v-else :data="filteredUsers" stripe style="width:100%" v-loading="loading"
      :default-sort="{ prop: 'current_score', order: 'descending' }">
      <el-table-column type="index" label="#" width="55" />
      <el-table-column prop="name" label="Player" min-width="180">
        <template #default="{ row }">
          <div class="player-cell">
            <div class="player-avatar">{{ row.name.charAt(0).toUpperCase() }}</div>
            <span class="player-name">{{ row.name }}</span>
            <PlayerTierBadge :tier="row.tier || 'normal'" />
            <el-tag v-if="!row.is_active" type="info" size="small">Inactive</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="current_score" label="Score" width="110" sortable>
        <template #default="{ row }">
          <span class="score-pill" :class="row.current_score > 0 ? 'score-pill-positive' : row.current_score < 0 ? 'score-pill-negative' : 'score-pill-zero'">
            {{ row.current_score > 0 ? '+' : '' }}{{ row.current_score }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="Value (VND)" width="160">
        <template #default="{ row }">
          <span class="vnd-value">{{ formatVND(pointsToVND(row.current_score, conversionRate)) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="Joined" width="130">
        <template #default="{ row }">
          <span class="date-value">{{ formatDate(row.created_at) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="Actions" min-width="260" align="right" fixed="right">
        <template #default="{ row }">
          <div class="actions-cell">
            <el-tooltip v-if="row.current_score < 0 && row.current_score <= debtThreshold" content="Manually trigger settlement" placement="top">
              <el-button size="small" type="warning" plain @click="emit('triggerSettlement', row)" :icon="Warning">
                Settle Debt
              </el-button>
            </el-tooltip>
            <el-button size="small" text @click="emit('edit', row)" :icon="Edit">Edit</el-button>
            <el-button size="small" text type="danger" @click="emit('delete', row)" :icon="Delete">Delete</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Edit, Delete, Search, Warning } from '@element-plus/icons-vue'
import type { User } from '@/types/user'
import { formatVND, pointsToVND } from '@/utils/formatters'
import PlayerTierBadge from '@/components/PlayerTierBadge.vue'

interface Props { users: User[]; loading?: boolean; conversionRate?: number; debtThreshold?: number }
const props = withDefaults(defineProps<Props>(), { loading: false, conversionRate: 22000, debtThreshold: -6 })
const emit = defineEmits<{ edit: [user: User]; delete: [user: User]; triggerSettlement: [user: User] }>()

const searchQuery = ref(''); const scoreFilter = ref('')

const filteredUsers = computed(() => {
  let r = props.users
  if (searchQuery.value) { const q = searchQuery.value.toLowerCase(); r = r.filter(u => u.name.toLowerCase().includes(q)) }
  if (scoreFilter.value === 'positive') r = r.filter(u => u.current_score > 0)
  else if (scoreFilter.value === 'negative') r = r.filter(u => u.current_score < 0)
  else if (scoreFilter.value === 'zero') r = r.filter(u => u.current_score === 0)
  return r
})

const formatDate = (d: string) => new Date(d).toLocaleDateString('vi-VN', { year: 'numeric', month: '2-digit', day: '2-digit' })
</script>

<style scoped>
.player-cell { display: flex; align-items: center; gap: 10px; }

.player-avatar {
  width: 30px; height: 30px; border-radius: 50%;
  background: linear-gradient(135deg, #dbeafe, #bfdbfe);
  color: #1d4ed8; font-size: 12px; font-weight: 700;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}

.player-name { font-size: 13px; font-weight: 600; color: var(--text-primary); }

.vnd-value { font-size: 12px; color: var(--text-muted); }
.date-value { font-size: 12px; color: var(--text-muted); }

.actions-cell { display: flex; justify-content: flex-end; gap: 4px; }
</style>
