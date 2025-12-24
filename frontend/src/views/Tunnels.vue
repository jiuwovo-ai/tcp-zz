<template>
  <div class="min-h-screen" :class="settingsStore.isDark ? 'bg-dark-bg' : 'bg-gray-50'">
    <!-- Header -->
    <header class="border-b" :class="settingsStore.isDark ? 'border-dark-border bg-dark-card' : 'border-gray-200 bg-white'">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <div class="flex items-center space-x-4">
            <router-link to="/" class="text-gray-400 hover:text-white transition-colors">
              <n-icon size="24"><ArrowBack /></n-icon>
            </router-link>
            <h1 class="text-xl font-bold" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ t('tunnels.title') }}</h1>
          </div>
          <div class="flex items-center space-x-4">
            <SettingsDropdown />
            <n-button type="primary" @click="showCreateModal = true">
              <template #icon><n-icon><Add /></n-icon></template>
              {{ t('tunnels.newTunnel') }}
            </n-button>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div class="card">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="text-left text-gray-400 text-sm border-b" :class="settingsStore.isDark ? 'border-dark-border' : 'border-gray-200'">
                <th class="pb-3 font-medium">{{ t('common.name') }}</th>
                <th class="pb-3 font-medium">{{ t('dashboard.localPort') }}</th>
                <th class="pb-3 font-medium">{{ t('dashboard.target') }}</th>
                <th class="pb-3 font-medium">{{ t('dashboard.protocol') }}</th>
                <th class="pb-3 font-medium">{{ t('tunnels.totalTraffic') }}</th>
                <th class="pb-3 font-medium">{{ t('dashboard.latency') }}</th>
                <th class="pb-3 font-medium">{{ t('common.status') }}</th>
                <th class="pb-3 font-medium text-right">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="tunnel in tunnels"
                :key="tunnel.rule.id"
                class="border-b transition-colors"
                :class="settingsStore.isDark ? 'border-dark-border hover:bg-dark-hover' : 'border-gray-100 hover:bg-gray-50'"
              >
                <td class="py-4 font-medium" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ tunnel.rule.name }}</td>
                <td class="py-4 font-mono" :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">{{ tunnel.rule.local_port }}</td>
                <td class="py-4 font-mono" :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">{{ tunnel.rule.target_ip }}:{{ tunnel.rule.target_port }}</td>
                <td class="py-4">
                  <n-tag :type="tunnel.rule.protocol === 'tcp' ? 'info' : 'warning'" size="small">
                    {{ tunnel.rule.protocol.toUpperCase() }}
                  </n-tag>
                </td>
                <td class="py-4" :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">
                  <div class="text-sm">
                    <span class="text-blue-400">↑ {{ formatBytes(tunnel.traffic.total_out) }}</span>
                    <span class="mx-2 text-gray-500">/</span>
                    <span class="text-green-400">↓ {{ formatBytes(tunnel.traffic.total_in) }}</span>
                  </div>
                </td>
                <td class="py-4">
                  <div class="flex items-center space-x-2">
                    <span
                      class="status-dot"
                      :class="getLatencyStatusClass(tunnel.latency)"
                    ></span>
                    <span :class="getLatencyTextClass(tunnel.latency)">
                      {{ formatLatency(tunnel.latency.latency) }}
                    </span>
                  </div>
                </td>
                <td class="py-4">
                  <n-switch
                    :value="tunnel.running"
                    :loading="toggleLoading[tunnel.rule.id]"
                    @update:value="(val) => handleToggle(tunnel.rule.id, val)"
                  />
                </td>
                <td class="py-4 text-right">
                  <n-space justify="end">
                    <n-button size="small" quaternary @click="handleEdit(tunnel)">
                      <template #icon><n-icon><Create /></n-icon></template>
                    </n-button>
                    <n-popconfirm @positive-click="handleDelete(tunnel.rule.id)">
                      <template #trigger>
                        <n-button size="small" quaternary type="error">
                          <template #icon><n-icon><Trash /></n-icon></template>
                        </n-button>
                      </template>
                      {{ t('tunnels.deleteTunnel') }}
                    </n-popconfirm>
                  </n-space>
                </td>
              </tr>
              <tr v-if="tunnels.length === 0">
                <td colspan="8" class="py-12 text-center text-gray-500">
                  <div class="flex flex-col items-center">
                    <n-icon size="48" class="mb-4 text-gray-600"><GitNetwork /></n-icon>
                    <p class="text-lg mb-2">{{ t('tunnels.noTunnelsTitle') }}</p>
                    <p class="text-sm mb-4">{{ t('tunnels.noTunnelsDesc') }}</p>
                    <n-button type="primary" @click="showCreateModal = true">
                      <template #icon><n-icon><Add /></n-icon></template>
                      {{ t('tunnels.createTunnel') }}
                    </n-button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </main>

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="showCreateModal" preset="card" :title="editingTunnel ? t('tunnels.editTunnel') : t('tunnels.newTunnel')" style="width: 500px;">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="100">
        <n-form-item :label="t('common.name')" path="name">
          <n-input v-model:value="form.name" placeholder="My Tunnel" />
        </n-form-item>

        <n-form-item :label="t('dashboard.localPort')" path="local_port">
          <n-input-number v-model:value="form.local_port" :min="1" :max="65535" placeholder="8080" style="width: 100%;" />
        </n-form-item>

        <n-form-item :label="t('tunnels.targetIP')" path="target_ip">
          <n-input v-model:value="form.target_ip" placeholder="192.168.1.100" />
        </n-form-item>

        <n-form-item :label="t('tunnels.targetPort')" path="target_port">
          <n-input-number v-model:value="form.target_port" :min="1" :max="65535" placeholder="80" style="width: 100%;" />
        </n-form-item>

        <n-form-item :label="t('dashboard.protocol')" path="protocol">
          <n-radio-group v-model:value="form.protocol">
            <n-radio-button value="tcp">TCP</n-radio-button>
            <n-radio-button value="udp">UDP</n-radio-button>
          </n-radio-group>
        </n-form-item>

        <n-form-item :label="t('common.enabled')" path="enabled">
          <n-switch v-model:value="form.enabled" />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="handleCancel">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="submitLoading" @click="handleSubmit">
            {{ editingTunnel ? t('common.update') : t('common.create') }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useMessage } from 'naive-ui'
import { ArrowBack, Add, Create, Trash, GitNetwork } from '@vicons/ionicons5'
import api from '../api'
import { useDashboardStore } from '../stores/dashboard'
import { useSettingsStore } from '../stores/settings'
import { useWebSocket } from '../composables/useWebSocket'
import { useI18n } from '../i18n'
import { formatBytes, formatLatency } from '../utils/format'
import SettingsDropdown from '../components/SettingsDropdown.vue'

const message = useMessage()
const store = useDashboardStore()
const settingsStore = useSettingsStore()
const { t } = useI18n()

const tunnels = ref([])
const showCreateModal = ref(false)
const editingTunnel = ref(null)
const submitLoading = ref(false)
const toggleLoading = ref({})
const formRef = ref(null)

const form = reactive({
  name: '',
  local_port: null,
  target_ip: '',
  target_port: null,
  protocol: 'tcp',
  enabled: true
})

const rules = computed(() => ({
  name: { required: true, message: t('tunnels.pleaseEnterName'), trigger: 'blur' },
  local_port: { required: true, type: 'number', message: t('tunnels.pleaseEnterLocalPort'), trigger: 'blur' },
  target_ip: { required: true, message: t('tunnels.pleaseEnterTargetIP'), trigger: 'blur' },
  target_port: { required: true, type: 'number', message: t('tunnels.pleaseEnterTargetPort'), trigger: 'blur' },
  protocol: { required: true, message: t('tunnels.pleaseSelectProtocol'), trigger: 'change' }
}))

const { connect: wsConnect } = useWebSocket((data) => {
  if (data.type === 'dashboard') {
    store.updateData(data.payload)
    tunnels.value = data.payload.tunnels || []
  }
})

function getLatencyStatusClass(latency) {
  if (!latency || latency.status === 'error' || latency.latency < 0) return 'status-error'
  if (latency.status === 'warning') return 'status-warning'
  return 'status-normal'
}

function getLatencyTextClass(latency) {
  if (!latency || latency.status === 'error' || latency.latency < 0) return 'text-red-400'
  if (latency.status === 'warning') return 'text-yellow-400'
  return 'text-green-400'
}

function resetForm() {
  form.name = ''
  form.local_port = null
  form.target_ip = ''
  form.target_port = null
  form.protocol = 'tcp'
  form.enabled = true
}

function handleEdit(tunnel) {
  editingTunnel.value = tunnel
  form.name = tunnel.rule.name
  form.local_port = tunnel.rule.local_port
  form.target_ip = tunnel.rule.target_ip
  form.target_port = tunnel.rule.target_port
  form.protocol = tunnel.rule.protocol
  form.enabled = tunnel.rule.enabled
  showCreateModal.value = true
}

function handleCancel() {
  showCreateModal.value = false
  editingTunnel.value = null
  resetForm()
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  submitLoading.value = true
  try {
    const payload = {
      name: form.name,
      local_port: form.local_port,
      target_ip: form.target_ip,
      target_port: form.target_port,
      protocol: form.protocol,
      enabled: form.enabled
    }

    if (editingTunnel.value) {
      await api.updateRule(editingTunnel.value.rule.id, payload)
      message.success(t('tunnels.tunnelUpdated'))
    } else {
      await api.createRule(payload)
      message.success(t('tunnels.tunnelCreated'))
    }

    handleCancel()
    await loadTunnels()
  } catch (error) {
    message.error(error.response?.data?.message || t('tunnels.operationFailed'))
  } finally {
    submitLoading.value = false
  }
}

async function handleToggle(id, enabled) {
  toggleLoading.value[id] = true
  try {
    await api.toggleRule(id, enabled)
    message.success(enabled ? t('tunnels.tunnelStarted') : t('tunnels.tunnelStopped'))
  } catch (error) {
    message.error(error.response?.data?.message || t('tunnels.operationFailed'))
  } finally {
    toggleLoading.value[id] = false
  }
}

async function handleDelete(id) {
  try {
    await api.deleteRule(id)
    message.success(t('tunnels.tunnelDeleted'))
    await loadTunnels()
  } catch (error) {
    message.error(error.response?.data?.message || t('tunnels.operationFailed'))
  }
}

async function loadTunnels() {
  try {
    const res = await api.getDashboard()
    if (res.success) {
      tunnels.value = res.data.tunnels || []
    }
  } catch (error) {
    console.error('Failed to load tunnels:', error)
  }
}

onMounted(() => {
  loadTunnels()
  wsConnect()
})
</script>
