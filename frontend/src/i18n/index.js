import { ref, computed } from 'vue'

const messages = {
  en: {
    common: {
      login: 'Login',
      logout: 'Logout',
      username: 'Username',
      password: 'Password',
      cancel: 'Cancel',
      confirm: 'Confirm',
      create: 'Create',
      update: 'Update',
      delete: 'Delete',
      save: 'Save',
      edit: 'Edit',
      actions: 'Actions',
      status: 'Status',
      enabled: 'Enabled',
      name: 'Name',
      settings: 'Settings',
      language: 'Language',
      theme: 'Theme',
      light: 'Light',
      dark: 'Dark',
      auto: 'Auto'
    },
    login: {
      title: 'Port Forward Dashboard',
      subtitle: 'Login to manage your tunnels',
      enterUsername: 'Enter username',
      enterPassword: 'Enter password',
      loginSuccess: 'Login successful',
      loginFailed: 'Login failed',
      invalidCredentials: 'Invalid credentials',
      pleaseEnterUsername: 'Please enter username',
      pleaseEnterPassword: 'Please enter password'
    },
    dashboard: {
      title: 'Port Forward Dashboard',
      manageTunnels: 'Manage Tunnels',
      totalUpload: 'Total Upload',
      totalDownload: 'Total Download',
      activeTunnels: 'Active Tunnels',
      uptime: 'Uptime',
      total: 'total',
      connections: 'connections',
      systemStatus: 'System Status',
      cpu: 'CPU',
      memory: 'Memory',
      networkIn: 'Network In',
      networkOut: 'Network Out',
      realtimeTraffic: 'Real-time Traffic',
      tunnels: 'Tunnels',
      manage: 'Manage',
      localPort: 'Local Port',
      target: 'Target',
      protocol: 'Protocol',
      traffic: 'Traffic',
      latency: 'Latency',
      running: 'Running',
      stopped: 'Stopped',
      noTunnels: 'No tunnels configured.',
      addOne: 'Add one',
      upload: 'Upload',
      download: 'Download'
    },
    tunnels: {
      title: 'Tunnel Management',
      newTunnel: 'New Tunnel',
      editTunnel: 'Edit Tunnel',
      totalTraffic: 'Total Traffic',
      createTunnel: 'Create Tunnel',
      noTunnelsTitle: 'No tunnels configured',
      noTunnelsDesc: 'Create your first tunnel to get started',
      deleteTunnel: 'Delete this tunnel?',
      tunnelCreated: 'Tunnel created',
      tunnelUpdated: 'Tunnel updated',
      tunnelDeleted: 'Tunnel deleted',
      tunnelStarted: 'Tunnel started',
      tunnelStopped: 'Tunnel stopped',
      operationFailed: 'Operation failed',
      targetIP: 'Target IP',
      targetPort: 'Target Port',
      pleaseEnterName: 'Please enter a name',
      pleaseEnterLocalPort: 'Please enter local port',
      pleaseEnterTargetIP: 'Please enter target IP',
      pleaseEnterTargetPort: 'Please enter target port',
      pleaseSelectProtocol: 'Please select protocol'
    },
    nodes: {
      title: 'Node Management',
      addNode: 'Add Node',
      editNode: 'Edit Node',
      deleteNode: 'Delete this node?',
      nodeCreated: 'Node created',
      nodeUpdated: 'Node updated',
      nodeDeleted: 'Node deleted',
      noNodes: 'No nodes configured',
      noNodesDesc: 'Add your first node to get started',
      address: 'Address',
      status: 'Status',
      online: 'Online',
      offline: 'Offline',
      memory: 'Memory',
      tunnels: 'Tunnels',
      tunnelList: 'Tunnel List',
      addTunnel: 'Add Tunnel',
      noTunnelsOnNode: 'No tunnels on this node',
      host: 'Host',
      port: 'Port',
      key: 'Key',
      nodeNamePlaceholder: 'Hong Kong Node',
      keyPlaceholder: 'Node authentication key',
      pleaseEnterName: 'Please enter node name',
      pleaseEnterHost: 'Please enter host address',
      pleaseEnterPort: 'Please enter port',
      pleaseEnterKey: 'Please enter node key',
      pleaseSelectNode: 'Please select a node',
      selectNode: 'Node',
      manageNodes: 'Manage Nodes',
      installCommand: 'Install Command',
      installTip: 'One-Click Installation',
      installTipDesc: 'Copy the command below and run it on your server to automatically install and configure the node agent.',
      oneLineInstall: 'One-line install command:',
      copy: 'Copy',
      copied: 'Copied to clipboard',
      viewFullScript: 'View full installation script',
      generateKey: 'Generate'
    }
  },
  zh: {
    common: {
      login: '登录',
      logout: '退出',
      username: '用户名',
      password: '密码',
      cancel: '取消',
      confirm: '确认',
      create: '创建',
      update: '更新',
      delete: '删除',
      save: '保存',
      edit: '编辑',
      actions: '操作',
      status: '状态',
      enabled: '启用',
      name: '名称',
      settings: '设置',
      language: '语言',
      theme: '主题',
      light: '浅色',
      dark: '深色',
      auto: '自动'
    },
    login: {
      title: '端口转发管理面板',
      subtitle: '登录以管理您的隧道',
      enterUsername: '请输入用户名',
      enterPassword: '请输入密码',
      loginSuccess: '登录成功',
      loginFailed: '登录失败',
      invalidCredentials: '用户名或密码错误',
      pleaseEnterUsername: '请输入用户名',
      pleaseEnterPassword: '请输入密码'
    },
    dashboard: {
      title: '端口转发管理面板',
      manageTunnels: '管理隧道',
      totalUpload: '总上传',
      totalDownload: '总下载',
      activeTunnels: '活跃隧道',
      uptime: '运行时间',
      total: '共',
      connections: '个连接',
      systemStatus: '系统状态',
      cpu: 'CPU',
      memory: '内存',
      networkIn: '网络入站',
      networkOut: '网络出站',
      realtimeTraffic: '实时流量',
      tunnels: '隧道列表',
      manage: '管理',
      localPort: '本地端口',
      target: '目标地址',
      protocol: '协议',
      traffic: '流量',
      latency: '延迟',
      running: '运行中',
      stopped: '已停止',
      noTunnels: '暂无隧道配置。',
      addOne: '添加一个',
      upload: '上传',
      download: '下载'
    },
    tunnels: {
      title: '隧道管理',
      newTunnel: '新建隧道',
      editTunnel: '编辑隧道',
      totalTraffic: '总流量',
      createTunnel: '创建隧道',
      noTunnelsTitle: '暂无隧道配置',
      noTunnelsDesc: '创建您的第一个隧道开始使用',
      deleteTunnel: '确定删除此隧道？',
      tunnelCreated: '隧道创建成功',
      tunnelUpdated: '隧道更新成功',
      tunnelDeleted: '隧道删除成功',
      tunnelStarted: '隧道已启动',
      tunnelStopped: '隧道已停止',
      operationFailed: '操作失败',
      targetIP: '目标 IP',
      targetPort: '目标端口',
      pleaseEnterName: '请输入名称',
      pleaseEnterLocalPort: '请输入本地端口',
      pleaseEnterTargetIP: '请输入目标 IP',
      pleaseEnterTargetPort: '请输入目标端口',
      pleaseSelectProtocol: '请选择协议'
    },
    nodes: {
      title: '节点管理',
      addNode: '添加节点',
      editNode: '编辑节点',
      deleteNode: '确定删除此节点？',
      nodeCreated: '节点创建成功',
      nodeUpdated: '节点更新成功',
      nodeDeleted: '节点删除成功',
      noNodes: '暂无节点',
      noNodesDesc: '添加您的第一个节点开始使用',
      address: '地址',
      status: '状态',
      online: '在线',
      offline: '离线',
      memory: '内存',
      tunnels: '隧道',
      tunnelList: '隧道列表',
      addTunnel: '添加隧道',
      noTunnelsOnNode: '该节点暂无隧道',
      host: '主机地址',
      port: '端口',
      key: '密钥',
      nodeNamePlaceholder: '香港节点',
      keyPlaceholder: '节点认证密钥',
      pleaseEnterName: '请输入节点名称',
      pleaseEnterHost: '请输入主机地址',
      pleaseEnterPort: '请输入端口',
      pleaseEnterKey: '请输入节点密钥',
      pleaseSelectNode: '请选择节点',
      selectNode: '节点',
      manageNodes: '管理节点',
      installCommand: '安装命令',
      installTip: '一键安装',
      installTipDesc: '复制下方命令到您的服务器执行，即可自动安装并配置节点 Agent。',
      oneLineInstall: '一键安装命令：',
      copy: '复制',
      copied: '已复制到剪贴板',
      viewFullScript: '查看完整安装脚本',
      generateKey: '生成'
    }
  }
}

const currentLocale = ref(localStorage.getItem('locale') || 'zh')

export function useI18n() {
  const locale = computed({
    get: () => currentLocale.value,
    set: (val) => {
      currentLocale.value = val
      localStorage.setItem('locale', val)
    }
  })

  function t(key) {
    const keys = key.split('.')
    let value = messages[currentLocale.value]
    for (const k of keys) {
      if (value && typeof value === 'object') {
        value = value[k]
      } else {
        return key
      }
    }
    return value || key
  }

  function setLocale(lang) {
    locale.value = lang
  }

  return {
    locale,
    t,
    setLocale,
    availableLocales: ['en', 'zh']
  }
}

export { messages, currentLocale }
