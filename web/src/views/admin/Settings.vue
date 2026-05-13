<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import {
  NCard,
  NTabs,
  NTabPane,
  NForm,
  NFormItem,
  NInput,
  NSwitch,
  NButton,
  NSpace,
  useMessage,
  NAlert,
  NInputNumber,
  NSelect,
  NPopconfirm,
} from 'naive-ui'
import apiClient from '@/api/client'

const message = useMessage()

const loading = ref(false)
const saving = ref(false)

const generalFormRef = ref<FormInst | null>(null)
const emailFormRef = ref<FormInst | null>(null)

const generalForm = ref({
  appName: 'AuthNas',
  appUrl: 'http://localhost:8080',
})

const securityForm = ref({
  emailVerificationRequired: false,
  signupRequiresApproval: false,
  mfaRequired: false,
  passwordMinLength: 8,
  passwordStrength: 3,
})

const emailForm = ref({
  enabled: false,
  smtpHost: '',
  smtpPort: 587,
  smtpUser: '',
  smtpPassword: '',
  fromAddress: '',
  fromName: 'AuthNas',
})

const sessionForm = ref({
  accessTokenExpiry: 15,
  refreshTokenExpiry: 7,
  maxSessionsPerUser: 5,
})

const rateLimitForm = ref({
  enabled: true,
  loginLimit: 5,
  registerLimit: 3,
  apiLimit: 60,
})

const securityOptions = [
  { label: '低 (1)', value: 1 },
  { label: '中 (2)', value: 2 },
  { label: '高 (3)', value: 3 },
  { label: '非常高 (4)', value: 4 },
]

const urlRule: FormRules = {
  appUrl: {
    required: true,
    trigger: 'blur',
    message: '请输入有效的 URL',
    validator: (_rule: any, value: string) => {
      if (!value) return false
      try {
        new URL(value)
        return true
      } catch {
        return false
      }
    },
  },
}

const smtpRules: FormRules = {
  smtpHost: {
    required: true,
    message: 'SMTP 主机不能为空',
    trigger: 'blur',
  },
  smtpUser: {
    required: true,
    message: 'SMTP 用户不能为空',
    trigger: 'blur',
  },
  fromAddress: {
    required: true,
    message: '发件人地址不能为空',
    trigger: 'blur',
  },
}

async function fetchConfig() {
  loading.value = true
  try {
    const [generalRes, securityRes, emailRes, sessionRes, ratelimitRes] = await Promise.all([
      apiClient.get('/api/admin/settings/general'),
      apiClient.get('/api/admin/settings/security'),
      apiClient.get('/api/admin/settings/email'),
      apiClient.get('/api/admin/settings/session'),
      apiClient.get('/api/admin/settings/ratelimit'),
    ])

    if (generalRes.data.success && generalRes.data.data) {
      generalForm.value.appName = generalRes.data.data.app_name || 'AuthNas'
      generalForm.value.appUrl = generalRes.data.data.app_url || 'http://localhost:8080'
    }
    if (securityRes.data.success && securityRes.data.data) {
      securityForm.value.emailVerificationRequired =
        securityRes.data.data.email_verification_required || false
      securityForm.value.signupRequiresApproval =
        securityRes.data.data.signup_requires_approval || false
      securityForm.value.mfaRequired = securityRes.data.data.mfa_required || false
      securityForm.value.passwordMinLength = securityRes.data.data.password_min_length || 8
      securityForm.value.passwordStrength = securityRes.data.data.password_strength || 0
    }
    if (emailRes.data.success && emailRes.data.data) {
      emailForm.value.enabled = emailRes.data.data.enabled || false
      emailForm.value.smtpHost = emailRes.data.data.smtp_host || ''
      emailForm.value.smtpPort = emailRes.data.data.smtp_port || 587
      emailForm.value.smtpUser = emailRes.data.data.smtp_user || ''
      emailForm.value.smtpPassword = ''
      emailForm.value.fromAddress = emailRes.data.data.from_email || ''
      emailForm.value.fromName = emailRes.data.data.from_name || 'AuthNas'
    }
    if (sessionRes.data.success && sessionRes.data.data) {
      sessionForm.value.accessTokenExpiry = sessionRes.data.data.access_token_expiry || 15
      sessionForm.value.refreshTokenExpiry = sessionRes.data.data.refresh_token_expiry || 7
      sessionForm.value.maxSessionsPerUser = sessionRes.data.data.max_sessions_per_user || 5
    }
    if (ratelimitRes.data.success && ratelimitRes.data.data) {
      rateLimitForm.value.enabled = ratelimitRes.data.data.enabled || true
      rateLimitForm.value.loginLimit = ratelimitRes.data.data.login_limit || 5
      rateLimitForm.value.registerLimit = ratelimitRes.data.data.register_limit || 3
      rateLimitForm.value.apiLimit = ratelimitRes.data.data.api_limit || 60
    }
  } catch (err) {
    console.error('获取配置失败:', err)
    message.error('获取配置失败')
  } finally {
    loading.value = false
  }
}

async function saveGeneralSettings() {
  await generalFormRef.value?.validate()
  saving.value = true
  try {
    await apiClient.post('/api/admin/settings/general', {
      app_name: generalForm.value.appName,
      app_url: generalForm.value.appUrl,
    })
    message.success('常规设置已保存')
  } catch (err: any) {
    message.error(err.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function saveSecuritySettings() {
  saving.value = true
  try {
    await apiClient.post('/api/admin/settings/security', {
      email_verification_required: securityForm.value.emailVerificationRequired,
      signup_requires_approval: securityForm.value.signupRequiresApproval,
      mfa_required: securityForm.value.mfaRequired,
      password_min_length: securityForm.value.passwordMinLength,
      password_strength: securityForm.value.passwordStrength,
    })
    message.success('安全设置已保存')
  } catch (err: any) {
    message.error(err.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function saveEmailSettings() {
  if (!emailForm.value.enabled) {
    message.warning('请先启用邮件功能')
    return
  }
  await emailFormRef.value?.validate()
  saving.value = true
  try {
    await apiClient.post('/api/admin/settings/email', {
      enabled: emailForm.value.enabled,
      smtp_host: emailForm.value.smtpHost,
      smtp_port: emailForm.value.smtpPort,
      smtp_user: emailForm.value.smtpUser,
      smtp_password: emailForm.value.smtpPassword,
      from_email: emailForm.value.fromAddress,
      from_name: emailForm.value.fromName,
    })
    message.success('邮件设置已保存')
  } catch (err: any) {
    message.error(err.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function saveSessionSettings() {
  saving.value = true
  try {
    await apiClient.post('/api/admin/settings/session', {
      access_token_expiry: sessionForm.value.accessTokenExpiry,
      refresh_token_expiry: sessionForm.value.refreshTokenExpiry,
      max_sessions_per_user: sessionForm.value.maxSessionsPerUser,
    })
    message.success('会话设置已保存')
  } catch (err: any) {
    message.error(err.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function saveRateLimitSettings() {
  saving.value = true
  try {
    await apiClient.post('/api/admin/settings/ratelimit', {
      enabled: rateLimitForm.value.enabled,
      login_limit: rateLimitForm.value.loginLimit,
      register_limit: rateLimitForm.value.registerLimit,
      api_limit: rateLimitForm.value.apiLimit,
    })
    message.success('限流设置已保存')
  } catch (err: any) {
    message.error(err.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function sendTestEmail() {
  if (!emailForm.value.enabled) {
    message.warning('请先启用邮件功能')
    return
  }
  try {
    await apiClient.post('/api/admin/settings/email/test', {
      smtpHost: emailForm.value.smtpHost,
      smtpPort: emailForm.value.smtpPort,
      smtpUser: emailForm.value.smtpUser,
      smtpPassword: emailForm.value.smtpPassword,
      fromAddress: emailForm.value.fromAddress,
    })
    message.success('测试邮件已发送，请检查收件箱')
  } catch (err: any) {
    message.error(err.response?.data?.message || '发送测试邮件失败')
  }
}

function resetToDefaults() {
  generalForm.value = { appName: 'AuthNas', appUrl: 'http://localhost:8080' }
  securityForm.value = {
    emailVerificationRequired: false,
    signupRequiresApproval: false,
    mfaRequired: false,
    passwordMinLength: 8,
    passwordStrength: 0,
  }
  emailForm.value = {
    enabled: false,
    smtpHost: '',
    smtpPort: 587,
    smtpUser: '',
    smtpPassword: '',
    fromAddress: '',
    fromName: 'AuthNas',
  }
  sessionForm.value = { accessTokenExpiry: 15, refreshTokenExpiry: 7, maxSessionsPerUser: 5 }
  rateLimitForm.value = { enabled: true, loginLimit: 5, registerLimit: 3, apiLimit: 60 }
  message.success('已重置为默认设置')
}

onMounted(() => {
  fetchConfig()
})
</script>

<template>
  <div class="page-container">
    <NCard title="系统设置">
      <template #header-extra>
        <NPopconfirm @positive-click="resetToDefaults">
          <template #trigger>
            <NButton secondary>重置默认设置</NButton>
          </template>
          确定要重置所有设置为默认值吗？
        </NPopconfirm>
      </template>

      <NTabs type="line">
        <NTabPane name="general" tab="常规">
          <NForm ref="generalFormRef" :model="generalForm" :rules="urlRule" label-placement="top">
            <NFormItem label="应用名称" path="appName">
              <NInput v-model:value="generalForm.appName" placeholder="AuthNas" />
            </NFormItem>
            <NFormItem label="应用 URL" path="appUrl">
              <NInput v-model:value="generalForm.appUrl" placeholder="http://localhost:8080" />
            </NFormItem>
          </NForm>
          <NSpace justify="end" style="margin-top: 20px">
            <NButton type="primary" :loading="saving" @click="saveGeneralSettings"
              >保存常规设置</NButton
            >
          </NSpace>
        </NTabPane>

        <NTabPane name="security" tab="安全">
          <NAlert type="info" style="margin-bottom: 16px">
            安全设置需要后端配置才能在生产环境使用。
          </NAlert>
          <NForm :model="securityForm" label-placement="top">
            <NFormItem label="需要邮箱验证">
              <NSwitch v-model:value="securityForm.emailVerificationRequired" />
            </NFormItem>
            <NFormItem label="注册需要审批">
              <NSwitch v-model:value="securityForm.signupRequiresApproval" />
            </NFormItem>
            <NFormItem label="所有用户需要 MFA">
              <NSwitch v-model:value="securityForm.mfaRequired" />
            </NFormItem>
            <NFormItem label="密码最小长度">
              <NInputNumber v-model:value="securityForm.passwordMinLength" :min="6" :max="128" />
            </NFormItem>
            <NFormItem label="密码强度要求">
              <NSelect v-model:value="securityForm.passwordStrength" :options="securityOptions" />
            </NFormItem>
          </NForm>
          <NSpace justify="end" style="margin-top: 20px">
            <NButton type="primary" :loading="saving" @click="saveSecuritySettings"
              >保存安全设置</NButton
            >
          </NSpace>
        </NTabPane>

        <NTabPane name="email" tab="邮件">
          <NAlert type="info" style="margin-bottom: 16px">
            邮件配置需要后端 SMTP 设置才能在生产环境使用。
          </NAlert>
          <NForm ref="emailFormRef" :model="emailForm" :rules="smtpRules" label-placement="top">
            <NFormItem label="启用邮件">
              <NSwitch v-model:value="emailForm.enabled" />
            </NFormItem>
            <NFormItem label="SMTP 主机" path="smtpHost">
              <NInput v-model:value="emailForm.smtpHost" placeholder="smtp.example.com" />
            </NFormItem>
            <NFormItem label="SMTP 端口">
              <NInputNumber v-model:value="emailForm.smtpPort" :min="1" :max="65535" />
            </NFormItem>
            <NFormItem label="SMTP 用户" path="smtpUser">
              <NInput v-model:value="emailForm.smtpUser" placeholder="user@example.com" />
            </NFormItem>
            <NFormItem label="SMTP 密码">
              <NInput
                v-model:value="emailForm.smtpPassword"
                type="password"
                placeholder="留空以保持当前密码"
              />
            </NFormItem>
            <NFormItem label="发件人地址" path="fromAddress">
              <NInput v-model:value="emailForm.fromAddress" placeholder="noreply@example.com" />
            </NFormItem>
            <NFormItem label="发件人名称">
              <NInput v-model:value="emailForm.fromName" placeholder="AuthNas" />
            </NFormItem>
          </NForm>
          <NSpace justify="end" style="margin-top: 20px">
            <NButton @click="sendTestEmail">发送测试邮件</NButton>
            <NButton type="primary" :loading="saving" @click="saveEmailSettings"
              >保存邮件设置</NButton
            >
          </NSpace>
        </NTabPane>

        <NTabPane name="session" tab="会话">
          <NForm :model="sessionForm" label-placement="top">
            <NFormItem label="访问令牌过期时间（分钟）">
              <NInputNumber v-model:value="sessionForm.accessTokenExpiry" :min="5" :max="1440" />
            </NFormItem>
            <NFormItem label="刷新令牌过期时间（天）">
              <NInputNumber v-model:value="sessionForm.refreshTokenExpiry" :min="1" :max="90" />
            </NFormItem>
            <NFormItem label="每个用户最大会话数">
              <NInputNumber v-model:value="sessionForm.maxSessionsPerUser" :min="1" :max="100" />
            </NFormItem>
          </NForm>
          <NSpace justify="end" style="margin-top: 20px">
            <NButton type="primary" :loading="saving" @click="saveSessionSettings"
              >保存会话设置</NButton
            >
          </NSpace>
        </NTabPane>

        <NTabPane name="ratelimit" tab="限流">
          <NForm :model="rateLimitForm" label-placement="top">
            <NFormItem label="启用限流">
              <NSwitch v-model:value="rateLimitForm.enabled" />
            </NFormItem>
            <NFormItem label="登录限制（每分钟）">
              <NInputNumber
                v-model:value="rateLimitForm.loginLimit"
                :min="1"
                :max="100"
                :disabled="!rateLimitForm.enabled"
              />
            </NFormItem>
            <NFormItem label="注册限制（每分钟）">
              <NInputNumber
                v-model:value="rateLimitForm.registerLimit"
                :min="1"
                :max="100"
                :disabled="!rateLimitForm.enabled"
              />
            </NFormItem>
            <NFormItem label="API 限制（每分钟）">
              <NInputNumber
                v-model:value="rateLimitForm.apiLimit"
                :min="10"
                :max="1000"
                :disabled="!rateLimitForm.enabled"
              />
            </NFormItem>
          </NForm>
          <NSpace justify="end" style="margin-top: 20px">
            <NButton type="primary" :loading="saving" @click="saveRateLimitSettings"
              >保存限流设置</NButton
            >
          </NSpace>
        </NTabPane>
      </NTabs>
    </NCard>
  </div>
</template>

<style scoped>
.page-container {
  padding: 40px 0;
}
</style>
