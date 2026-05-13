<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, NAlert, useMessage, NProgress } from 'naive-ui'
import { authApi } from '@/api/auth'

const router = useRouter()
const message = useMessage()

const email = ref('')
const username = ref('')
const password = ref('')
const invitationCode = ref('')
const loading = ref(false)
const error = ref('')
const policyLoading = ref(true)

const publicConfig = ref({
  signupRequiresApproval: false,
  emailVerification: false,
  passwordMinLength: 8,
  passwordStrength: 3,
})

const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/
const usernameRegex = /^[a-zA-Z0-9_]{3,20}$/

function validateEmail(email: string): boolean {
  return emailRegex.test(email)
}

function validateUsername(username: string): boolean {
  return usernameRegex.test(username)
}

function calculatePasswordStrength(pwd: string): { score: number; label: string; color: string } {
  if (!pwd) return { score: 0, label: '', color: '#d0d0d0' }

  let score = 0
  if (pwd.length >= 8) score++
  if (pwd.length >= 12) score++
  if (/[a-z]/.test(pwd) && /[A-Z]/.test(pwd)) score++
  if (/\d/.test(pwd)) score++
  if (/[!@#$%^&*(),.?":{}|<>]/.test(pwd)) score++

  const levels = [
    { score: 0, label: '非常弱', color: '#f44336' },
    { score: 1, label: '弱', color: '#ff9800' },
    { score: 2, label: '一般', color: '#ffeb3b' },
    { score: 3, label: '良好', color: '#8bc34a' },
    { score: 4, label: '强', color: '#4caf50' },
  ]

  return levels[Math.min(score, 4)]
}

const passwordStrength = computed(() => calculatePasswordStrength(password.value))
const requiresEmail = computed(
  () => publicConfig.value.emailVerification || publicConfig.value.signupRequiresApproval
)
const requiresInvitation = computed(() => publicConfig.value.signupRequiresApproval)
const passwordMinLengthLabel = computed(() => publicConfig.value.passwordMinLength || 8)
const passwordStrengthThreshold = computed(() => publicConfig.value.passwordStrength || 0)

onMounted(async () => {
  try {
    const [publicConfigResponse, csrfResponse] = await Promise.all([
      authApi.getPublicConfig(),
      authApi.getCsrfToken(),
    ])

    if (publicConfigResponse.data.success && publicConfigResponse.data.data) {
      publicConfig.value.signupRequiresApproval =
        publicConfigResponse.data.data.signup_requires_approval || false
      publicConfig.value.emailVerification =
        publicConfigResponse.data.data.email_verification || false
      publicConfig.value.passwordMinLength = publicConfigResponse.data.data.password_min_length || 8
      publicConfig.value.passwordStrength = publicConfigResponse.data.data.password_strength || 3
    }

    if (csrfResponse.data.success && csrfResponse.data.data) {
      const token = csrfResponse.data.data.token
      if (token) {
        document.cookie = `csrf_token=${token}; path=/`
      }
    }
  } catch {
    error.value = '加载注册策略失败，请刷新后重试'
  } finally {
    policyLoading.value = false
  }
})

async function handleRegister() {
  if (
    (!requiresEmail.value && !username.value) ||
    !password.value ||
    (requiresEmail.value && !email.value) ||
    !username.value
  ) {
    error.value = '请填写所有必填字段'
    return
  }

  if (requiresEmail.value && !validateEmail(email.value)) {
    error.value = '请输入有效的邮箱地址'
    return
  }

  if (!validateUsername(username.value)) {
    error.value = '用户名必须为 3-20 个字符，只能包含字母、数字和下划线'
    return
  }

  if (password.value.length < passwordMinLengthLabel.value) {
    error.value = `密码长度不能少于 ${passwordMinLengthLabel.value} 位`
    return
  }

  if (passwordStrength.value.score < passwordStrengthThreshold.value) {
    error.value = '密码强度太弱，请使用更安全的密码'
    return
  }

  if (requiresInvitation.value && !invitationCode.value) {
    error.value = '当前注册需要有效邀请码'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await authApi.register({
      email: email.value,
      username: username.value,
      password: password.value,
      inviteId: invitationCode.value || undefined,
    })
    if (response.data.success) {
      message.success('注册成功')
      router.push('/login')
    }
  } catch (err: any) {
    error.value = err.response?.data?.message || '注册失败'
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
          <h1>创建账户</h1>
          <p>注册新账户</p>
        </div>
      </template>

      <NAlert v-if="policyLoading" type="info" role="status" style="margin-bottom: 20px">
        正在加载注册策略...
      </NAlert>

      <NAlert
        v-if="error"
        type="error"
        role="alert"
        aria-live="assertive"
        style="margin-bottom: 20px"
      >
        {{ error }}
      </NAlert>

      <NForm aria-label="注册表单" @submit.prevent="handleRegister">
        <NAlert v-if="requiresInvitation" type="warning" role="status" style="margin-bottom: 20px">
          当前系统已开启注册审批，必须提供有效邀请码才能注册。
        </NAlert>

        <NAlert
          v-else-if="publicConfig.emailVerification"
          type="info"
          role="status"
          style="margin-bottom: 20px"
        >
          注册后需要先完成邮箱验证才能正常使用账户。
        </NAlert>

        <NFormItem :label="requiresEmail ? '邮箱' : '邮箱（可选）'" path="email">
          <NInput
            v-model:value="email"
            placeholder="请输入邮箱"
            size="large"
            aria-label="邮箱"
            type="email"
            autocomplete="email"
          />
        </NFormItem>

        <NFormItem label="用户名" path="username">
          <NInput
            v-model:value="username"
            placeholder="请输入用户名"
            size="large"
            aria-label="用户名"
            autocomplete="username"
          />
        </NFormItem>

        <NFormItem label="密码" path="password">
          <NInput
            v-model:value="password"
            type="password"
            placeholder="请输入密码"
            size="large"
            aria-label="密码"
            autocomplete="new-password"
            @keydown.enter="handleRegister"
          />
        </NFormItem>

        <div class="password-hint">密码至少 {{ passwordMinLengthLabel }} 位。</div>

        <div v-if="password" class="password-strength">
          <div class="strength-bar">
            <NProgress
              :percentage="(passwordStrength.score / 4) * 100"
              :color="passwordStrength.color"
              :show-indicator="false"
              :height="6"
            />
          </div>
          <span class="strength-label" :style="{ color: passwordStrength.color }">
            {{ passwordStrength.label }}
          </span>
        </div>

        <NFormItem
          :label="requiresInvitation ? '邀请码（必填）' : '邀请码（可选）'"
          path="invitationCode"
        >
          <NInput v-model:value="invitationCode" placeholder="请输入邀请码" size="large" />
        </NFormItem>

        <NButton
          type="primary"
          attr-type="submit"
          :loading="loading"
          :disabled="policyLoading"
          block
          size="large"
        >
          注册
        </NButton>
      </NForm>

      <div class="links">
        <router-link to="/login">已有账户？立即登录</router-link>
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

.password-strength {
  margin-bottom: 20px;
}

.password-hint {
  margin-bottom: 12px;
  color: var(--text-color-2, #666);
  font-size: 13px;
}

.strength-bar {
  margin-bottom: 6px;
}

.strength-label {
  font-size: 13px;
}

.links {
  margin-top: 24px;
  text-align: center;
  font-size: 14px;
}
</style>
