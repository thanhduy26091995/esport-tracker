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
  home: number
  away: number
  other: number
}>()

const chartData = computed(() => ({
  labels: ['Đội nhà', 'Đội khách', 'Khác'],
  datasets: [{
    data: [props.home, props.away, props.other],
    backgroundColor: ['#3b82f6', '#f59e0b', '#6b7280'],
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
