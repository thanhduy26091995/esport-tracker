<template>
  <div class="chart-wrapper">
    <Doughnut :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Doughnut } from 'vue-chartjs'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'

ChartJS.register(ArcElement, Tooltip, Legend)

const props = defineProps<{
  handicap: number
  exactScore: number
  overUnder: number
  custom: number
}>()

const chartData = computed(() => ({
  labels: ['Handicap', 'Tỷ số', 'Tài/Xỉu', 'Kèo phụ'],
  datasets: [{
    data: [props.handicap, props.exactScore, props.overUnder, props.custom],
    backgroundColor: ['#3b82f6', '#f59e0b', '#8b5cf6', '#10b981'],
    borderWidth: 0,
  }],
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'bottom' as const },
  },
  cutout: '65%',
}
</script>

<style scoped>
.chart-wrapper {
  position: relative;
  height: 180px;
  width: 100%;
}
</style>
