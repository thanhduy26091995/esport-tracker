<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useWcAuthStore } from '@/stores/wcAuthStore'
import type { WcMatch, WcStage } from '@/types/wc'

defineProps<{
  matches: WcMatch[]
  hasMore: boolean
}>()

const { t } = useI18n()
const router = useRouter()
const wcAuthStore = useWcAuthStore()

const STAGE_LABELS: Record<WcStage, string> = {
  group: 'Vòng bảng',
  r32: 'Vòng 32',
  r16: 'Vòng 16',
  qf: 'Tứ kết',
  sf: 'Bán kết',
  final: 'Chung kết',
  third_place: 'Tranh hạng 3',
}

function stageLabel(stage: WcStage, groupName?: string): string {
  return groupName || STAGE_LABELS[stage] || stage
}

function formatMatchTime(iso: string): string {
  return new Date(iso).toLocaleString('vi-VN', {
    timeZone: 'Asia/Ho_Chi_Minh',
    weekday: 'short',
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function goToSchedule() {
  router.push('/asean-cup')
}
</script>

<template>
  <div class="wc-upcoming-widget">
    <div class="wc-upcoming-header">
      <span class="wc-upcoming-title">{{ t('dashboard.aseanUpcomingTitle') }}</span>
      <div class="wc-upcoming-header-actions">
        <router-link
          v-if="!wcAuthStore.isAdmin"
          :to="wcAuthStore.isLoggedIn ? '/asean-cup/predict' : '/asean-cup/login'"
          class="predict-shortcut-link"
        >
          🎯 Dự đoán
        </router-link>
        <router-link to="/asean-cup" class="view-all-link">Xem lịch đầy đủ →</router-link>
      </div>
    </div>
    <div class="wc-upcoming-list">
      <div
        v-for="m in matches"
        :key="m.id"
        class="wc-upcoming-card"
        :class="{ 'wc-upcoming-card--live': m.status === 'live' }"
        @click="goToSchedule"
      >
        <div class="wc-upcoming-meta">{{ stageLabel(m.stage, m.group_name) }}</div>
        <div class="wc-upcoming-teams">
          <span class="team">{{ m.home_team }}</span>
          <span class="vs">vs</span>
          <span class="team">{{ m.away_team }}</span>
        </div>
        <div class="wc-upcoming-time" :class="{ live: m.status === 'live' }">
          {{ m.status === 'live' ? '🔴 LIVE' : formatMatchTime(m.match_date) }}
        </div>
      </div>
      <div
        v-if="hasMore"
        class="wc-upcoming-card wc-upcoming-card--more"
        @click="goToSchedule"
      >
        <span>Xem thêm →</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.wc-upcoming-widget {
  background: var(--surface-card, #1e2130);
  border-radius: 12px;
  padding: 12px 16px;
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.08));
}

.wc-upcoming-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.wc-upcoming-title {
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text-primary, #e2e8f0);
}

.wc-upcoming-header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.view-all-link {
  font-size: 0.78rem;
  color: var(--color-primary, #4f8ef7);
  text-decoration: none;
  white-space: nowrap;
}

.view-all-link:hover {
  text-decoration: underline;
}

.predict-shortcut-link {
  font-size: 0.75rem;
  font-weight: 600;
  color: #2563eb;
  background: rgba(37, 99, 235, 0.1);
  border: 1px solid rgba(37, 99, 235, 0.25);
  padding: 3px 10px;
  border-radius: 6px;
  text-decoration: none;
  white-space: nowrap;
  transition: background 0.15s;
}

.predict-shortcut-link:hover {
  background: rgba(37, 99, 235, 0.18);
}

.wc-upcoming-list {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding-bottom: 4px;
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.15) transparent;
}

.wc-upcoming-list::-webkit-scrollbar {
  height: 4px;
}

.wc-upcoming-list::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);
  border-radius: 2px;
}

.wc-upcoming-card {
  flex: 0 0 auto;
  min-width: 148px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.08));
  border-radius: 8px;
  padding: 10px 12px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 5px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.wc-upcoming-card:hover {
  border-color: rgba(79, 142, 247, 0.45);
  box-shadow: 0 0 0 1px rgba(79, 142, 247, 0.2);
}

.wc-upcoming-card--live {
  border-color: #16a34a;
  background: linear-gradient(135deg, rgba(22, 163, 74, 0.06), rgba(255, 255, 255, 0.03));
  animation: glow-live 2s ease-in-out infinite alternate;
}

@keyframes glow-live {
  from { box-shadow: 0 0 0 2px rgba(22, 163, 74, 0.20), 0 0 8px  rgba(22, 163, 74, 0.12); }
  to   { box-shadow: 0 0 0 2px rgba(22, 163, 74, 0.40), 0 0 20px rgba(22, 163, 74, 0.28); }
}

.wc-upcoming-card--more {
  min-width: 96px;
  justify-content: center;
  align-items: center;
  color: var(--color-primary, #4f8ef7);
  font-size: 0.82rem;
  font-weight: 500;
  border-style: dashed;
}

.wc-upcoming-meta {
  font-size: 0.68rem;
  color: var(--text-muted, #94a3b8);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.wc-upcoming-teams {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--text-primary, #e2e8f0);
}

.vs {
  color: var(--text-muted, #94a3b8);
  font-weight: 400;
  font-size: 0.72rem;
}

.wc-upcoming-time {
  font-size: 0.75rem;
  color: var(--text-secondary, #cbd5e1);
  white-space: nowrap;
}

.wc-upcoming-time.live {
  color: #22c55e;
  font-weight: 600;
}
</style>
