<template>
  <div class="h-80 w-full">
    <Line v-if="chartData" :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup lang="ts">
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'
import { Line } from 'vue-chartjs'
import type { MonthlyReport } from '~/types/api'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
)

const props = defineProps<{
  monthlyReports: MonthlyReport[]
}>()

const chartData = computed(() => {
  const list = props.monthlyReports?.filter(m => m.max_vdot > 0 || m.avg_vdot > 0) ?? []
  if (list.length === 0) return null
  return {
    labels: list.map(m => formatLabel(m.year_month)),
    datasets: [
      {
        label: '最高VDOT',
        data: list.map(m => m.max_vdot),
        borderColor: 'rgb(59, 130, 246)',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
        tension: 0.2,
        fill: false,
      },
      {
        label: '平均VDOT',
        data: list.map(m => m.avg_vdot),
        borderColor: 'rgb(34, 197, 94)',
        backgroundColor: 'rgba(34, 197, 94, 0.1)',
        tension: 0.2,
        fill: false,
      },
    ],
  }
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'top' as const,
    },
    title: {
      display: false,
    },
  },
  scales: {
    y: {
      beginAtZero: false,
      title: {
        display: true,
        text: 'VDOT',
      },
    },
  },
}

function formatLabel(ym: string): string {
  if (!ym || ym.length < 7) return ym
  const [y, m] = ym.split('-')
  return `${y}/${m}`
}
</script>
