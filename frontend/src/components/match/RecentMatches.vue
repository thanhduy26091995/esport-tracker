<template>
  <div>
    <div v-if="loading" class="rm-loading">
      <el-icon class="animate-spin" :size="24" style="color:var(--text-muted)"><Loading /></el-icon>
    </div>
    <div v-else-if="displayMatches.length === 0" class="rm-empty">{{ t('matches.noRecent') }}</div>
    <div v-else class="rm-list">
      <template v-for="item in displayMatches" :key="item.id">
        <!-- Bonus row -->
        <div v-if="item.type === 'bonus'" class="rm-card rm-card--bonus">
          <div class="rm-top">
            <div class="flex items-center gap-2">
              <span class="rm-type rm-type--bonus">{{ t('matches.bonus.tag') }}</span>
              <span class="rm-time">{{ formatRelativeTime(item.bonus_date ?? '') }}</span>
            </div>
          </div>
          <div class="rm-bonus-row">
            <span class="rm-bonus-player">{{ item.user?.name ?? '—' }}</span>
            <span class="rm-bonus-pts">+{{ item.points }}</span>
            <span v-if="item.description" class="rm-bonus-desc">{{ item.description }}</span>
          </div>
        </div>

        <!-- Match row -->
        <div v-else class="rm-card" @click="$emit('matchClick', item)">
          <div class="rm-top">
            <div class="flex items-center gap-2">
              <span class="rm-type" :class="{
                'rm-type--1v1': item.match_type === '1v1',
                'rm-type--2v2': item.match_type === '2v2',
                'rm-type--1v2': item.match_type === '1v2',
              }">
                {{ getMatchTypeLabel(item.match_type ?? '1v1') }}
              </span>
              <span class="rm-time">{{ formatRelativeTime(item.match_date ?? '') }}</span>
            </div>
            <span v-if="item.is_locked" class="rm-locked">
              <el-icon :size="10"><Lock /></el-icon> {{ t('matches.locked') }}
            </span>
          </div>
          <div class="rm-teams">
            <div class="rm-team">
              <span v-for="p in team1(item)" :key="p.id" class="rm-player" :class="{ 'rm-player--win': item.winner_team === 1 }">
                {{ p.user.name }}
                <span class="rm-delta" :class="p.point_change >= 0 ? 'rm-delta--pos' : 'rm-delta--neg'">
                  {{ p.point_change > 0 ? '+' : '' }}{{ p.point_change }}
                </span>
              </span>
            </div>
            <div class="rm-vs">{{ t('common.vs') }}</div>
            <div class="rm-team rm-team--right">
              <span v-for="p in team2(item)" :key="p.id" class="rm-player" :class="{ 'rm-player--win': item.winner_team === 2 }">
                {{ p.user.name }}
                <span class="rm-delta" :class="p.point_change >= 0 ? 'rm-delta--pos' : 'rm-delta--neg'">
                  {{ p.point_change > 0 ? '+' : '' }}{{ p.point_change }}
                </span>
              </span>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
import { Loading, Lock } from '@element-plus/icons-vue'
import type { MatchFeedItem, MatchParticipant } from '@/types/match'
import { formatRelativeTime } from '@/utils/date'
import { getMatchTypeLabel } from '@/utils/tournamentLabels'

interface Props { matches: MatchFeedItem[]; loading?: boolean; limit?: number }
const props = withDefaults(defineProps<Props>(), { loading: false, limit: 5 })
defineEmits<{ viewAll: []; matchClick: [item: MatchFeedItem] }>()

const displayMatches = computed(() => props.matches.slice(0, props.limit))
const team1 = (m: MatchFeedItem): MatchParticipant[] => (m.participants ?? []).filter(p => p.team_number === 1)
const team2 = (m: MatchFeedItem): MatchParticipant[] => (m.participants ?? []).filter(p => p.team_number === 2)
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
.rm-type--1v2 { background: var(--color-warning-bg, #fff7ed); color: var(--color-warning, #d97706); }
.rm-type--bonus { background: #fdf4ff; color: #9333ea; }

.rm-card--bonus { background: #fdf4ff; border-color: #e9d5ff; cursor: default; }
.rm-bonus-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; font-size: 12px; }
.rm-bonus-player { font-weight: 700; color: var(--text-primary); }
.rm-bonus-pts { font-weight: 800; color: #9333ea; }
.rm-bonus-desc { color: var(--text-muted); flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

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
