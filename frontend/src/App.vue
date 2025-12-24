<template>
  <n-config-provider :theme="currentTheme" :locale="currentNaiveLocale" :date-locale="currentNaiveDateLocale">
    <n-message-provider>
      <n-dialog-provider>
        <router-view :key="settingsStore.locale" />
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { darkTheme, lightTheme, zhCN, dateZhCN, enUS, dateEnUS } from 'naive-ui'
import { useAuthStore } from './stores/auth'
import { useSettingsStore } from './stores/settings'

const authStore = useAuthStore()
const settingsStore = useSettingsStore()

const currentTheme = computed(() => {
  return settingsStore.isDark ? darkTheme : lightTheme
})

const currentNaiveLocale = computed(() => {
  return settingsStore.locale === 'zh' ? zhCN : enUS
})

const currentNaiveDateLocale = computed(() => {
  return settingsStore.locale === 'zh' ? dateZhCN : dateEnUS
})

onMounted(() => {
  authStore.initAuth()
  settingsStore.init()
})
</script>
