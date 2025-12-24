<template>
  <div class="min-h-screen flex items-center justify-center" :class="settingsStore.isDark ? 'bg-dark-bg' : 'bg-gray-100'">
    <div class="card w-full max-w-md relative">
      <!-- Settings Button -->
      <div class="absolute top-4 right-4">
        <SettingsDropdown />
      </div>
      
      <div class="text-center mb-8">
        <h1 class="text-2xl font-bold mb-2" :class="settingsStore.isDark ? 'text-white' : 'text-gray-900'">{{ t('login.title') }}</h1>
        <p class="text-gray-400">{{ t('login.subtitle') }}</p>
      </div>

      <n-form ref="formRef" :model="form" :rules="rules">
        <n-form-item path="username" :label="t('common.username')">
          <n-input
            v-model:value="form.username"
            :placeholder="t('login.enterUsername')"
            size="large"
            @keyup.enter="handleLogin"
          />
        </n-form-item>

        <n-form-item path="password" :label="t('common.password')">
          <n-input
            v-model:value="form.password"
            type="password"
            :placeholder="t('login.enterPassword')"
            size="large"
            show-password-on="click"
            @keyup.enter="handleLogin"
          />
        </n-form-item>

        <n-button
          type="primary"
          block
          size="large"
          :loading="loading"
          @click="handleLogin"
        >
          {{ t('common.login') }}
        </n-button>
      </n-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { useAuthStore } from '../stores/auth'
import { useSettingsStore } from '../stores/settings'
import { useI18n } from '../i18n'
import SettingsDropdown from '../components/SettingsDropdown.vue'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const { t } = useI18n()

const formRef = ref(null)
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules = computed(() => ({
  username: { required: true, message: t('login.pleaseEnterUsername'), trigger: 'blur' },
  password: { required: true, message: t('login.pleaseEnterPassword'), trigger: 'blur' }
}))

async function handleLogin() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    await authStore.login(form.username, form.password)
    message.success(t('login.loginSuccess'))
    router.push('/')
  } catch (error) {
    message.error(error.message || t('login.loginFailed'))
  } finally {
    loading.value = false
  }
}
</script>
