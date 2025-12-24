import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useDashboardStore = defineStore('dashboard', () => {
  const system = ref({
    cpu_percent: 0,
    memory_percent: 0,
    memory_used: 0,
    memory_total: 0,
    net_bytes_in: 0,
    net_bytes_out: 0,
    net_rate_in: 0,
    net_rate_out: 0,
    active_tunnels: 0,
    uptime: 0
  })

  const global = ref({
    total_in: 0,
    total_out: 0,
    rate_in: 0,
    rate_out: 0
  })

  const tunnels = ref([])

  const trafficHistory = ref([])
  const maxHistoryLength = 60

  const activeTunnelCount = computed(() => {
    return tunnels.value.filter(t => t.running).length
  })

  const totalConnections = computed(() => {
    return tunnels.value.reduce((sum, t) => sum + (t.traffic?.connections || 0), 0)
  })

  function updateData(data) {
    if (data.system) system.value = data.system
    if (data.global) global.value = data.global
    if (data.tunnels) tunnels.value = data.tunnels

    // 添加到历史记录
    trafficHistory.value.push({
      timestamp: data.timestamp || Date.now() / 1000,
      rate_in: data.global?.rate_in || 0,
      rate_out: data.global?.rate_out || 0
    })

    if (trafficHistory.value.length > maxHistoryLength) {
      trafficHistory.value.shift()
    }
  }

  function clearHistory() {
    trafficHistory.value = []
  }

  return {
    system,
    global,
    tunnels,
    trafficHistory,
    activeTunnelCount,
    totalConnections,
    updateData,
    clearHistory
  }
})
