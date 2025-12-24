<template>
  <n-dropdown :options="options" @select="handleSelect" trigger="click">
    <n-button quaternary circle>
      <template #icon>
        <n-icon size="20"><Settings /></n-icon>
      </template>
    </n-button>
  </n-dropdown>

  <!-- 修改密码弹窗 -->
  <n-modal v-model:show="showPasswordModal" preset="dialog" :title="t('settings.changePassword')">
    <n-form ref="formRef" :model="passwordForm" :rules="rules">
      <n-form-item :label="t('settings.oldPassword')" path="oldPassword">
        <n-input v-model:value="passwordForm.oldPassword" type="password" :placeholder="t('settings.oldPasswordPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('settings.newUsername')" path="newUsername">
        <n-input v-model:value="passwordForm.newUsername" :placeholder="t('settings.newUsernamePlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('settings.newPassword')" path="newPassword">
        <n-input v-model:value="passwordForm.newPassword" type="password" :placeholder="t('settings.newPasswordPlaceholder')" />
      </n-form-item>
      <n-form-item :label="t('settings.confirmPassword')" path="confirmPassword">
        <n-input v-model:value="passwordForm.confirmPassword" type="password" :placeholder="t('settings.confirmPasswordPlaceholder')" />
      </n-form-item>
    </n-form>
    <template #action>
      <n-button @click="showPasswordModal = false">{{ t('common.cancel') }}</n-button>
      <n-button type="primary" :loading="saving" @click="handleChangePassword">{{ t('common.save') }}</n-button>
    </template>
  </n-modal>
</template>

<script setup>
import { ref, computed, h } from 'vue'
import { NIcon, useMessage } from 'naive-ui'
import { Settings, Language, Moon, Sunny, Desktop, LockClosed } from '@vicons/ionicons5'
import { useSettingsStore } from '../stores/settings'
import { useI18n } from '../i18n'
import api from '../api'

const settingsStore = useSettingsStore()
const { t, locale } = useI18n()
const message = useMessage()

const showPasswordModal = ref(false)
const saving = ref(false)
const formRef = ref(null)
const passwordForm = ref({
  oldPassword: '',
  newUsername: '',
  newPassword: '',
  confirmPassword: ''
})

const rules = computed(() => ({
  oldPassword: {
    required: true,
    message: t('settings.oldPasswordRequired'),
    trigger: 'blur'
  },
  newPassword: {
    trigger: 'blur',
    validator: (rule, value) => {
      if (value && value.length < 6) {
        return new Error(t('settings.passwordTooShort'))
      }
      return true
    }
  },
  confirmPassword: {
    trigger: 'blur',
    validator: (rule, value) => {
      if (passwordForm.value.newPassword && value !== passwordForm.value.newPassword) {
        return new Error(t('settings.passwordMismatch'))
      }
      return true
    }
  }
}))

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
  },
  {
    type: 'divider',
    key: 'd2'
  },
  {
    label: t('settings.changePassword'),
    key: 'change-password',
    icon: renderIcon(LockClosed)
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
  } else if (key === 'change-password') {
    passwordForm.value = {
      oldPassword: '',
      newUsername: '',
      newPassword: '',
      confirmPassword: ''
    }
    showPasswordModal.value = true
  }
}

async function handleChangePassword() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  if (!passwordForm.value.newPassword && !passwordForm.value.newUsername) {
    message.warning(t('settings.noChanges'))
    return
  }

  saving.value = true
  try {
    const res = await api.changePassword(
      passwordForm.value.oldPassword,
      passwordForm.value.newPassword || '',
      passwordForm.value.newUsername || ''
    )
    if (res.success) {
      message.success(t('settings.passwordChanged'))
      showPasswordModal.value = false
      // 如果修改了密码，需要重新登录
      if (passwordForm.value.newPassword) {
        localStorage.removeItem('token')
        localStorage.removeItem('expiresAt')
        window.location.href = '/login'
      }
    } else {
      message.error(res.message || t('settings.changeFailed'))
    }
  } catch (err) {
    message.error(err.response?.data?.message || t('settings.changeFailed'))
  } finally {
    saving.value = false
  }
}
</script>
