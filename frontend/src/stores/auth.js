import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const expiresAt = ref(parseInt(localStorage.getItem('expiresAt') || '0'))

  const isAuthenticated = computed(() => {
    return token.value && expiresAt.value > Date.now() / 1000
  })

  async function login(username, password) {
    const response = await api.login(username, password)
    if (response.success) {
      token.value = response.data.token
      expiresAt.value = response.data.expires_at
      localStorage.setItem('token', token.value)
      localStorage.setItem('expiresAt', expiresAt.value.toString())
      api.setToken(token.value)
      return true
    }
    throw new Error(response.message || 'Login failed')
  }

  function logout() {
    token.value = ''
    expiresAt.value = 0
    localStorage.removeItem('token')
    localStorage.removeItem('expiresAt')
    api.setToken('')
  }

  function initAuth() {
    if (token.value) {
      api.setToken(token.value)
    }
  }

  return {
    token,
    expiresAt,
    isAuthenticated,
    login,
    logout,
    initAuth
  }
})
