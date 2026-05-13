<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NResult, NButton, NSpace } from 'naive-ui'

const route = useRoute()
const router = useRouter()

const status = computed(() => {
  const s = route.query.status as string
  return s ? parseInt(s) : 500
})

const statusString = computed(() => {
  const s = status.value
  if (s === 400) return 'error'
  if (s === 401) return 'error'
  return s.toString() as 'info' | 'success' | 'warning' | 'error' | '500' | '404' | '403' | '418'
})

const title = computed(() => {
  const s = status.value
  if (s === 400) return '请求错误'
  if (s === 401) return '未授权'
  if (s === 403) return '禁止访问'
  if (s === 404) return '页面不存在'
  if (s === 500) return '服务器内部错误'
  return '错误'
})

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

const description = computed(() => {
  const msg = route.query.message as string
  if (!msg) return '发生了意外的错误'
  return escapeHtml(msg)
})

function goHome() {
  router.push('/')
}

function goLogin() {
  router.push('/login')
}
</script>

<template>
  <div class="error-container">
    <NResult :status="statusString" :title="title" :description="description">
      <template #footer>
        <NSpace justify="center">
          <NButton @click="goHome">返回首页</NButton>
          <NButton type="primary" @click="goLogin">前往登录</NButton>
        </NSpace>
      </template>
    </NResult>
  </div>
</template>

<style scoped>
.error-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: var(--bg);
}
</style>
