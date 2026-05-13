<script setup lang="ts">
import {
  NConfigProvider,
  NMessageProvider,
  NDialogProvider,
  NNotificationProvider,
  type GlobalThemeOverrides,
} from 'naive-ui'
import { RouterView } from 'vue-router'
import { zhCN } from 'naive-ui'

const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#6366f1',
    primaryColorHover: '#818cf8',
    primaryColorPressed: '#4f46e5',
    primaryColorSuppl: '#818cf8',
    borderRadius: '8px',
    fontFamily:
      "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'PingFang SC', 'Microsoft YaHei', sans-serif",
  },
  Button: {
    fontWeight: '500',
    borderRadiusMedium: '8px',
  },
  Card: {
    borderRadius: '12px',
  },
  Input: {
    borderRadius: '8px',
  },
}
</script>

<template>
  <NConfigProvider :locale="zhCN" :theme-overrides="themeOverrides">
    <NMessageProvider>
      <NDialogProvider>
        <NNotificationProvider>
          <RouterView />
        </NNotificationProvider>
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}
</style>
