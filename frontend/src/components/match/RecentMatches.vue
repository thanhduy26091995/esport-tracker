<template>
  <div>
    <div v-if="loading" class="rm-loading">
      <el-icon class="animate-spin" :size="24" style="color:var(--text-muted)"><Loading /></el-icon>
    </div>
    <div v-else-if="displayMatches.length === 0" class="rm-empty">No recent matches</div>
    <div v-else class="rm-list">
      <div
        v-for="match in displayMatches" :key="match.id"
        class="rm-card"
        @click="$emit('matchClick', match)"
      >
        <div class="rm-top">
          <div class="flex items-center gap-2">
            <span class="rm-type" :class="match.match_type === '1v1' ? 'rm-type--1v1' : 'rm-type--2v2'">
              {{ match.match_type }}
            </span>
            <span class="rm-time">{{ formatRelativeTime(match.match_date) }}</span>
          </div>
          <span v-if="match.is_locked" class="rm-locked">
            <el-icon :size="10"><Lock /></el-icon> Locked
          </span>
        </div>
        <div class="rm-teams">
          <div class="rm-team">
            <span v-for="p in team1(match)" :key="p.id" class="rm-player" :class="{ 'rm-player--win': match.winner_team === 1 }">
              {{ p.user.name }}
              <span class="rm-delta" :class="p.point_change >= 0 ? 'rm-delta--pos' : 'rm-delta--neg'">
                {{ p.point_change > 0 ? '+' : '' }}{{ p.point_change }}
              </span>
            </span>
          </div>
          <div class="rm-vs">VS</div>
          <div class="rm-team rm-team--right">
            <span v-for="p in team2(match)" :key="p.id" class="rm-player" :class="{ 'rm-player--win': match.winner_team === 2 }">
              {{ p.user.name }}
              <span class="rm-delta" :class="p.point_change >= 0 ? 'rm-delta--pos' : 'rm-delta--neg'">
                {{ p.point_change > 0 ? '+' : '' }}{{ p.point_change }}
              </span>
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Loading, Lock } from '@element-plus/icons-vue'
import type { Match, MatchParticipant } from '@/types/match'
import { formatRelativeTime } from '@/utils/date'

interface Props { matches: Match[]; loading?: boolean; limit?: number }
const props = withDefaults(defineProps<Props>(), { loading: false, limit: 5 })
defineEmits<{ viewAll: []; matchClick: [match: Match] }>()

const displayMatches = computed(() => props.matches.slice(0, props.limit))
const team1 = (m: Match): MatchParticipant[] => m.participants.filter(p => p.team_number === 1)
const team2 = (m: Match): MatchParticipant[] => m.participants.filter(p => p.team_number === 2)
</script>

<style scoped>
.rm-loading, .rm-empty {
  display: flex; justify-content: center; align-items: center;
  padding: 40px 0; font-size: 13px; color: var(--text-muted);
}

.rm-list { display: flex; flex-direction: column; gap: 8px; }

.rm-card {
  border-radius: 12px;
  border: 1px solid var(--border-subtle);
  padding: 12px;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
  background: var(--surface-card);
}
.rm-card:hover { background: var(--surface-page); border-color: var(--border-default); }

.rm-top {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 10px;
}

.rm-type {
  font-size: 10px; font-weight: 800; letter-spacing: 0.04em;
  padding: 2px 8px; border-radius: 20px;
}
.rm-type--1v1 { background: var(--color-info-bg); color: var(--color-info); }
.rm-type--2v2 { background: var(--color-success-bg); color: var(--color-success); }

.rm-time { font-size: 11px; color: var(--text-muted); }

.rm-locked { font-size: 11px; color: var(--color-warning); display: flex; align-items: center; gap: 3px; }

.rm-teams { display: grid; grid-template-columns: 1fr auto 1fr; gap: 8px; align-items: center; }

.rm-team { display: flex; flex-direction: column; gap: 4px; }
.rm-team--right { align-items: flex-end; text-align: right; }

.rm-player { font-size: 12px; font-weight: 500; color: var(--text-secondary); display: flex; align-items: center; gap: 4px; }
.rm-player--win { color: var(--color-success); font-weight: 700; }
.rm-team--right .rm-player { flex-direction: row-reverse; }

.rm-delta { font-size: 10px; font-weight: 700; }
.rm-delta--pos { color: var(--color-success); }
.rm-delta--neg { color: var(--color-danger); }

.rm-vs { font-size: 11px; font-weight: 900; color: var(--border-default); text-align: center; }
</style>
