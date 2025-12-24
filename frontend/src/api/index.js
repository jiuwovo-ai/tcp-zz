import axios from 'axios'

const instance = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

let authToken = ''

instance.interceptors.request.use(config => {
  if (authToken) {
    config.headers.Authorization = `Bearer ${authToken}`
  }
  return config
})

instance.interceptors.response.use(
  response => response.data,
  error => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('expiresAt')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default {
  setToken(token) {
    authToken = token
  },

  async login(username, password) {
    return instance.post('/login', { username, password })
  },

  async getDashboard() {
    return instance.get('/dashboard')
  },

  async getRules() {
    return instance.get('/rules')
  },

  async createRule(rule) {
    return instance.post('/rules', rule)
  },

  async updateRule(id, rule) {
    return instance.put(`/rules/${id}`, rule)
  },

  async deleteRule(id) {
    return instance.delete(`/rules/${id}`)
  },

  async toggleRule(id, enabled) {
    return instance.post(`/rules/${id}/toggle`, { enabled })
  },

  async getSystemStats() {
    return instance.get('/system')
  },

  // 节点管理 API
  async getNodes() {
    return instance.get('/nodes')
  },

  async createNode(node) {
    return instance.post('/nodes', node)
  },

  async updateNode(id, node) {
    return instance.put(`/nodes/${id}`, node)
  },

  async deleteNode(id) {
    return instance.delete(`/nodes/${id}`)
  },

  // 节点规则 API
  async getNodeRules(nodeId) {
    const params = nodeId ? { node_id: nodeId } : {}
    return instance.get('/node-rules', { params })
  },

  async createNodeRule(rule) {
    return instance.post('/node-rules', rule)
  },

  async updateNodeRule(id, rule) {
    return instance.put(`/node-rules/${id}`, rule)
  },

  async deleteNodeRule(id) {
    return instance.delete(`/node-rules/${id}`)
  },

  async toggleNodeRule(id, enabled) {
    return instance.post(`/node-rules/${id}/toggle`, { enabled })
  },

  async getNodeInstallScript(nodeId) {
    return instance.get(`/nodes/${nodeId}/install`)
  }
}
