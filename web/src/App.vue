<script setup lang="ts">
import {
  NConfigProvider,
  NMessageProvider,
  NDialogProvider,
  NNotificationProvider,
  darkTheme,
} from 'naive-ui'
import type { GlobalThemeOverrides } from 'naive-ui'
import { ref, provide, watch, onMounted, computed } from 'vue'

const savedTheme = localStorage.getItem('sre-theme')
const isDark = ref(savedTheme ? savedTheme === 'dark' : true)
const theme = computed(() => isDark.value ? darkTheme : null)

// Light mode theme overrides: tell Naive UI to use white card backgrounds
const lightOverrides: GlobalThemeOverrides = {
  Card: {
    color: '#ffffff',
    colorEmbedded: '#f7f7f7',
  },
  DataTable: {
    tdColor: '#ffffff',
    thColor: '#fafafa',
  },
  Layout: {
    color: '#f5f5f5',
    siderColor: '#ffffff',
    headerColor: '#ffffff',
  },
  Modal: {
    color: '#ffffff',
  },
  Drawer: {
    color: '#ffffff',
  },
  Select: {
    peers: {
      InternalSelectMenu: {
        color: '#ffffff',
      },
    },
  },
}

const themeOverrides = computed<GlobalThemeOverrides>(() =>
  isDark.value ? {} : lightOverrides
)

function applyBodyClass(dark: boolean) {
  if (dark) {
    document.body.classList.remove('light-theme')
  } else {
    document.body.classList.add('light-theme')
  }
}

onMounted(() => {
  applyBodyClass(isDark.value)
})

watch(isDark, (val) => {
  localStorage.setItem('sre-theme', val ? 'dark' : 'light')
  applyBodyClass(val)
})

provide('toggleTheme', () => {
  isDark.value = !isDark.value
})
provide('isDark', isDark)
</script>

<template>
  <NConfigProvider :theme="theme" :theme-overrides="themeOverrides">
    <NMessageProvider>
      <NDialogProvider>
        <NNotificationProvider>
          <router-view />
        </NNotificationProvider>
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>

<style>
body {
  margin: 0;
  padding: 0;
}
</style>
