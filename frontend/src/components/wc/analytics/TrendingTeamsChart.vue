<template>
  <div class="chart-wrapper">
    <Bar :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Bar } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, BarElement, Tooltip } from 'chart.js'
import type { TeamCountEntry } from '@/types/wc'

ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip)

const props = defineProps<{ teams: TeamCountEntry[] }>()

const chartData = computed(() => ({
  labels: props.teams.map(t => t.team),
  datasets: [{
    data: props.teams.map(t => t.bet_count),
    backgroundColor: 'rgba(22,163,74,0.75)',
    borderRadius: 4,
  }],
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  indexAxis: 'y' as const,
  plugins: { legend: { display: false } },
  scales: {
    x: { beginAtZero: true, ticks: { precision: 0 } },
  },
}
</script>

<style scoped>
.chart-wrapper {
  position: relative;
  height: 240px;
  width: 100%;
}
</style>
