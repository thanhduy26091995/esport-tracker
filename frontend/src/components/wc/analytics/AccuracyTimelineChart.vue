<template>
  <div class="chart-wrapper">
    <Line :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  LineElement,
  PointElement,
  Title,
  Tooltip,
  Legend,
  Filler,
  type ChartOptions,
} from 'chart.js'
import type { AnalyticsTimelinePoint } from '@/types/wc'

ChartJS.register(CategoryScale, LinearScale, LineElement, PointElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{ points: AnalyticsTimelinePoint[] }>()

const chartData = computed(() => ({
  labels: props.points.map(p => p.period),
  datasets: [
    {
      label: 'Accuracy %',
      data: props.points.map(p => Math.round(p.accuracy * 100)),
      borderColor: '#16a34a',
      backgroundColor: 'rgba(22,163,74,0.1)',
      tension: 0.3,
      fill: true,
      pointRadius: 4,
      pointBackgroundColor: '#16a34a',
    },
  ],
}))

const chartOptions: ChartOptions<'line'> = {
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    y: {
      min: 0,
      max: 100,
      ticks: { callback: (v: number | string) => `${v}%` },
    },
  },
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (ctx) => `${ctx.raw as number}%`,
      },
    },
  },
}
</script>

<style scoped>
.chart-wrapper {
  position: relative;
  height: 200px;
  width: 100%;
}
</style>
