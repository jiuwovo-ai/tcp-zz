<template>
  <n-dropdown :options="options" @select="handleSelect" trigger="click">
    <n-button quaternary circle>
      <template #icon>
        <n-icon size="20"><Settings /></n-icon>
      </template>
    </n-button>
  </n-dropdown>
</template>

<script setup>
import { computed, h } from 'vue'
import { NIcon } from 'naive-ui'
import { Settings, Language, Moon, Sunny, Desktop } from '@vicons/ionicons5'
import { useSettingsStore } from '../stores/settings'
import { useI18n } from '../i18n'

const settingsStore = useSettingsStore()
const { t, locale } = useI18n()

const renderIcon = (icon) => {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const options = computed(() => [
  {
    type: 'group',
    label: t('common.language'),
    key: 'language-group',
    children: [
      {
        label: '中文',
        key: 'lang-zh',
        icon: renderIcon(Language),
        props: {
          style: locale.value === 'zh' ? 'color: #3b82f6; font-weight: bold;' : ''
        }
      },
      {
        label: 'English',
        key: 'lang-en',
        icon: renderIcon(Language),
        props: {
          style: locale.value === 'en' ? 'color: #3b82f6; font-weight: bold;' : ''
        }
      }
    ]
  },
  {
    type: 'divider',
    key: 'd1'
  },
  {
    type: 'group',
    label: t('common.theme'),
    key: 'theme-group',
    children: [
      {
        label: t('common.light'),
        key: 'theme-light',
        icon: renderIcon(Sunny),
        props: {
          style: settingsStore.themeMode === 'light' ? 'color: #3b82f6; font-weight: bold;' : ''
        }
      },
      {
        label: t('common.dark'),
        key: 'theme-dark',
        icon: renderIcon(Moon),
        props: {
          style: settingsStore.themeMode === 'dark' ? 'color: #3b82f6; font-weight: bold;' : ''
        }
      },
      {
        label: t('common.auto'),
        key: 'theme-auto',
        icon: renderIcon(Desktop),
        props: {
          style: settingsStore.themeMode === 'auto' ? 'color: #3b82f6; font-weight: bold;' : ''
        }
      }
    ]
  }
])

function handleSelect(key) {
  if (key.startsWith('lang-')) {
    const lang = key.replace('lang-', '')
    settingsStore.setLocale(lang)
    locale.value = lang
  } else if (key.startsWith('theme-')) {
    const theme = key.replace('theme-', '')
    settingsStore.setThemeMode(theme)
  }
}
</script>
