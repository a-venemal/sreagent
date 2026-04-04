<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { aiApi } from '@/api'

const message = useMessage()
const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)

const form = reactive({
  enabled: false,
  provider: 'openai',
  api_key: '',
  base_url: '',
  model: '',
})

const providerOptions = [
  { label: 'OpenAI', value: 'openai' },
  { label: 'Azure OpenAI', value: 'azure' },
  { label: 'Ollama (Local)', value: 'ollama' },
  { label: 'Custom / Compatible', value: 'custom' },
]

async function fetchConfig() {
  loading.value = true
  try {
    const res = await aiApi.getConfig()
    if (res.data.data) {
      const d = res.data.data
      form.enabled = d.enabled
      form.provider = d.provider || 'openai'
      form.api_key = d.api_key || ''
      form.base_url = d.base_url || ''
      form.model = d.model || ''
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
    await aiApi.updateConfig({ ...form })
    message.success(t('common.savedSuccess'))
  } catch (err: any) {
    message.error(err.message)
  } finally {
    saving.value = false
  }
}

async function testConnection() {
  testing.value = true
  try {
    const res = await aiApi.testConnection()
    if (res.data.data?.success) {
      message.success(t('settings.aiTestSuccess'))
    } else {
      message.error(res.data.data?.message || t('settings.aiTestFailed'))
    }
  } catch (err: any) {
    message.error(err.message)
  } finally {
    testing.value = false
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
        <n-form-item :label="t('settings.aiEnabled')">
          <n-switch v-model:value="form.enabled" />
        </n-form-item>

        <n-form-item :label="t('settings.aiProvider')">
          <n-select v-model:value="form.provider" :options="providerOptions" style="width: 100%" />
        </n-form-item>

        <n-form-item :label="t('settings.aiApiKey')">
          <n-input
            v-model:value="form.api_key"
            type="password"
            show-password-on="click"
            :placeholder="t('settings.aiApiKeyPlaceholder')"
          />
        </n-form-item>

        <n-form-item :label="t('settings.aiBaseUrl')">
          <n-input v-model:value="form.base_url" :placeholder="t('settings.aiBaseUrlPlaceholder')" />
        </n-form-item>

        <n-form-item :label="t('settings.aiModel')">
          <n-input v-model:value="form.model" :placeholder="t('settings.aiModelPlaceholder')" />
        </n-form-item>

        <n-space>
          <n-button type="primary" :loading="saving" @click="saveConfig">
            {{ t('common.save') }}
          </n-button>
          <n-button :loading="testing" @click="testConnection">
            {{ t('settings.testConnection') }}
          </n-button>
        </n-space>
      </n-form>
    </div>
  </n-spin>
</template>
