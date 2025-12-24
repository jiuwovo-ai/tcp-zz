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
            <h1 class="text-xl font-bold" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ t('nodes.title') }}</h1>
          </div>
          <div class="flex items-center space-x-4">
            <SettingsDropdown />
            <n-button type="primary" @click="showNodeModal = true">
              <template #icon><n-icon><Add /></n-icon></template>
              {{ t('nodes.addNode') }}
            </n-button>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Nodes Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
        <div
          v-for="node in nodes"
          :key="node.id"
          class="card cursor-pointer hover:border-blue-500 transition-colors"
          @click="selectNode(node)"
        >
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center space-x-3">
              <span
                class="status-dot"
                :class="node.online ? 'status-normal' : 'status-error'"
              ></span>
              <h3 class="font-semibold" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ node.name }}</h3>
            </div>
            <n-space>
              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-button size="small" quaternary @click.stop="showInstallCommand(node)">
                    <template #icon><n-icon><Terminal /></n-icon></template>
                  </n-button>
                </template>
                {{ t('nodes.installCommand') }}
              </n-tooltip>
              <n-button size="small" quaternary @click.stop="editNode(node)">
                <template #icon><n-icon><Create /></n-icon></template>
              </n-button>
              <n-popconfirm @positive-click="deleteNode(node.id)">
                <template #trigger>
                  <n-button size="small" quaternary type="error" @click.stop>
                    <template #icon><n-icon><Trash /></n-icon></template>
                  </n-button>
                </template>
                {{ t('nodes.deleteNode') }}
              </n-popconfirm>
            </n-space>
          </div>

          <div class="space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="text-gray-400">{{ t('nodes.address') }}</span>
              <span :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">{{ node.host }}:{{ node.port }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-400">{{ t('nodes.status') }}</span>
              <span :class="node.online ? 'text-green-400' : 'text-red-400'">
                {{ node.online ? t('nodes.online') : t('nodes.offline') }}
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-400">CPU</span>
              <span :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">{{ node.cpu_percent?.toFixed(1) || 0 }}%</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-400">{{ t('nodes.memory') }}</span>
              <span :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">{{ node.mem_percent?.toFixed(1) || 0 }}%</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-400">{{ t('nodes.tunnels') }}</span>
              <span :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">{{ node.active_tunnels || 0 }} / {{ node.tunnel_count || 0 }}</span>
            </div>
          </div>
        </div>

        <!-- Empty State -->
        <div v-if="nodes.length === 0" class="col-span-full">
          <div class="card py-12 text-center">
            <n-icon size="48" class="mb-4 text-gray-600"><Server /></n-icon>
            <p class="text-lg mb-2" :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">{{ t('nodes.noNodes') }}</p>
            <p class="text-sm text-gray-500 mb-4">{{ t('nodes.noNodesDesc') }}</p>
            <n-button type="primary" @click="showNodeModal = true">
              <template #icon><n-icon><Add /></n-icon></template>
              {{ t('nodes.addNode') }}
            </n-button>
          </div>
        </div>
      </div>

      <!-- Selected Node Tunnels -->
      <div v-if="selectedNode" class="card">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">
            {{ selectedNode.name }} - {{ t('nodes.tunnelList') }}
          </h3>
          <n-button type="primary" size="small" @click="showRuleModal = true">
            <template #icon><n-icon><Add /></n-icon></template>
            {{ t('nodes.addTunnel') }}
          </n-button>
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
                <th class="pb-3 font-medium text-right">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="rule in nodeRules"
                :key="rule.id"
                class="border-b transition-colors"
                :class="settingsStore.isDark ? 'border-dark-border hover:bg-dark-hover' : 'border-gray-100 hover:bg-gray-50'"
              >
                <td class="py-4 font-medium" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ rule.name }}</td>
                <td class="py-4 font-mono" :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">{{ rule.local_port }}</td>
                <td class="py-4 font-mono" :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">{{ rule.target_ip }}:{{ rule.target_port }}</td>
                <td class="py-4">
                  <n-tag :type="rule.protocol === 'tcp' ? 'info' : 'warning'" size="small">
                    {{ rule.protocol?.toUpperCase() }}
                  </n-tag>
                </td>
                <td class="py-4" :class="settingsStore.isDark ? 'text-gray-300' : 'text-gray-600'">
                  <span class="text-blue-400">↑ {{ formatBytes(rule.bytes_out || 0) }}</span>
                  <span class="mx-2 text-gray-500">/</span>
                  <span class="text-green-400">↓ {{ formatBytes(rule.bytes_in || 0) }}</span>
                </td>
                <td class="py-4">
                  <span :class="getLatencyClass(rule.latency)">
                    {{ rule.latency >= 0 ? rule.latency + 'ms' : 'N/A' }}
                  </span>
                </td>
                <td class="py-4">
                  <n-switch
                    :value="rule.enabled"
                    :loading="toggleLoading[rule.id]"
                    @update:value="(val) => toggleRule(rule.id, val)"
                  />
                </td>
                <td class="py-4 text-right">
                  <n-space justify="end">
                    <n-button size="small" quaternary @click="editRule(rule)">
                      <template #icon><n-icon><Create /></n-icon></template>
                    </n-button>
                    <n-popconfirm @positive-click="deleteRule(rule.id)">
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
              <tr v-if="nodeRules.length === 0">
                <td colspan="8" class="py-8 text-center text-gray-500">
                  {{ t('nodes.noTunnelsOnNode') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </main>

    <!-- Node Modal -->
    <n-modal v-model:show="showNodeModal" preset="card" :title="editingNode ? t('nodes.editNode') : t('nodes.addNode')" style="width: 500px;">
      <n-form ref="nodeFormRef" :model="nodeForm" :rules="nodeFormRulesComputed" label-placement="left" label-width="100">
        <n-form-item :label="t('common.name')" path="name">
          <n-input v-model:value="nodeForm.name" :placeholder="t('nodes.nodeNamePlaceholder')" />
        </n-form-item>

        <n-form-item :label="t('nodes.host')" path="host">
          <n-input v-model:value="nodeForm.host" placeholder="192.168.1.100" />
        </n-form-item>

        <n-form-item :label="t('nodes.key')" path="key">
          <n-input-group>
            <n-input v-model:value="nodeForm.key" :placeholder="t('nodes.keyPlaceholder')" style="flex: 1;" />
            <n-button @click="generateRandomKey">{{ t('nodes.generateKey') }}</n-button>
          </n-input-group>
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="cancelNodeModal">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="nodeSubmitLoading" @click="submitNode">
            {{ editingNode ? t('common.update') : t('common.create') }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Rule Modal -->
    <n-modal v-model:show="showRuleModal" preset="card" :title="editingRule ? t('tunnels.editTunnel') : t('tunnels.newTunnel')" style="width: 500px;">
      <n-form ref="ruleFormRef" :model="ruleForm" :rules="ruleFormRules" label-placement="left" label-width="100">
        <n-form-item :label="t('nodes.selectNode')" path="node_id">
          <n-select
            v-model:value="ruleForm.node_id"
            :options="nodeOptions"
            :placeholder="t('nodes.pleaseSelectNode')"
          />
        </n-form-item>

        <n-form-item :label="t('common.name')" path="name">
          <n-input v-model:value="ruleForm.name" placeholder="My Tunnel" />
        </n-form-item>

        <n-form-item :label="t('dashboard.localPort')" path="local_port">
          <n-input-number v-model:value="ruleForm.local_port" :min="1" :max="65535" placeholder="8080" style="width: 100%;" />
        </n-form-item>

        <n-form-item :label="t('tunnels.targetIP')" path="target_ip">
          <n-input v-model:value="ruleForm.target_ip" placeholder="192.168.1.100" />
        </n-form-item>

        <n-form-item :label="t('tunnels.targetPort')" path="target_port">
          <n-input-number v-model:value="ruleForm.target_port" :min="1" :max="65535" placeholder="80" style="width: 100%;" />
        </n-form-item>

        <n-form-item :label="t('dashboard.protocol')" path="protocol">
          <n-radio-group v-model:value="ruleForm.protocol">
            <n-radio-button value="tcp">TCP</n-radio-button>
            <n-radio-button value="udp">UDP</n-radio-button>
          </n-radio-group>
        </n-form-item>

        <n-form-item :label="t('common.enabled')" path="enabled">
          <n-switch v-model:value="ruleForm.enabled" />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="cancelRuleModal">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="ruleSubmitLoading" @click="submitRule">
            {{ editingRule ? t('common.update') : t('common.create') }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Install Command Modal -->
    <n-modal v-model:show="showInstallModal" preset="card" :title="t('nodes.installCommand')" style="width: 720px;">
      <div class="space-y-6">
        <!-- Node Info Header -->
        <div class="flex items-center space-x-4 p-4 rounded-lg" :class="settingsStore.isDark ? 'bg-dark-hover' : 'bg-blue-50'">
          <div class="w-12 h-12 rounded-full flex items-center justify-center" :class="settingsStore.isDark ? 'bg-blue-500/20' : 'bg-blue-100'">
            <n-icon size="24" class="text-blue-500"><Server /></n-icon>
          </div>
          <div>
            <h3 class="font-semibold text-lg" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ installNodeName }}</h3>
            <p class="text-sm text-gray-400">{{ installNodeHost }}</p>
          </div>
        </div>

        <!-- Steps -->
        <div class="space-y-4">
          <div class="flex items-start space-x-3">
            <div class="w-6 h-6 rounded-full bg-blue-500 text-white flex items-center justify-center text-sm font-medium flex-shrink-0">1</div>
            <div class="flex-1">
              <p class="font-medium mb-1" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ t('nodes.installStep1') }}</p>
              <p class="text-sm text-gray-400">{{ t('nodes.installStep1Desc') }}</p>
            </div>
          </div>

          <div class="flex items-start space-x-3">
            <div class="w-6 h-6 rounded-full bg-blue-500 text-white flex items-center justify-center text-sm font-medium flex-shrink-0">2</div>
            <div class="flex-1">
              <p class="font-medium mb-2" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ t('nodes.installStep2') }}</p>
              <div class="relative">
                <div class="p-3 rounded-lg font-mono text-sm overflow-x-auto" :class="settingsStore.isDark ? 'bg-gray-900 text-green-400' : 'bg-gray-900 text-green-400'">
                  <code class="break-all whitespace-pre-wrap">{{ installCommand }}</code>
                </div>
                <n-button
                  size="small"
                  type="primary"
                  class="absolute top-2 right-2"
                  @click="copyCommand"
                >
                  <template #icon><n-icon><Copy /></n-icon></template>
                  {{ copied ? t('nodes.copied') : t('nodes.copy') }}
                </n-button>
              </div>
            </div>
          </div>

          <div class="flex items-start space-x-3">
            <div class="w-6 h-6 rounded-full bg-blue-500 text-white flex items-center justify-center text-sm font-medium flex-shrink-0">3</div>
            <div class="flex-1">
              <p class="font-medium mb-1" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ t('nodes.installStep3') }}</p>
              <p class="text-sm text-gray-400">{{ t('nodes.installStep3Desc') }}</p>
            </div>
          </div>
        </div>

        <!-- Full Script Collapse -->
        <n-collapse>
          <n-collapse-item :title="t('nodes.viewFullScript')" name="script">
            <div class="p-3 rounded-lg font-mono text-xs overflow-x-auto max-h-64" :class="settingsStore.isDark ? 'bg-gray-900 text-gray-300' : 'bg-gray-900 text-gray-300'">
              <pre class="whitespace-pre-wrap">{{ installScript }}</pre>
            </div>
          </n-collapse-item>
        </n-collapse>
      </div>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showInstallModal = false">{{ t('common.confirm') }}</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { ArrowBack, Add, Create, Trash, Server, Terminal, Copy } from '@vicons/ionicons5'
import api from '../api'
import { useSettingsStore } from '../stores/settings'
import { useI18n } from '../i18n'
import { formatBytes } from '../utils/format'
import SettingsDropdown from '../components/SettingsDropdown.vue'

const message = useMessage()
const settingsStore = useSettingsStore()
const { t } = useI18n()

const nodes = ref([])
const selectedNode = ref(null)
const nodeRules = ref([])
const toggleLoading = ref({})

// Install Modal
const showInstallModal = ref(false)
const installCommand = ref('')
const installScript = ref('')
const installNodeName = ref('')
const installNodeHost = ref('')
const copied = ref(false)

// Node Modal
const showNodeModal = ref(false)
const editingNode = ref(null)
const nodeSubmitLoading = ref(false)
const nodeFormRef = ref(null)
const nodeForm = reactive({
  name: '',
  host: '',
  key: ''
})

// Rule Modal
const showRuleModal = ref(false)
const editingRule = ref(null)
const ruleSubmitLoading = ref(false)
const ruleFormRef = ref(null)
const ruleForm = reactive({
  node_id: null,
  name: '',
  local_port: null,
  target_ip: '',
  target_port: null,
  protocol: 'tcp',
  enabled: true
})

// Node options for select
const nodeOptions = computed(() => {
  return nodes.value.map(node => ({
    label: `${node.name} (${node.host})`,
    value: node.id
  }))
})

const nodeFormRulesComputed = computed(() => ({
  name: { required: true, message: t('nodes.pleaseEnterName'), trigger: 'blur' },
  host: { required: true, message: t('nodes.pleaseEnterHost'), trigger: 'blur' },
  key: { required: true, message: t('nodes.pleaseEnterKey'), trigger: 'blur' }
}))

const ruleFormRules = computed(() => ({
  node_id: { required: true, message: t('nodes.pleaseSelectNode'), trigger: 'change' },
  name: { required: true, message: t('tunnels.pleaseEnterName'), trigger: 'blur' },
  local_port: { required: true, type: 'number', message: t('tunnels.pleaseEnterLocalPort'), trigger: 'blur' },
  target_ip: { required: true, message: t('tunnels.pleaseEnterTargetIP'), trigger: 'blur' },
  target_port: { required: true, type: 'number', message: t('tunnels.pleaseEnterTargetPort'), trigger: 'blur' },
  protocol: { required: true, message: t('tunnels.pleaseSelectProtocol'), trigger: 'change' }
}))

function getLatencyClass(latency) {
  if (latency < 0) return 'text-red-400'
  if (latency < 100) return 'text-green-400'
  if (latency < 300) return 'text-yellow-400'
  return 'text-red-400'
}

async function loadNodes() {
  try {
    const res = await api.getNodes()
    if (res.success) {
      nodes.value = res.data || []
    }
  } catch (error) {
    console.error('Failed to load nodes:', error)
  }
}

async function loadNodeRules(nodeId) {
  try {
    const res = await api.getNodeRules(nodeId)
    if (res.success) {
      nodeRules.value = res.data || []
    }
  } catch (error) {
    console.error('Failed to load node rules:', error)
  }
}

function selectNode(node) {
  selectedNode.value = node
  loadNodeRules(node.id)
}

function editNode(node) {
  editingNode.value = node
  nodeForm.name = node.name
  nodeForm.host = node.host
  nodeForm.key = node.key
  showNodeModal.value = true
}

function cancelNodeModal() {
  showNodeModal.value = false
  editingNode.value = null
  nodeForm.name = ''
  nodeForm.host = ''
  nodeForm.key = ''
}

function generateRandomKey() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let key = ''
  for (let i = 0; i < 32; i++) {
    key += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  nodeForm.key = key
}

async function submitNode() {
  try {
    await nodeFormRef.value?.validate()
  } catch {
    return
  }

  nodeSubmitLoading.value = true
  try {
    const payload = {
      name: nodeForm.name,
      host: nodeForm.host,
      port: 9090,
      key: nodeForm.key
    }

    if (editingNode.value) {
      await api.updateNode(editingNode.value.id, payload)
      message.success(t('nodes.nodeUpdated'))
    } else {
      await api.createNode(payload)
      message.success(t('nodes.nodeCreated'))
    }

    cancelNodeModal()
    await loadNodes()
  } catch (error) {
    message.error(error.response?.data?.message || t('tunnels.operationFailed'))
  } finally {
    nodeSubmitLoading.value = false
  }
}

async function deleteNode(id) {
  try {
    await api.deleteNode(id)
    message.success(t('nodes.nodeDeleted'))
    if (selectedNode.value?.id === id) {
      selectedNode.value = null
      nodeRules.value = []
    }
    await loadNodes()
  } catch (error) {
    message.error(error.response?.data?.message || t('tunnels.operationFailed'))
  }
}

function editRule(rule) {
  editingRule.value = rule
  ruleForm.node_id = rule.node_id || selectedNode.value?.id
  ruleForm.name = rule.name
  ruleForm.local_port = rule.local_port
  ruleForm.target_ip = rule.target_ip
  ruleForm.target_port = rule.target_port
  ruleForm.protocol = rule.protocol
  ruleForm.enabled = rule.enabled
  showRuleModal.value = true
}

function cancelRuleModal() {
  showRuleModal.value = false
  editingRule.value = null
  ruleForm.node_id = null
  ruleForm.name = ''
  ruleForm.local_port = null
  ruleForm.target_ip = ''
  ruleForm.target_port = null
  ruleForm.protocol = 'tcp'
  ruleForm.enabled = true
}

async function submitRule() {
  try {
    await ruleFormRef.value?.validate()
  } catch {
    return
  }

  ruleSubmitLoading.value = true
  try {
    const payload = {
      node_id: ruleForm.node_id,
      name: ruleForm.name,
      local_port: ruleForm.local_port,
      target_ip: ruleForm.target_ip,
      target_port: ruleForm.target_port,
      protocol: ruleForm.protocol,
      enabled: ruleForm.enabled
    }

    if (editingRule.value) {
      await api.updateNodeRule(editingRule.value.id, payload)
      message.success(t('tunnels.tunnelUpdated'))
    } else {
      await api.createNodeRule(payload)
      message.success(t('tunnels.tunnelCreated'))
    }

    cancelRuleModal()
    // Reload rules for the selected node
    if (ruleForm.node_id) {
      const targetNode = nodes.value.find(n => n.id === payload.node_id)
      if (targetNode) {
        selectNode(targetNode)
      }
    }
    await loadNodes()
  } catch (error) {
    message.error(error.response?.data?.message || t('tunnels.operationFailed'))
  } finally {
    ruleSubmitLoading.value = false
  }
}

async function toggleRule(id, enabled) {
  toggleLoading.value[id] = true
  try {
    await api.toggleNodeRule(id, enabled)
    message.success(enabled ? t('tunnels.tunnelStarted') : t('tunnels.tunnelStopped'))
    if (selectedNode.value) {
      await loadNodeRules(selectedNode.value.id)
    }
  } catch (error) {
    message.error(error.response?.data?.message || t('tunnels.operationFailed'))
  } finally {
    toggleLoading.value[id] = false
  }
}

async function deleteRule(id) {
  try {
    await api.deleteNodeRule(id)
    message.success(t('tunnels.tunnelDeleted'))
    if (selectedNode.value) {
      await loadNodeRules(selectedNode.value.id)
    }
  } catch (error) {
    message.error(error.response?.data?.message || t('tunnels.operationFailed'))
  }
}

async function showInstallCommand(node) {
  try {
    const res = await api.getNodeInstallScript(node.id)
    if (res.success) {
      installCommand.value = res.data.command
      installScript.value = res.data.script
      installNodeName.value = node.name
      installNodeHost.value = node.host
      showInstallModal.value = true
    }
  } catch (error) {
    message.error(error.response?.data?.message || t('tunnels.operationFailed'))
  }
}

function copyCommand() {
  navigator.clipboard.writeText(installCommand.value)
  copied.value = true
  message.success(t('nodes.copied'))
  setTimeout(() => {
    copied.value = false
  }, 2000)
}

onMounted(() => {
  loadNodes()
})
</script>
