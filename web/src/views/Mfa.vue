<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NCard, NInput, NButton, NSpace, NAlert, useMessage } from 'naive-ui'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const authStore = useAuthStore()

const code = ref('')
const loading = ref(false)
const error = ref('')
const mfaToken = ref('')

onMounted(() => {
  mfaToken.value = (route.query.token as string) || ''
  if (!mfaToken.value) {
    error.value = 'MFA 令牌缺失，请重新登录'
  }
})

async function handleSubmit() {
  if (!code.value) {
    error.value = '请输入验证码'
    return
  }

  if (code.value.length !== 6) {
    error.value = '请输入 6 位数字验证码'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await authApi.totpVerify({ token: code.value, mfaToken: mfaToken.value })
    if (response.data.success && response.data.data) {
      authStore.setTokens(response.data.data.accessToken, response.data.data.expiresAt)
      if ('user' in response.data.data && response.data.data.user) {
        authStore.setUser(response.data.data.user)
      }
      message.success('MFA 验证成功')
      router.push('/profile')
    }
  } catch (err: any) {
    error.value = err.response?.data?.message || '验证码错误，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-container">
    <NCard class="auth-card" :bordered="false">
      <template #header>
        <div class="page-header">
          <h1>多因素认证</h1>
          <p>请输入您的验证码</p>
        </div>
      </template>

      <NAlert type="info" role="status" style="margin-bottom: 20px">
        请输入您身份验证器应用中的 6 位数字验证码
      </NAlert>

      <NAlert
        v-if="error"
        type="error"
        role="alert"
        aria-live="assertive"
        style="margin-bottom: 16px"
      >
        {{ error }}
      </NAlert>

      <NSpace vertical>
        <NInput
          v-model:value="code"
          placeholder="请输入 6 位验证码"
          size="large"
          maxlength="6"
          aria-label="验证码"
          autocomplete="one-time-code"
          inputmode="numeric"
          @keydown.enter="handleSubmit"
        />
        <NSpace justify="end">
          <NButton type="primary" :loading="loading" size="large" @click="handleSubmit">
            验证
          </NButton>
        </NSpace>
      </NSpace>

      <div class="links">
        <router-link to="/login">返回登录</router-link>
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
  font-size: 24px;
  margin-bottom: 8px;
  color: var(--text-h);
}

.page-header p {
  color: var(--text);
  font-size: 15px;
}

.links {
  margin-top: 24px;
  text-align: center;
  font-size: 14px;
}
</style>
