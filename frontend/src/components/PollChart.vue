<template>
  <div class="chart-wrapper">
    <canvas ref="chartCanvas"></canvas>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted } from 'vue'
import {
  Chart,
  BarController,
  BarElement,
  CategoryScale,
  LinearScale,
  Title,
  Tooltip,
  Legend,
  ChartConfiguration
} from 'chart.js'

// 注册Chart.js组件
Chart.register(
  BarController,
  BarElement,
  CategoryScale,
  LinearScale,
  Title,
  Tooltip,
  Legend
)

interface Option {
  id: number
  text: string
  vote_count: number
  poll_id: number
  created_at: string
  updated_at: string
}

interface Poll {
  id: number
  title: string
  description: string
  is_active: boolean
  options: Option[]
  created_at: string
  updated_at: string
}

interface Props {
  pollData: Poll
}

const props = defineProps<Props>()

const chartCanvas = ref<HTMLCanvasElement | null>(null)
let chartInstance: Chart | null = null

const createChart = () => {
  if (!chartCanvas.value || !props.pollData) return

  // 销毁已存在的图表
  if (chartInstance) {
    chartInstance.destroy()
  }

  const ctx = chartCanvas.value.getContext('2d')
  if (!ctx) return

  const labels = props.pollData.options.map(option => option.text)
  const data = props.pollData.options.map(option => option.vote_count)
  const colors = [
    '#FF6384',
    '#36A2EB',
    '#FFCE56',
    '#4BC0C0',
    '#9966FF',
    '#FF9F40'
  ]

  const config: ChartConfiguration = {
    type: 'bar',
    data: {
      labels: labels,
      datasets: [{
        label: '票数',
        data: data,
        backgroundColor: colors.slice(0, data.length),
        borderColor: colors.slice(0, data.length),
        borderWidth: 1
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        y: {
          beginAtZero: true,
          ticks: {
            stepSize: 1
          }
        }
      },
      plugins: {
        title: {
          display: true,
          text: '投票结果统计'
        },
        legend: {
          display: false
        },
        tooltip: {
          callbacks: {
            label: function(context) {
              const total = data.reduce((sum, value) => sum + value, 0)
              const percentage = total > 0 ? ((context.parsed.y / total) * 100).toFixed(1) : '0'
              return `${context.parsed.y} 票 (${percentage}%)`
            }
          }
        }
      },
      animation: {
        duration: 1000,
        easing: 'easeInOutQuart'
      }
    }
  }

  chartInstance = new Chart(ctx, config)
}

const updateChart = () => {
  if (!chartInstance || !props.pollData) return

  const labels = props.pollData.options.map(option => option.text)
  const data = props.pollData.options.map(option => option.vote_count)

  chartInstance.data.labels = labels
  chartInstance.data.datasets[0].data = data
  chartInstance.update('active')
}

// 监听数据变化
watch(() => props.pollData, () => {
  if (chartInstance) {
    updateChart()
  } else {
    createChart()
  }
}, { deep: true })

onMounted(() => {
  createChart()
})

onUnmounted(() => {
  if (chartInstance) {
    chartInstance.destroy()
  }
})
</script>

<style scoped>
.chart-wrapper {
  position: relative;
  height: 400px;
  width: 100%;
}
</style> 