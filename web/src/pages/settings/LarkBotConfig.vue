<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { larkBotApi } from '@/api'

const message = useMessage()
const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)

const form = reactive({
  bot_enabled: false,
  app_id: '',
  app_secret: '',
  default_webhook: '',
  verification_token: '',
  encrypt_key: '',
})

async function fetchConfig() {
  loading.value = true
  try {
    const res = await larkBotApi.getConfig()
    if (res.data.data) {
      const d = res.data.data
      form.bot_enabled = d.bot_enabled
      form.app_id = d.app_id || ''
      form.app_secret = d.app_secret || ''
      form.default_webhook = d.default_webhook || ''
      form.verification_token = d.verification_token || ''
      form.encrypt_key = d.encrypt_key || ''
    }
  } catch (err: any) {
    message.error(err.message)
  } finally {
    loading.value = false
  }
}

async function saveConfig() {
  saving.value = true
  try {
    await larkBotApi.updateConfig({ ...form })
    message.success(t('common.savedSuccess'))
  } catch (err: any) {
    message.error(err.message)
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  fetchConfig()
})
</script>

<template>
  <n-spin :show="loading">
    <div style="max-width: 640px; margin: 0 auto; padding: 24px 0">
      <n-form label-placement="top">
        <n-form-item :label="t('settings.larkBotEnabled')">
          <n-switch v-model:value="form.bot_enabled" />
        </n-form-item>

        <n-form-item :label="t('settings.larkAppId')">
          <n-input v-model:value="form.app_id" :placeholder="t('settings.larkAppIdPlaceholder')" />
        </n-form-item>

        <n-form-item :label="t('settings.larkAppSecret')">
          <n-input
            v-model:value="form.app_secret"
            type="password"
            show-password-on="click"
            :placeholder="t('settings.larkAppSecretPlaceholder')"
          />
        </n-form-item>

        <n-form-item :label="t('settings.larkDefaultWebhook')">
          <n-input v-model:value="form.default_webhook" :placeholder="t('settings.larkWebhookPlaceholder')" />
        </n-form-item>

        <n-form-item :label="t('settings.larkVerificationToken')">
          <n-input v-model:value="form.verification_token" :placeholder="t('settings.larkTokenPlaceholder')" />
        </n-form-item>

        <n-form-item :label="t('settings.larkEncryptKey')">
          <n-input
            v-model:value="form.encrypt_key"
            type="password"
            show-password-on="click"
            :placeholder="t('settings.larkEncryptKeyPlaceholder')"
          />
        </n-form-item>

        <n-alert type="info" :bordered="false" style="margin-bottom: 16px">
          {{ t('settings.larkCallbackHint') }}: <n-text code>/lark/event</n-text>
        </n-alert>

        <n-button type="primary" :loading="saving" @click="saveConfig">
          {{ t('common.save') }}
        </n-button>
      </n-form>
    </div>
  </n-spin>
</template>
