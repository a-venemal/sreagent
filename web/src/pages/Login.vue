<script setup lang="ts">
import { ref, inject, onMounted } from 'vue'
import type { Ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import { authApi } from '@/api'
import { GlobeOutline, SunnyOutline, MoonOutline, LogInOutline } from '@vicons/ionicons5'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const authStore = useAuthStore()
const { t, locale } = useI18n()

const isDark = inject<Ref<boolean>>('isDark', ref(true))
const toggleTheme = inject<() => void>('toggleTheme', () => {})

const form = ref({
  username: '',
  password: '',
})
const loading = ref(false)

// OIDC SSO state
const oidcEnabled = ref(false)
const oidcLoginUrl = ref('')
const oidcLoading = ref(false)

const langOptions = [
  { label: '简体中文', value: 'zh-CN' },
  { label: 'English', value: 'en' },
]

function handleLangChange(val: string) {
  locale.value = val
  localStorage.setItem('locale', val)
}

async function handleLogin() {
  if (!form.value.username || !form.value.password) {
    message.warning(t('auth.pleaseEnter') || 'Please enter username and password')
    return
  }

  loading.value = true
  try {
    await authStore.login(form.value.username, form.value.password)
    message.success(t('auth.loginSuccess'))
    router.push((route.query.redirect as string) || '/dashboard')
  } catch (err: any) {
    message.error(err.message || t('auth.loginFailed'))
  } finally {
    loading.value = false
  }
}

function handleSSOLogin() {
  if (oidcLoginUrl.value) {
    window.location.href = oidcLoginUrl.value
  }
}

async function checkOIDCConfig() {
  try {
    const { data } = await authApi.getOIDCConfig()
    if (data.data.enabled && data.data.login_url) {
      oidcEnabled.value = true
      oidcLoginUrl.value = data.data.login_url
    }
  } catch {
    // OIDC not configured, that's fine
  }
}

onMounted(() => {
  checkOIDCConfig()
})
</script>

<template>
  <div class="login-container" :class="{ light: !isDark }">
    <div class="login-bg">
      <div class="grid-lines" :class="{ light: !isDark }"></div>
      <div class="glow-orb orb-1"></div>
      <div class="glow-orb orb-2"></div>
    </div>

    <!-- Top right controls: language + theme -->
    <div class="login-controls">
      <n-select
        :value="locale"
        :options="langOptions"
        size="small"
        style="width: 120px"
        @update:value="handleLangChange"
      />
      <n-button text @click="toggleTheme" style="padding: 4px 8px">
        <n-icon :component="isDark ? SunnyOutline : MoonOutline" :size="18" />
      </n-button>
    </div>

    <div class="login-card" :class="{ light: !isDark }">
      <div class="login-header">
        <h1 class="logo-text">
          <span class="gradient-text">SRE</span><span class="agent-text" :class="{ light: !isDark }">Agent</span>
        </h1>
        <p class="login-subtitle" :class="{ light: !isDark }">{{ t('auth.subtitle') }}</p>
      </div>

      <n-form @submit.prevent="handleLogin">
        <n-form-item :label="t('auth.username')" :show-feedback="false" style="margin-bottom: 20px">
          <n-input
            v-model:value="form.username"
            :placeholder="t('auth.enterUsername') || 'Enter username'"
            size="large"
            :autofocus="true"
          />
        </n-form-item>

        <n-form-item :label="t('auth.password')" :show-feedback="false" style="margin-bottom: 28px">
          <n-input
            v-model:value="form.password"
            type="password"
            :placeholder="t('auth.enterPassword') || 'Enter password'"
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
          style="height: 44px; font-size: 16px"
        >
          {{ t('auth.signIn') }}
        </n-button>
      </n-form>

      <!-- SSO Login -->
      <div v-if="oidcEnabled" class="sso-section">
        <n-divider>
          <n-text depth="3" style="font-size: 12px">{{ t('auth.orContinueWith') }}</n-text>
        </n-divider>
        <n-button
          block
          size="large"
          secondary
          @click="handleSSOLogin"
          :loading="oidcLoading"
          style="height: 44px; font-size: 14px"
        >
          <template #icon><n-icon :component="LogInOutline" /></template>
          {{ t('auth.ssoLogin') }}
        </n-button>
      </div>

      <div class="login-footer">
        <n-text depth="3" style="font-size: 12px">
          {{ t('auth.defaultCredentials') }}
        </n-text>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  background: #0a0a0f;
  overflow: hidden;
  transition: background 0.3s ease;
}

.login-container.light {
  background: #f0f2f5;
}

.login-controls {
  position: absolute;
  top: 20px;
  right: 24px;
  z-index: 10;
  display: flex;
  align-items: center;
  gap: 8px;
}

.login-bg {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
}

.grid-lines {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.03) 1px, transparent 1px);
  background-size: 60px 60px;
}

.grid-lines.light {
  background-image:
    linear-gradient(rgba(0, 0, 0, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 0, 0, 0.04) 1px, transparent 1px);
}

.glow-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
}

.orb-1 {
  width: 400px; height: 400px;
  background: rgba(24, 160, 88, 0.15);
  top: 10%; right: 20%;
}

.orb-2 {
  width: 300px; height: 300px;
  background: rgba(112, 192, 232, 0.1);
  bottom: 20%; left: 15%;
}

.login-card {
  width: 400px;
  padding: 48px 40px;
  background: rgba(24, 24, 28, 0.8);
  backdrop-filter: blur(20px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  position: relative;
  z-index: 1;
  transition: background 0.3s ease, border-color 0.3s ease;
}

.login-card.light {
  background: rgba(255, 255, 255, 0.9);
  border-color: rgba(0, 0, 0, 0.08);
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.08);
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
}

.logo-text {
  font-size: 36px;
  font-weight: 700;
  margin: 0 0 8px 0;
  letter-spacing: -1px;
}

.agent-text {
  color: #fff;
  font-weight: 300;
  transition: color 0.3s ease;
}

.agent-text.light {
  color: #333;
}

.login-subtitle {
  color: rgba(255, 255, 255, 0.45);
  font-size: 14px;
  margin: 0;
  transition: color 0.3s ease;
}

.login-subtitle.light {
  color: rgba(0, 0, 0, 0.45);
}

.sso-section {
  margin-top: 16px;
}

.login-footer {
  text-align: center;
  margin-top: 24px;
}
</style>
