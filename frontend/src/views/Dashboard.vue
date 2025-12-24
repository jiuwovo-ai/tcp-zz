<template>
  <div class="min-h-screen" :class="settingsStore.isDark ? 'bg-dark-bg' : 'bg-gray-50'">
    <!-- Header -->
    <header class="border-b" :class="settingsStore.isDark ? 'border-dark-border bg-dark-card' : 'border-gray-200 bg-white'">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <div class="flex items-center space-x-4">
            <h1 class="text-xl font-bold" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ t('dashboard.title') }}</h1>
            <span
              class="status-dot"
              :class="wsConnected ? 'status-normal' : 'status-error'"
            ></span>
          </div>
          <div class="flex items-center space-x-4">
            <router-link to="/nodes">
              <n-button quaternary>{{ t('nodes.manageNodes') }}</n-button>
            </router-link>
            <router-link to="/tunnels">
              <n-button quaternary>{{ t('dashboard.manageTunnels') }}</n-button>
            </router-link>
            <SettingsDropdown />
            <n-button quaternary @click="handleLogout">{{ t('common.logout') }}</n-button>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Stats Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div class="card">
          <div class="flex items-center justify-between">
            <div>
              <p class="stat-label">{{ t('dashboard.totalUpload') }}</p>
              <p class="stat-value text-blue-400">{{ formatBytes(store.global.total_out) }}</p>
              <p class="text-sm text-gray-500 mt-1">{{ formatBytesRate(store.global.rate_out) }}</p>
            </div>
            <n-icon size="40" class="text-blue-400 opacity-50">
              <ArrowUp />
            </n-icon>
          </div>
        </div>

        <div class="card">
          <div class="flex items-center justify-between">
            <div>
              <p class="stat-label">{{ t('dashboard.totalDownload') }}</p>
              <p class="stat-value text-green-400">{{ formatBytes(store.global.total_in) }}</p>
              <p class="text-sm text-gray-500 mt-1">{{ formatBytesRate(store.global.rate_in) }}</p>
            </div>
            <n-icon size="40" class="text-green-400 opacity-50">
              <ArrowDown />
            </n-icon>
          </div>
        </div>

        <div class="card">
          <div class="flex items-center justify-between">
            <div>
              <p class="stat-label">{{ t('dashboard.activeTunnels') }}</p>
              <p class="stat-value text-purple-400">{{ store.activeTunnelCount }}</p>
              <p class="text-sm text-gray-500 mt-1">{{ store.tunnels.length }} {{ t('dashboard.total') }}</p>
            </div>
            <n-icon size="40" class="text-purple-400 opacity-50">
              <GitNetwork />
            </n-icon>
          </div>
        </div>

        <div class="card">
          <div class="flex items-center justify-between">
            <div>
              <p class="stat-label">{{ t('dashboard.uptime') }}</p>
              <p class="stat-value text-yellow-400">{{ formatUptime(store.system.uptime) }}</p>
              <p class="text-sm text-gray-500 mt-1">{{ store.totalConnections }} {{ t('dashboard.connections') }}</p>
            </div>
            <n-icon size="40" class="text-yellow-400 opacity-50">
              <Time />
            </n-icon>
          </div>
        </div>
      </div>

      <!-- System Stats & Traffic Chart -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        <!-- System Stats -->
        <div class="card">
          <h3 class="text-lg font-semibold mb-4" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ t('dashboard.systemStatus') }}</h3>
          <div class="space-y-4">
            <div>
              <div class="flex justify-between text-sm mb-1">
                <span class="text-gray-400">{{ t('dashboard.cpu') }}</span>
                <span :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ formatPercent(store.system.cpu_percent) }}</span>
              </div>
              <n-progress
                type="line"
                :percentage="store.system.cpu_percent"
                :show-indicator="false"
                :height="8"
                :border-radius="4"
                :color="getProgressColor(store.system.cpu_percent)"
              />
            </div>
            <div>
              <div class="flex justify-between text-sm mb-1">
                <span class="text-gray-400">{{ t('dashboard.memory') }}</span>
                <span :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ formatPercent(store.system.memory_percent) }}</span>
              </div>
              <n-progress
                type="line"
                :percentage="store.system.memory_percent"
                :show-indicator="false"
                :height="8"
                :border-radius="4"
                :color="getProgressColor(store.system.memory_percent)"
              />
            </div>
            <div class="pt-2 border-t" :class="settingsStore.isDark ? 'border-dark-border' : 'border-gray-200'">
              <div class="flex justify-between text-sm">
                <span class="text-gray-400">{{ t('dashboard.networkIn') }}</span>
                <span class="text-green-400">{{ formatBytesRate(store.system.net_rate_in) }}</span>
              </div>
              <div class="flex justify-between text-sm mt-2">
                <span class="text-gray-400">{{ t('dashboard.networkOut') }}</span>
                <span class="text-blue-400">{{ formatBytesRate(store.system.net_rate_out) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Traffic Chart -->
        <div class="card lg:col-span-2">
          <h3 class="text-lg font-semibold mb-4" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ t('dashboard.realtimeTraffic') }}</h3>
          <div ref="chartRef" class="h-64"></div>
        </div>
      </div>

      <!-- Tunnel List -->
      <div class="card">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ t('dashboard.tunnels') }}</h3>
          <router-link to="/tunnels">
            <n-button size="small" type="primary">{{ t('dashboard.manage') }}</n-button>
          </router-link>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="text-left text-gray-400 text-sm border-b" :class="settingsStore.isDark ? 'border-dark-border' : 'border-gray-200'">
                <th class="pb-3 font-medium">{{ t('common.name') }}</th>
                <th class="pb-3 font-medium">{{ t('dashboard.localPort') }}</th>
                <th class="pb-3 font-medium">{{ t('dashboard.target') }}</th>
                <th class="pb-3 font-medium">{{ t('dashboard.protocol') }}</th>
                <th class="pb-3 font-medium">{{ t('dashboard.traffic') }}</th>
                <th class="pb-3 font-medium">{{ t('dashboard.latency') }}</th>
                <th class="pb-3 font-medium">{{ t('common.status') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="tunnel in store.tunnels"
                :key="tunnel.rule.id"
                class="border-b transition-colors"
                :class="settingsStore.isDark ? 'border-dark-border hover:bg-dark-hover' : 'border-gray-100 hover:bg-gray-50'"
              >
                <td class="py-3" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ tunnel.rule.name }}</td>
                <td class="py-3" :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">{{ tunnel.rule.local_port }}</td>
                <td class="py-3" :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">{{ tunnel.rule.target_ip }}:{{ tunnel.rule.target_port }}</td>
                <td class="py-3">
                  <n-tag :type="tunnel.rule.protocol === 'tcp' ? 'info' : 'warning'" size="small">
                    {{ tunnel.rule.protocol.toUpperCase() }}
                  </n-tag>
                </td>
                <td class="py-3" :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">
                  <span class="text-blue-400">↑{{ formatBytesRate(tunnel.traffic.bytes_out_rate) }}</span>
                  <span class="mx-1 text-gray-500">/</span>
                  <span class="text-green-400">↓{{ formatBytesRate(tunnel.traffic.bytes_in_rate) }}</span>
                </td>
                <td class="py-3">
                  <span :class="getLatencyClass(tunnel.latency)">
                    {{ formatLatency(tunnel.latency.latency) }}
                  </span>
                </td>
                <td class="py-3">
                  <div class="flex items-center space-x-2">
                    <span
                      class="status-dot"
                      :class="tunnel.running ? 'status-normal' : 'status-error'"
                    ></span>
                    <span :class="tunnel.running ? 'text-green-400' : 'text-gray-500'">
                      {{ tunnel.running ? t('dashboard.running') : t('dashboard.stopped') }}
                    </span>
                  </div>
                </td>
              </tr>
              <tr v-if="store.tunnels.length === 0">
                <td colspan="7" class="py-8 text-center text-gray-500">
                  {{ t('dashboard.noTunnels') }} <router-link to="/tunnels" class="text-blue-400 hover:underline">{{ t('dashboard.addOne') }}</router-link>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { NIcon } from 'naive-ui'
import { ArrowUp, ArrowDown, GitNetwork, Time } from '@vicons/ionicons5'
import * as echarts from 'echarts'
import { useDashboardStore } from '../stores/dashboard'
import { useAuthStore } from '../stores/auth'
import { useSettingsStore } from '../stores/settings'
import { useWebSocket } from '../composables/useWebSocket'
import { useI18n } from '../i18n'
import { formatBytes, formatBytesRate, formatUptime, formatPercent, formatLatency } from '../utils/format'
import SettingsDropdown from '../components/SettingsDropdown.vue'

const router = useRouter()
const store = useDashboardStore()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const { t } = useI18n()

const chartRef = ref(null)
let chart = null

const { connected: wsConnected, connect: wsConnect } = useWebSocket((data) => {
  if (data.type === 'dashboard') {
    store.updateData(data.payload)
    updateChart()
  }
})

function getProgressColor(value) {
  if (value < 60) return '#22c55e'
  if (value < 80) return '#eab308'
  return '#ef4444'
}

function getLatencyClass(latency) {
  if (!latency || latency.status === 'error') return 'text-red-400'
  if (latency.status === 'warning') return 'text-yellow-400'
  return 'text-green-400'
}

function initChart() {
  if (!chartRef.value) return

  chart = echarts.init(chartRef.value, settingsStore.isDark ? 'dark' : null)
  chart.setOption({
    backgroundColor: 'transparent',
    grid: {
      left: 50,
      right: 20,
      top: 20,
      bottom: 30
    },
    tooltip: {
      trigger: 'axis',
      backgroundColor: settingsStore.isDark ? '#1f1f1f' : '#fff',
      borderColor: settingsStore.isDark ? '#262626' : '#e2e8f0',
      textStyle: { color: settingsStore.isDark ? '#fff' : '#0f172a' },
      formatter: (params) => {
        const time = new Date(params[0].value[0] * 1000).toLocaleTimeString()
        let html = `<div class="font-medium">${time}</div>`
        params.forEach(p => {
          html += `<div>${p.marker} ${p.seriesName}: ${formatBytesRate(p.value[1])}</div>`
        })
        return html
      }
    },
    legend: {
      data: [t('dashboard.upload'), t('dashboard.download')],
      textStyle: { color: settingsStore.isDark ? '#a1a1aa' : '#64748b' },
      bottom: 0
    },
    xAxis: {
      type: 'time',
      axisLine: { lineStyle: { color: settingsStore.isDark ? '#262626' : '#e2e8f0' } },
      axisLabel: { color: settingsStore.isDark ? '#a1a1aa' : '#64748b', formatter: '{HH}:{mm}:{ss}' },
      splitLine: { show: false }
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      axisLabel: {
        color: settingsStore.isDark ? '#a1a1aa' : '#64748b',
        formatter: (value) => formatBytesRate(value)
      },
      splitLine: { lineStyle: { color: settingsStore.isDark ? '#262626' : '#e2e8f0' } }
    },
    series: [
      {
        name: t('dashboard.upload'),
        type: 'line',
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#3b82f6', width: 2 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(59, 130, 246, 0.3)' },
            { offset: 1, color: 'rgba(59, 130, 246, 0)' }
          ])
        },
        data: []
      },
      {
        name: t('dashboard.download'),
        type: 'line',
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#22c55e', width: 2 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(34, 197, 94, 0.3)' },
            { offset: 1, color: 'rgba(34, 197, 94, 0)' }
          ])
        },
        data: []
      }
    ]
  })

  window.addEventListener('resize', () => chart?.resize())
}

function updateChart() {
  if (!chart) return

  const uploadData = store.trafficHistory.map(h => [h.timestamp, h.rate_out])
  const downloadData = store.trafficHistory.map(h => [h.timestamp, h.rate_in])

  chart.setOption({
    series: [
      { data: uploadData },
      { data: downloadData }
    ]
  })
}

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

onMounted(() => {
  initChart()
  wsConnect()
})

onUnmounted(() => {
  if (chart) {
    chart.dispose()
    chart = null
  }
})
</script>
