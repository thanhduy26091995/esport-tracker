<template>
  <div class="stat-card" :class="`stat-card--${type}`">
    <div class="stat-icon-wrap">
      <el-icon :size="20" class="stat-icon">
        <component :is="icon" />
      </el-icon>
    </div>
    <div class="stat-body">
      <div class="stat-label">{{ title }}</div>
      <div class="stat-value">
        <span v-if="loading" class="stat-skeleton" />
        <template v-else>{{ displayValue }}</template>
      </div>
      <div v-if="trend && !loading" class="stat-trend" :class="trend > 0 ? 'stat-trend--up' : 'stat-trend--down'">
        <el-icon :size="11"><component :is="trend > 0 ? ArrowUp : ArrowDown" /></el-icon>
        {{ Math.abs(trend) }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ArrowUp, ArrowDown, User } from '@element-plus/icons-vue'
import { formatNumber } from '@/utils/formatters'

interface Props {
  title: string
  value: number | string
  icon?: any
  trend?: number
  loading?: boolean
  type?: 'default' | 'success' | 'warning' | 'danger' | 'info'
}

const props = withDefaults(defineProps<Props>(), {
  icon: User,
  loading: false,
  type: 'default'
})

const displayValue = computed(() =>
  typeof props.value === 'number' ? formatNumber(props.value) : props.value
)
</script>

<style scoped>
.stat-card {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 20px;
  background: var(--surface-card);
  border-radius: 16px;
  border: 1px solid var(--border-default);
  box-shadow: var(--shadow-card);
  transition: box-shadow 0.2s ease, transform 0.2s ease;
}

.stat-card:hover {
  box-shadow: var(--shadow-card-hover);
  transform: translateY(-1px);
}

/* Icon wrapper */
.stat-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon {
  font-size: 20px;
}

/* Type-based icon colors */
.stat-card--info .stat-icon-wrap {
  background: linear-gradient(135deg, #dbeafe, #bfdbfe);
  color: #1d4ed8;
}

.stat-card--success .stat-icon-wrap {
  background: linear-gradient(135deg, #dcfce7, #bbf7d0);
  color: #15803d;
}

.stat-card--warning .stat-icon-wrap {
  background: linear-gradient(135deg, #fef9c3, #fde68a);
  color: #b45309;
}

.stat-card--danger .stat-icon-wrap {
  background: linear-gradient(135deg, #fee2e2, #fecaca);
  color: #b91c1c;
}

.stat-card--default .stat-icon-wrap {
  background: linear-gradient(135deg, #f1f5f9, #e2e8f0);
  color: #475569;
}

/* Body */
.stat-body {
  flex: 1;
  min-width: 0;
}

.stat-label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 6px;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  line-height: 1.1;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.stat-skeleton {
  display: inline-block;
  width: 80px;
  height: 28px;
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 6px;
  vertical-align: middle;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* Trend */
.stat-trend {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: 11px;
  font-weight: 600;
  margin-top: 4px;
}

.stat-trend--up { color: var(--color-success); }
.stat-trend--down { color: var(--color-danger); }
</style>
