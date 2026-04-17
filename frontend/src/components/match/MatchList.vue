<template>
  <div>
    <!-- Filters -->
    <div class="filter-bar">
      <el-select v-model="typeFilter" :placeholder="t('matches.filterType')" clearable class="w-32">
        <el-option :label="t('matches.allTypes')" value="" />
        <el-option :label="t('matches.types.oneVsOne')" value="1v1" />
        <el-option :label="t('matches.types.twoVsTwo')" value="2v2" />
      </el-select>
      <el-select v-model="dateFilter" :placeholder="t('matches.filterDate')" clearable class="w-40">
        <el-option :label="t('matches.allTime')" value="" />
        <el-option :label="t('common.today')" value="today" />
        <el-option :label="t('common.thisWeek')" value="week" />
        <el-option :label="t('common.thisMonth')" value="month" />
      </el-select>
      <el-select v-model="statusFilter" :placeholder="t('matches.filterStatus')" clearable class="w-36">
        <el-option :label="t('common.all')" value="" />
        <el-option :label="t('common.normal')" value="normal" />
        <el-option :label="t('matches.locked')" value="locked" />
      </el-select>
      <span class="filter-count">{{ t('matches.filterCount', { filtered: filteredMatches.length, total: matches.length }) }}</span>
    </div>

    <div v-if="!loading && filteredMatches.length === 0" class="empty-state">
      <svg class="empty-state-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
      <p class="empty-state-title">{{ hasFilters ? t('matches.noMatchesFound') : t('matches.noMatches') }}</p>
      <p class="empty-state-desc">{{ hasFilters ? t('matches.tryAdjustFilters') : t('matches.startRecording') }}</p>
    </div>

    <div v-else class="match-list">
      <div
        v-for="match in paginatedMatches" :key="match.id"
        class="match-card"
        :class="{ 'match-card--locked': match.is_locked }"
      >
        <!-- Top row -->
        <div class="match-top">
          <div class="flex items-center gap-2">
            <span class="match-type" :class="match.match_type === '1v1' ? 'match-type--1v1' : 'match-type--2v2'">
              {{ getMatchTypeLabel(match.match_type) }}
            </span>
            <span class="match-date">{{ formatDateTime(match.match_date) }}</span>
            <span v-if="match.is_locked" class="match-locked">
              <el-icon :size="11"><Lock /></el-icon> {{ t('matches.locked') }}
            </span>
          </div>
          <el-button v-if="showActions && !match.is_locked" type="danger" size="small" text :icon="Delete" @click="handleDelete(match)">
            {{ t('common.delete') }}
          </el-button>
        </div>

        <!-- Teams -->
        <div class="match-teams">
          <div class="match-team">
            <div class="match-team-label" :class="{ 'match-team-label--win': match.winner_team === 1 }">
              {{ t('matches.team1') }}
              <el-icon v-if="match.winner_team === 1" :size="12"><Trophy /></el-icon>
            </div>
            <div v-for="p in team1(match)" :key="p.id" class="match-player" :class="{ 'match-player--win': match.winner_team === 1 }">
              <span class="match-player-name">{{ p.user.name }}</span>
              <span class="match-pts" :class="p.point_change >= 0 ? 'match-pts--pos' : 'match-pts--neg'">
                {{ p.point_change > 0 ? '+' : '' }}{{ p.point_change }}
              </span>
            </div>
          </div>
          <div class="match-vs">{{ t('common.vs') }}</div>
          <div class="match-team match-team--right">
            <div class="match-team-label" :class="{ 'match-team-label--win': match.winner_team === 2 }">
              <el-icon v-if="match.winner_team === 2" :size="12"><Trophy /></el-icon>
              {{ t('matches.team2') }}
            </div>
            <div v-for="p in team2(match)" :key="p.id" class="match-player match-player--right" :class="{ 'match-player--win': match.winner_team === 2 }">
              <span class="match-pts" :class="p.point_change >= 0 ? 'match-pts--pos' : 'match-pts--neg'">
                {{ p.point_change > 0 ? '+' : '' }}{{ p.point_change }}
              </span>
              <span class="match-player-name">{{ p.user.name }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="filteredMatches.length > pageSize" class="mt-6 flex justify-center">
      <el-pagination v-model:current-page="currentPage" :page-size="pageSize" :total="filteredMatches.length" layout="prev, pager, next" @current-change="handlePageChange" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
import { Delete, Lock, Trophy } from '@element-plus/icons-vue'
import type { Match, MatchParticipant } from '@/types/match'
import { formatDateTime, isToday } from '@/utils/date'
import { getMatchTypeLabel } from '@/utils/tournamentLabels'

interface Props { matches: Match[]; loading?: boolean; showActions?: boolean }
const props = withDefaults(defineProps<Props>(), { loading: false, showActions: true })
const emit = defineEmits<{ delete: [match: Match] }>()

const typeFilter = ref(''); const dateFilter = ref(''); const statusFilter = ref('')
const currentPage = ref(1); const pageSize = 20

const team1 = (m: Match): MatchParticipant[] => m.participants.filter(p => p.team_number === 1)
const team2 = (m: Match): MatchParticipant[] => m.participants.filter(p => p.team_number === 2)

const hasFilters = computed(() => typeFilter.value || dateFilter.value || statusFilter.value)

const filteredMatches = computed(() => {
  let r = props.matches
  if (typeFilter.value) r = r.filter(m => m.match_type === typeFilter.value)
  if (dateFilter.value) {
    const now = new Date()
    r = r.filter(m => {
      const d = new Date(m.match_date)
      if (dateFilter.value === 'today') return isToday(m.match_date)
      if (dateFilter.value === 'week') return d >= new Date(now.getTime() - 7 * 86400000)
      if (dateFilter.value === 'month') return d >= new Date(now.getTime() - 30 * 86400000)
      return true
    })
  }
  if (statusFilter.value === 'locked') r = r.filter(m => m.is_locked)
  else if (statusFilter.value === 'normal') r = r.filter(m => !m.is_locked)
  return r
})

const paginatedMatches = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredMatches.value.slice(start, start + pageSize)
})

const handleDelete = (match: Match) => emit('delete', match)
const handlePageChange = (page: number) => { currentPage.value = page; window.scrollTo({ top: 0, behavior: 'smooth' }) }
</script>

<style scoped>
.match-list { display: flex; flex-direction: column; gap: 10px; }

.match-card {
  background: var(--surface-card);
  border: 1px solid var(--border-default);
  border-radius: 14px;
  padding: 16px;
  box-shadow: var(--shadow-card);
  transition: box-shadow 0.15s;
}
.match-card:hover { box-shadow: var(--shadow-card-hover); }
.match-card--locked { background: #fffbeb; border-color: #fde68a; }

.match-top {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 14px;
}

.match-type {
  font-size: 10px; font-weight: 800; letter-spacing: 0.05em;
  padding: 3px 8px; border-radius: 20px;
}
.match-type--1v1 { background: var(--color-info-bg); color: var(--color-info); }
.match-type--2v2 { background: var(--color-success-bg); color: var(--color-success); }

.match-date { font-size: 12px; color: var(--text-muted); }
.match-locked { font-size: 11px; color: var(--color-warning); display: flex; align-items: center; gap: 3px; font-weight: 600; }

.match-teams { display: grid; grid-template-columns: minmax(0, 1fr) 40px minmax(0, 1fr); gap: 12px; align-items: start; }

.match-team { display: flex; flex-direction: column; gap: 6px; min-width: 0; }
.match-team--right { align-items: stretch; }
.match-team--right .match-team-label { justify-content: flex-end; }

.match-team-label {
  font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em;
  color: var(--text-muted); padding-bottom: 6px; border-bottom: 2px solid var(--border-default);
  display: flex; align-items: center; gap: 4px;
}
.match-team-label--win { color: var(--color-success); border-bottom-color: var(--color-success); }

.match-player {
  display: flex; align-items: center; justify-content: space-between;
  padding: 5px 10px; border-radius: 8px; background: var(--surface-page);
  font-size: 12px; color: var(--text-secondary); gap: 6px; width: 100%; min-width: 0;
}
.match-player--win { background: var(--color-success-bg); color: var(--color-success); font-weight: 600; }
.match-player--right { flex-direction: row-reverse; }

.match-player-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.match-pts { font-size: 11px; font-weight: 700; }
.match-pts--pos { color: var(--color-success); }
.match-pts--neg { color: var(--color-danger); }

.match-vs { display: flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 900; color: var(--border-default); padding-top: 22px; }
</style>
