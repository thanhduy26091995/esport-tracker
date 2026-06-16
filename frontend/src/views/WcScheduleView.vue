<template>
  <div class="page-wrapper">
    <div class="page-container">
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title wc-page-title">
            🏆 World Cup 2026
          </h1>
          <p class="page-subtitle">{{ t('wc.schedule') }}</p>
        </div>
        <div class="wc-schedule-header-right">
          <router-link
            v-if="wcAuthStore.isLoggedIn && wcAuthStore.isAdmin"
            :to="{ name: 'wc-admin' }"
            class="wc-admin-link"
          >
            {{ t('wc.adminPanel') }}
          </router-link>
          <router-link
            v-else-if="wcAuthStore.isLoggedIn && !wcAuthStore.isAdmin"
            :to="{ name: 'wc-predict' }"
            class="wc-predict-link"
          >
            {{ t('wc.predicting') }}
          </router-link>
          <router-link
            v-else-if="!wcAuthStore.isLoggedIn"
            :to="{ name: 'wc-login' }"
            class="wc-login-link"
          >
            {{ t('wc.login') }}
          </router-link>
        </div>
      </div>

      <WcGroupFilter v-model="selectedFilter" />

      <div v-if="store.loading" class="wc-loading">
        <el-skeleton :rows="5" animated />
      </div>

      <template v-else-if="groupedMatches.length > 0">
        <div v-for="group in groupedMatches" :key="group.date" class="wc-date-group">
          <div class="wc-date-heading">{{ group.dateLabel }}</div>
          <div class="wc-match-list">
            <WcMatchCard
              v-for="match in group.matches"
              :key="match.id"
              :match="match"
            />
          </div>
        </div>
      </template>

      <div v-else class="empty-state">
        <div class="empty-state-icon">🏟️</div>
        <div class="empty-state-title">{{ t('wc.noMatches') }}</div>
        <div class="empty-state-desc">{{ t('common.empty') }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useWcStore } from '@/stores/wcStore'
import { useWcAuthStore } from '@/stores/wcAuthStore'
import WcGroupFilter from '@/components/wc/WcGroupFilter.vue'
import WcMatchCard from '@/components/wc/WcMatchCard.vue'
import type { WcMatch } from '@/types/wc'

const { t } = useI18n()
const store = useWcStore()
const wcAuthStore = useWcAuthStore()

const selectedFilter = ref('')

const filteredMatches = computed(() => {
  if (!selectedFilter.value) return store.matches
  // e.g. 'group_A' → stage=group, group=A; 'r16' → stage=r16
  if (selectedFilter.value.startsWith('group_')) {
    const g = selectedFilter.value.replace('group_', 'Group ')
    return store.matches.filter(m => m.stage === 'group' && m.group_name === g)
  }
  return store.matches.filter(m => m.stage === selectedFilter.value)
})

interface DateGroup { date: string; dateLabel: string; matches: WcMatch[] }

const groupedMatches = computed((): DateGroup[] => {
  const map = new Map<string, WcMatch[]>()
  for (const m of filteredMatches.value) {
    const d = m.match_date.slice(0, 10)
    if (!map.has(d)) map.set(d, [])
    map.get(d)!.push(m)
  }
  const groups: DateGroup[] = []
  for (const [date, matches] of map) {
    const dt = new Date(date)
    const dateLabel = dt.toLocaleDateString('vi-VN', {
      weekday: 'long', day: 'numeric', month: 'long', year: 'numeric',
    })
    groups.push({ date, dateLabel, matches })
  }
  return groups.sort((a, b) => a.date.localeCompare(b.date))
})

watch(selectedFilter, () => {
  const filter: Record<string, string> = {}
  if (selectedFilter.value.startsWith('group_')) {
    filter.stage = 'group'
    filter.group = selectedFilter.value.replace('group_', 'Group ')
  } else if (selectedFilter.value) {
    filter.stage = selectedFilter.value
  }
  store.fetchMatches(filter)
})

onMounted(() => store.fetchMatches())
</script>

<style scoped>
.wc-page-title {
  color: #16a34a;
}

.wc-schedule-header-right {
  display: flex;
  align-items: center;
}

.wc-admin-link {
  font-size: 13px;
  font-weight: 600;
  color: #ef4444;
  background: rgba(239, 68, 68, 0.08);
  border: 1px solid rgba(239, 68, 68, 0.2);
  padding: 6px 14px;
  border-radius: 8px;
  text-decoration: none;
  transition: background 0.15s;
}

.wc-admin-link:hover {
  background: rgba(239, 68, 68, 0.14);
}

.wc-login-link {
  font-size: 13px;
  font-weight: 600;
  color: #16a34a;
  background: rgba(22, 163, 74, 0.08);
  border: 1px solid rgba(22, 163, 74, 0.2);
  padding: 6px 14px;
  border-radius: 8px;
  text-decoration: none;
  transition: background 0.15s;
}

.wc-login-link:hover {
  background: rgba(22, 163, 74, 0.14);
}

.wc-predict-link {
  font-size: 13px;
  font-weight: 600;
  color: #2563eb;
  background: rgba(37, 99, 235, 0.08);
  border: 1px solid rgba(37, 99, 235, 0.2);
  padding: 6px 14px;
  border-radius: 8px;
  text-decoration: none;
  transition: background 0.15s;
}

.wc-predict-link:hover {
  background: rgba(37, 99, 235, 0.14);
}

.wc-loading {
  margin-top: 16px;
}

.wc-date-group {
  margin-bottom: 24px;
}

.wc-date-heading {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 0 4px 8px;
  border-bottom: 1px solid var(--border-subtle);
  margin-bottom: 10px;
}

.wc-match-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
</style>
