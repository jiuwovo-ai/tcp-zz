import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'

export const useSettingsStore = defineStore('settings', () => {
  // 主题: 'light' | 'dark' | 'auto'
  const themeMode = ref(localStorage.getItem('themeMode') || 'auto')
  
  // 语言: 'en' | 'zh'
  const locale = ref(localStorage.getItem('locale') || 'zh')

  // 系统主题偏好
  const systemPrefersDark = ref(
    window.matchMedia('(prefers-color-scheme: dark)').matches
  )

  // 监听系统主题变化
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  mediaQuery.addEventListener('change', (e) => {
    systemPrefersDark.value = e.matches
  })

  // 计算实际主题
  const isDark = computed(() => {
    if (themeMode.value === 'auto') {
      return systemPrefersDark.value
    }
    return themeMode.value === 'dark'
  })

  // 应用主题到 DOM
  function applyTheme() {
    if (isDark.value) {
      document.documentElement.classList.add('dark')
      document.documentElement.classList.remove('light')
    } else {
      document.documentElement.classList.add('light')
      document.documentElement.classList.remove('dark')
    }
  }

  // 设置主题模式
  function setThemeMode(mode) {
    themeMode.value = mode
    localStorage.setItem('themeMode', mode)
    applyTheme()
  }

  // 设置语言
  function setLocale(lang) {
    locale.value = lang
    localStorage.setItem('locale', lang)
  }

  // 初始化
  function init() {
    applyTheme()
  }

  // 监听主题变化
  watch(isDark, () => {
    applyTheme()
  })

  return {
    themeMode,
    locale,
    isDark,
    systemPrefersDark,
    setThemeMode,
    setLocale,
    init
  }
})
