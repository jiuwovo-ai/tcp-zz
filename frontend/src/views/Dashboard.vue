<template>
  <div class="min-h-screen" :class="settingsStore.isDark ? 'bg-dark-bg' : 'bg-gray-50'">
    <!-- Header -->
    <header class="border-b backdrop-blur-md sticky top-0 z-50" :class="settingsStore.isDark ? 'border-dark-border bg-dark-card/80' : 'border-gray-200 bg-white/80'">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <div class="flex items-center space-x-4">
            <div class="flex items-center space-x-3">
              <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                <n-icon size="18" class="text-white"><GitNetwork /></n-icon>
              </div>
              <h1 class="text-xl font-bold gradient-text">{{ t('dashboard.title') }}</h1>
            </div>
            <span
              class="status-dot"
              :class="wsConnected ? 'status-normal' : 'status-error'"
            ></span>
          </div>
          <div class="flex items-center space-x-2">
            <router-link to="/nodes">
              <n-button quaternary size="small" class="!rounded-lg">
                <template #icon><n-icon><Server /></n-icon></template>
                {{ t('nodes.manageNodes') }}
              </n-button>
            </router-link>
            <router-link to="/tunnels">
              <n-button quaternary size="small" class="!rounded-lg">
                <template #icon><n-icon><GitNetwork /></n-icon></template>
                {{ t('dashboard.manageTunnels') }}
              </n-button>
            </router-link>
            <SettingsDropdown />
            <n-button quaternary size="small" class="!rounded-lg" @click="handleLogout">
              <template #icon><n-icon><LogOut /></n-icon></template>
              {{ t('common.logout') }}
            </n-button>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Stats Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div class="card card-gradient-blue animate-fade-in-up" style="opacity: 0;">
          <div class="flex items-center justify-between">
            <div>
              <p class="stat-label">{{ t('dashboard.totalUpload') }}</p>
              <p class="stat-value text-blue-400">{{ formatBytes(store.global.total_out) }}</p>
              <p class="text-sm text-gray-500 mt-2">{{ formatBytesRate(store.global.rate_out) }}</p>
            </div>
            <div class="icon-container icon-container-blue">
              <n-icon size="24" class="text-blue-400">
                <ArrowUp />
              </n-icon>
            </div>
          </div>
        </div>

        <div class="card card-gradient-green animate-fade-in-up animate-delay-1" style="opacity: 0;">
          <div class="flex items-center justify-between">
            <div>
              <p class="stat-label">{{ t('dashboard.totalDownload') }}</p>
              <p class="stat-value text-green-400">{{ formatBytes(store.global.total_in) }}</p>
              <p class="text-sm text-gray-500 mt-2">{{ formatBytesRate(store.global.rate_in) }}</p>
            </div>
            <div class="icon-container icon-container-green">
              <n-icon size="24" class="text-green-400">
                <ArrowDown />
              </n-icon>
            </div>
          </div>
        </div>

        <div class="card card-gradient-purple animate-fade-in-up animate-delay-2" style="opacity: 0;">
          <div class="flex items-center justify-between">
            <div>
              <p class="stat-label">{{ t('dashboard.activeTunnels') }}</p>
              <p class="stat-value text-purple-400">{{ store.activeTunnelCount }}</p>
              <p class="text-sm text-gray-500 mt-2">{{ store.tunnels.length }} {{ t('dashboard.total') }}</p>
            </div>
            <div class="icon-container icon-container-purple">
              <n-icon size="24" class="text-purple-400">
                <GitNetwork />
              </n-icon>
            </div>
          </div>
        </div>

        <div class="card card-gradient-yellow animate-fade-in-up animate-delay-3" style="opacity: 0;">
          <div class="flex items-center justify-between">
            <div>
              <p class="stat-label">{{ t('dashboard.uptime') }}</p>
              <p class="stat-value text-yellow-400">{{ formatUptime(store.system.uptime) }}</p>
              <p class="text-sm text-gray-500 mt-2">{{ store.totalConnections }} {{ t('dashboard.connections') }}</p>
            </div>
            <div class="icon-container icon-container-yellow">
              <n-icon size="24" class="text-yellow-400">
                <Time />
              </n-icon>
            </div>
          </div>
        </div>
      </div>

      <!-- Traffic Chart -->
      <div class="card mb-8">
        <h3 class="text-lg font-semibold mb-4" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ t('dashboard.realtimeTraffic') }}</h3>
        <div ref="chartRef" class="h-64"></div>
      </div>

      <!-- Tunnel List -->
      <div class="card">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ t('dashboard.tunnels') }}</h3>
          <n-space>
            <n-button size="small" type="primary" @click="openAddTunnelModal">
              <template #icon><n-icon><Add /></n-icon></template>
              {{ t('tunnels.newTunnel') }}
            </n-button>
            <router-link to="/tunnels">
              <n-button size="small">{{ t('dashboard.manage') }}</n-button>
            </router-link>
          </n-space>
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
                <td class="py-3" :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">{{ tunnel.node_host }}:{{ tunnel.rule.local_port }}</td>
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
                  <span class="badge" :class="tunnel.running ? 'badge-success' : 'badge-error'">
                    <span class="w-1.5 h-1.5 rounded-full mr-1.5" :class="tunnel.running ? 'bg-green-400' : 'bg-red-400'"></span>
                    {{ tunnel.running ? t('dashboard.running') : t('dashboard.stopped') }}
                  </span>
                </td>
              </tr>
              <tr v-if="store.tunnels.length === 0">
                <td colspan="7" class="py-8 text-center text-gray-500">
                  {{ t('dashboard.noTunnels') }} <a href="#" @click.prevent="openAddTunnelModal" class="text-blue-400 hover:underline">{{ t('dashboard.addOne') }}</a>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </main>

    <!-- Add Tunnel Modal -->
    <n-modal v-model:show="showTunnelModal" preset="card" :title="t('tunnels.newTunnel')" style="width: 500px;">
      <n-form ref="tunnelFormRef" :model="tunnelForm" :rules="tunnelFormRules" label-placement="left" label-width="100">
        <n-form-item :label="t('nodes.selectNode')" path="node_id">
          <n-select
            v-model:value="tunnelForm.node_id"
            :options="nodeOptions"
            :placeholder="t('nodes.pleaseSelectNode')"
            :loading="nodesLoading"
          />
        </n-form-item>

        <n-form-item :label="t('common.name')" path="name">
          <n-input v-model:value="tunnelForm.name" placeholder="My Tunnel" />
        </n-form-item>

        <n-form-item :label="t('dashboard.localPort')" path="local_port">
          <n-input-number v-model:value="tunnelForm.local_port" :min="1" :max="65535" placeholder="8080" style="width: 100%;" />
        </n-form-item>

        <n-form-item :label="t('tunnels.targetIP')" path="target_ip">
          <n-input v-model:value="tunnelForm.target_ip" placeholder="192.168.1.100" />
        </n-form-item>

        <n-form-item :label="t('tunnels.targetPort')" path="target_port">
          <n-input-number v-model:value="tunnelForm.target_port" :min="1" :max="65535" placeholder="80" style="width: 100%;" />
        </n-form-item>

        <n-form-item :label="t('dashboard.protocol')" path="protocol">
          <n-radio-group v-model:value="tunnelForm.protocol">
            <n-radio-button value="tcp">TCP</n-radio-button>
            <n-radio-button value="udp">UDP</n-radio-button>
          </n-radio-group>
        </n-form-item>

        <n-form-item :label="t('common.enabled')" path="enabled">
          <n-switch v-model:value="tunnelForm.enabled" />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showTunnelModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="tunnelSubmitLoading" @click="submitTunnel">
            {{ t('common.create') }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { NIcon, useMessage } from 'naive-ui'
import { ArrowUp, ArrowDown, GitNetwork, Time, Add, Server, LogOut } from '@vicons/ionicons5'
import * as echarts from 'echarts'
import { useDashboardStore } from '../stores/dashboard'
import { useAuthStore } from '../stores/auth'
import { useSettingsStore } from '../stores/settings'
import { useWebSocket } from '../composables/useWebSocket'
import { useI18n } from '../i18n'
import { formatBytes, formatBytesRate, formatUptime, formatPercent, formatLatency } from '../utils/format'
import SettingsDropdown from '../components/SettingsDropdown.vue'
import api from '../api'

const router = useRouter()
const message = useMessage()
const store = useDashboardStore()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const { t } = useI18n()

const chartRef = ref(null)
let chart = null

// Tunnel Modal
const showTunnelModal = ref(false)
const tunnelSubmitLoading = ref(false)
const tunnelFormRef = ref(null)
const nodesLoading = ref(false)
const nodes = ref([])

const tunnelForm = reactive({
  node_id: null,
  name: '',
  local_port: null,
  target_ip: '',
  target_port: null,
  protocol: 'tcp',
  enabled: true
})

const nodeOptions = computed(() => {
  return nodes.value.map(node => ({
    label: `${node.name} (${node.host})`,
    value: node.id
  }))
})

const tunnelFormRules = computed(() => ({
  node_id: { required: true, message: t('nodes.pleaseSelectNode'), trigger: 'change' },
  name: { required: true, message: t('tunnels.pleaseEnterName'), trigger: 'blur' },
  local_port: { required: true, type: 'number', message: t('tunnels.pleaseEnterLocalPort'), trigger: 'blur' },
  target_ip: { required: true, message: t('tunnels.pleaseEnterTargetIP'), trigger: 'blur' },
  target_port: { required: true, type: 'number', message: t('tunnels.pleaseEnterTargetPort'), trigger: 'blur' },
  protocol: { required: true, message: t('tunnels.pleaseSelectProtocol'), trigger: 'change' }
}))

async function loadNodes() {
  nodesLoading.value = true
  try {
    const res = await api.getNodes()
    if (res.success) {
      nodes.value = res.data || []
    }
  } catch (error) {
    console.error('Failed to load nodes:', error)
  } finally {
    nodesLoading.value = false
  }
}

function openAddTunnelModal() {
  loadNodes()
  showTunnelModal.value = true
}

function resetTunnelForm() {
  tunnelForm.node_id = null
  tunnelForm.name = ''
  tunnelForm.local_port = null
  tunnelForm.target_ip = ''
  tunnelForm.target_port = null
  tunnelForm.protocol = 'tcp'
  tunnelForm.enabled = true
}

async function submitTunnel() {
  try {
    await tunnelFormRef.value?.validate()
  } catch {
    return
  }

  tunnelSubmitLoading.value = true
  try {
    const payload = {
      node_id: tunnelForm.node_id,
      name: tunnelForm.name,
      local_port: tunnelForm.local_port,
      target_ip: tunnelForm.target_ip,
      target_port: tunnelForm.target_port,
      protocol: tunnelForm.protocol,
      enabled: tunnelForm.enabled
    }

    await api.createNodeRule(payload)
    message.success(t('tunnels.tunnelCreated'))
    showTunnelModal.value = false
    resetTunnelForm()
  } catch (error) {
    message.error(error.response?.data?.message || t('tunnels.operationFailed'))
  } finally {
    tunnelSubmitLoading.value = false
  }
}

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
      left: 70,
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
