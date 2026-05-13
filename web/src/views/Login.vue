<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, NAlert, NCheckbox, useMessage } from 'naive-ui'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const authStore = useAuthStore()

const username = ref(localStorage.getItem('remembered_username') || '')
const password = ref('')
const rememberMe = ref(!!localStorage.getItem('remembered_username'))
const loading = ref(false)
const error = ref('')
const passkeyLoading = ref(false)

const showPasskeyLoginOption = computed(() => {
  return username.value && authStore.isAuthenticated === false
})

onMounted(async () => {
  try {
    await authApi.getCsrfToken()
  } catch {}
})

async function handleLogin() {
  if (!username.value || !password.value) {
    error.value = '请输入用户名和密码'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = (await authApi.login({
      input: username.value,
      password: password.value,
    })) as any
    if (response.data.success && response.data.data?.mfaRequired) {
      router.push({ path: '/mfa', query: { token: response.data.data?.mfaToken || '' } })
      return
    }

    if (response.data.success && response.data.data) {
      if (rememberMe.value) {
        localStorage.setItem('remembered_username', username.value)
      } else {
        localStorage.removeItem('remembered_username')
      }
      authStore.setTokens(response.data.data.accessToken, response.data.data.expiresAt)
      authStore.setUser(response.data.data.user)
      message.success('登录成功')
      const redirect = route.query.redirect as string
      router.push(redirect || '/profile')
    }
  } catch (err: any) {
    error.value = err.response?.data?.message || '登录失败'
  } finally {
    loading.value = false
  }
}

async function handlePasskeyLogin() {
  if (!username.value) {
    error.value = '请先输入用户名'
    return
  }

  passkeyLoading.value = true
  error.value = ''

  try {
    const startResponse = await authApi.passkeyStart({ username: username.value })
    if (!startResponse.data.success || !startResponse.data.data) {
      throw new Error('无法开始通行密钥认证')
    }

    const challenge = startResponse.data.data.challenge

    const credential = await navigator.credentials.get({
      publicKey: {
        challenge: Uint8Array.from(atob(challenge), (c) => c.charCodeAt(0)),
        userVerification: 'preferred',
      },
    })

    const endResponse = await authApi.passkeyEnd({
      credentialId: btoa(String.fromCharCode(...new Uint8Array((credential as any).rawId))),
      challenge: challenge,
      response: JSON.stringify(credential),
    })

    if (
      endResponse.data.success &&
      endResponse.data.data &&
      'mfaRequired' in endResponse.data.data
    ) {
      const mfaToken = (endResponse.data.data as any).mfaToken || ''
      const userId = (endResponse.data.data as any).userId || ''
      router.push({ path: '/mfa', query: { token: mfaToken, userId } })
      return
    }

    if (
      endResponse.data.success &&
      endResponse.data.data &&
      'accessToken' in endResponse.data.data
    ) {
      if (rememberMe.value) {
        localStorage.setItem('remembered_username', username.value)
      } else {
        localStorage.removeItem('remembered_username')
      }
      authStore.setTokens(endResponse.data.data.accessToken, endResponse.data.data.expiresAt)
      if ('user' in endResponse.data.data && endResponse.data.data.user) {
        authStore.setUser(endResponse.data.data.user)
      }
      message.success('登录成功')
      const redirect = route.query.redirect as string
      router.push(redirect || '/profile')
    }
  } catch (err: any) {
    if (err.name === 'NotAllowedError') {
      error.value = '通行密钥认证已取消'
    } else {
      error.value = err.response?.data?.message || err.message || '通行密钥认证失败'
    }
  } finally {
    passkeyLoading.value = false
  }
}
</script>

<template>
  <div class="auth-container">
    <NCard class="auth-card" :bordered="false">
      <template #header>
        <div class="page-header">
          <h1>欢迎回来</h1>
          <p>请登录您的账户</p>
        </div>
      </template>

      <NAlert
        v-if="error"
        type="error"
        role="alert"
        aria-live="assertive"
        style="margin-bottom: 20px"
      >
        {{ error }}
      </NAlert>

      <NForm aria-label="登录表单" @submit.prevent="handleLogin">
        <NFormItem label="用户名" path="username">
          <NInput
            v-model:value="username"
            placeholder="请输入用户名"
            size="large"
            aria-label="用户名"
          />
        </NFormItem>

        <NFormItem label="密码" path="password">
          <NInput
            v-model:value="password"
            type="password"
            placeholder="请输入密码"
            size="large"
            aria-label="密码"
            @keydown.enter="handleLogin"
          />
        </NFormItem>

        <NFormItem>
          <NCheckbox v-model:checked="rememberMe">记住我</NCheckbox>
        </NFormItem>

        <NButton type="primary" attr-type="submit" :loading="loading" block size="large">
          登录
        </NButton>

        <div v-if="showPasskeyLoginOption || username" class="passkey-section">
          <div class="divider">
            <span>或者</span>
          </div>
          <NButton
            block
            size="large"
            :loading="passkeyLoading"
            :disabled="!username"
            @click="handlePasskeyLogin"
          >
            使用通行密钥登录
          </NButton>
        </div>
      </NForm>

      <div class="links">
        <router-link to="/register">创建账户</router-link>
        <router-link to="/reset-password">忘记密码？</router-link>
      </div>
    </NCard>
  </div>
</template>

<style scoped>
.auth-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 80px);
  padding: 40px 20px;
}

.auth-card {
  width: 100%;
  max-width: 420px;
}

.page-header {
  text-align: center;
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 28px;
  margin-bottom: 8px;
  color: var(--text-h);
}

.page-header p {
  color: var(--text);
  font-size: 15px;
}

.links {
  display: flex;
  justify-content: space-between;
  margin-top: 24px;
  font-size: 14px;
}

.passkey-section {
  margin-top: 8px;
}

.divider {
  display: flex;
  align-items: center;
  margin: 20px 0;
  color: #9ca3af;
  font-size: 14px;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--border);
}

.divider span {
  padding: 0 12px;
}
</style>
