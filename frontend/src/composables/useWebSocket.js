import { ref, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'

export function useWebSocket(onMessage) {
  const ws = ref(null)
  const connected = ref(false)
  let reconnectTimer = null
  let reconnectAttempts = 0
  const maxReconnectAttempts = 10

  function connect() {
    const authStore = useAuthStore()
    if (!authStore.token) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/api/ws?token=${authStore.token}`

    try {
      ws.value = new WebSocket(url)

      ws.value.onopen = () => {
        connected.value = true
        reconnectAttempts = 0
        console.log('WebSocket connected')
      }

      ws.value.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          onMessage?.(data)
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e)
        }
      }

      ws.value.onclose = () => {
        connected.value = false
        console.log('WebSocket disconnected')
        scheduleReconnect()
      }

      ws.value.onerror = (error) => {
        console.error('WebSocket error:', error)
      }
    } catch (error) {
      console.error('Failed to create WebSocket:', error)
      scheduleReconnect()
    }
  }

  function scheduleReconnect() {
    if (reconnectAttempts >= maxReconnectAttempts) {
      console.log('Max reconnect attempts reached')
      return
    }

    const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000)
    reconnectAttempts++

    reconnectTimer = setTimeout(() => {
      console.log(`Reconnecting... (attempt ${reconnectAttempts})`)
      connect()
    }, delay)
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }

    if (ws.value) {
      ws.value.close()
      ws.value = null
    }

    connected.value = false
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    connected,
    connect,
    disconnect
  }
}
