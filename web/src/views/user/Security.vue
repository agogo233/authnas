<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import {
  NCard,
  NTabs,
  NTabPane,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NAlert,
  NModal,
  NImage,
  useMessage,
  NProgress,
  NPopconfirm,
} from 'naive-ui'
import { userApi, passkeyApi, totpApi } from '@/api/auth'
import type { Passkey, Session } from '@/types'

const message = useMessage()

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordLoading = ref(false)

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

const passwordStrength = computed(() => calculatePasswordStrength(newPassword.value))

async function handlePasswordChange() {
  if (!oldPassword.value || !newPassword.value || !confirmPassword.value) {
    message.error('请填写所有字段')
    return
  }

  if (newPassword.value !== confirmPassword.value) {
    message.error('两次输入的密码不一致')
    return
  }

  if (passwordStrength.value.score < 3) {
    message.error('密码强度太弱')
    return
  }

  passwordLoading.value = true
  try {
    await userApi.updatePassword({
      oldPassword: oldPassword.value,
      newPassword: newPassword.value,
    })
    message.success('密码修改成功')
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (err: any) {
    message.error(err.response?.data?.message || '修改密码失败')
  } finally {
    passwordLoading.value = false
  }
}

const totpEnabled = ref(false)
const totpLoading = ref(false)
const showTotpSetup = ref(false)
const totpSecret = ref('')
const totpUrl = ref('')
const totpVerifyCode = ref('')
const totpVerifyLoading = ref(false)
const showDisableTotpModal = ref(false)
const disableTotpCode = ref('')

async function checkTotpStatus() {
  try {
    const response = await userApi.getMe()
    if (response.data.success && response.data.data) {
      totpEnabled.value = response.data.data.hasTotp ?? false
    }
  } catch {
    totpEnabled.value = false
  }
}

async function handleEnableTotp() {
  totpLoading.value = true
  try {
    const response = await totpApi.register()
    if (response.data.success && response.data.data) {
      totpSecret.value = response.data.data.secret
      totpUrl.value = response.data.data.qr_code_uri
      showTotpSetup.value = true
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '启用 TOTP 失败')
  } finally {
    totpLoading.value = false
  }
}

async function handleVerifyTotp() {
  if (!totpVerifyCode.value || totpVerifyCode.value.length !== 6) {
    message.error('请输入 6 位数字验证码')
    return
  }

  totpVerifyLoading.value = true
  try {
    await totpApi.verify({ token: totpVerifyCode.value })
    message.success('TOTP 启用成功')
    showTotpSetup.value = false
    totpVerifyCode.value = ''
    totpEnabled.value = true
    await loadPasskeys()
  } catch (err: any) {
    message.error(err.response?.data?.message || '验证码错误')
  } finally {
    totpVerifyLoading.value = false
  }
}

async function handleDisableTotp() {
  showDisableTotpModal.value = true
  disableTotpCode.value = ''
}

async function confirmDisableTotp() {
  if (!disableTotpCode.value || disableTotpCode.value.length !== 6) {
    message.error('请输入 6 位数字验证码')
    return
  }
  try {
    await totpApi.delete({ token: disableTotpCode.value })
    message.success('TOTP 已禁用')
    showDisableTotpModal.value = false
    disableTotpCode.value = ''
    totpEnabled.value = false
  } catch (err: any) {
    message.error(err.response?.data?.message || '禁用 TOTP 失败')
  }
}

const passkeys = ref<Passkey[]>([])
const passkeyLoading = ref(false)
const passkeyName = ref('')
const showPasskeySetup = ref(false)
const registeringPasskey = ref(false)

async function loadPasskeys() {
  passkeyLoading.value = true
  try {
    const response = await passkeyApi.getPasskeys()
    if (response.data.success && response.data.data) {
      passkeys.value = response.data.data
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '加载通行密钥失败')
  } finally {
    passkeyLoading.value = false
  }
}

async function handleRegisterPasskey() {
  registeringPasskey.value = true
  try {
    const startResponse = await passkeyApi.registrationStart()
    if (!startResponse.data.success || !startResponse.data.data) {
      throw new Error('无法开始通行密钥注册')
    }

    const { challenge, options: optionsStr } = startResponse.data.data
    const options = JSON.parse(optionsStr)

    const credential = await navigator.credentials.create({
      publicKey: {
        challenge: Uint8Array.from(atob(challenge), (c) => c.charCodeAt(0)),
        rp: options.rp,
        user: {
          id: Uint8Array.from(atob(options.user.id), (c) => c.charCodeAt(0)),
          name: options.user.name,
          displayName: options.user.displayName,
        },
        pubKeyCredParams: options.pubKeyCredParams,
        timeout: options.timeout,
        attestation: options.attestation,
        excludeCredentials:
          options.excludeCredentials?.map((cred: any) => ({
            id: Uint8Array.from(atob(cred.id), (c) => c.charCodeAt(0)),
            type: cred.type,
            transports: cred.transports,
          })) || [],
        authenticatorSelection: options.authenticatorSelection,
      },
    })

    if (!credential) {
      throw new Error('无法创建凭证')
    }

    const credentialJson = JSON.stringify(credential)
    const endResponse = await passkeyApi.registrationEnd({
      challenge,
      options: credentialJson,
      name: passkeyName.value || '通行密钥',
    })

    if (endResponse.data.success) {
      message.success('通行密钥注册成功')
      showPasskeySetup.value = false
      passkeyName.value = ''
      await loadPasskeys()
    }
  } catch (err: any) {
    if (err.name === 'NotAllowedError') {
      message.error('通行密钥注册已取消')
    } else {
      message.error(err.message || '注册通行密钥失败')
    }
  } finally {
    registeringPasskey.value = false
  }
}

async function handleDeletePasskey(id: string) {
  try {
    await passkeyApi.deletePasskey(id)
    message.success('通行密钥已删除')
    await loadPasskeys()
  } catch (err: any) {
    message.error(err.response?.data?.message || '删除通行密钥失败')
  }
}

const sessions = ref<Session[]>([])
const sessionLoading = ref(false)

async function loadSessions() {
  try {
    const response = await userApi.getSessions()
    if (response.data.success && response.data.data) {
      sessions.value = response.data.data
    }
  } catch {
    sessions.value = []
  }
}

async function handleRevokeSession(sessionId: string) {
  try {
    await userApi.deleteSession(sessionId)
    message.success('会话已撤销')
    await loadSessions()
  } catch (err: any) {
    message.error(err.response?.data?.message || '撤销会话失败')
  }
}

async function handleRevokeAllSessions() {
  sessionLoading.value = true
  try {
    await userApi.deleteAllSessions()
    message.success('所有会话已撤销')
    await loadSessions()
  } catch (err: any) {
    message.error(err.response?.data?.message || '撤销所有会话失败')
  } finally {
    sessionLoading.value = false
  }
}

onMounted(async () => {
  await checkTotpStatus()
  await loadPasskeys()
  await loadSessions()
})
</script>

<template>
  <div class="page-container">
    <NCard title="安全设置" :bordered="false">
      <NTabs type="line" animated>
        <NTabPane name="password" tab="修改密码">
          <NForm
            :show-feedback="false"
            label-placement="left"
            label-width="140"
            @submit.prevent="handlePasswordChange"
          >
            <NFormItem label="当前密码">
              <NInput v-model:value="oldPassword" type="password" placeholder="请输入当前密码" />
            </NFormItem>
            <NFormItem label="新密码">
              <NInput v-model:value="newPassword" type="password" placeholder="请输入新密码" />
            </NFormItem>
            <div v-if="newPassword" class="password-strength">
              <NProgress
                :percentage="(passwordStrength.score / 4) * 100"
                :color="passwordStrength.color"
                :show-indicator="false"
                :height="6"
              />
              <span class="strength-label" :style="{ color: passwordStrength.color }">
                {{ passwordStrength.label }}
              </span>
            </div>
            <NFormItem label="确认新密码">
              <NInput
                v-model:value="confirmPassword"
                type="password"
                placeholder="请再次输入新密码"
              />
            </NFormItem>
            <NSpace justify="end">
              <NButton type="primary" attr-type="submit" :loading="passwordLoading"
                >更新密码</NButton
              >
            </NSpace>
          </NForm>
        </NTabPane>

        <NTabPane name="mfa" tab="两步验证">
          <div class="section">
            <NAlert
              v-if="totpEnabled"
              type="success"
              title="TOTP 已启用"
              style="margin-bottom: 16px"
            >
              您的账户已启用两步验证保护
            </NAlert>
            <NAlert v-else type="info" style="margin-bottom: 16px">
              TOTP 目前未启用，启用后可添加额外的安全保护
            </NAlert>

            <NButton
              v-if="!totpEnabled"
              type="primary"
              :loading="totpLoading"
              @click="handleEnableTotp"
            >
              启用 TOTP
            </NButton>
            <NButton v-else type="error" @click="handleDisableTotp"> 禁用 TOTP </NButton>
          </div>
        </NTabPane>

        <NTabPane name="passkeys" tab="通行密钥">
          <div class="section">
            <NAlert type="info" style="margin-bottom: 16px">
              通行密钥允许您使用 WebAuthn 安全地无密码登录
            </NAlert>

            <div v-if="passkeys.length > 0" class="list">
              <div v-for="passkey in passkeys" :key="passkey.id" class="list-item">
                <div class="item-info">
                  <strong>{{ passkey.name || '通行密钥' }}</strong>
                  <span class="item-meta">
                    创建时间: {{ new Date(passkey.createdAt).toLocaleDateString('zh-CN') }}
                    <span v-if="passkey.lastUsedAt">
                      | 最后使用: {{ new Date(passkey.lastUsedAt).toLocaleDateString('zh-CN') }}
                    </span>
                  </span>
                </div>
                <NPopconfirm @positive-click="handleDeletePasskey(passkey.id)">
                  <template #trigger>
                    <NButton size="small" type="error">删除</NButton>
                  </template>
                  确定要删除此通行密钥吗？
                </NPopconfirm>
              </div>
            </div>

            <NButton type="primary" style="margin-top: 16px" @click="showPasskeySetup = true">
              添加通行密钥
            </NButton>
          </div>
        </NTabPane>

        <NTabPane name="sessions" tab="会话管理">
          <div class="section">
            <NAlert type="info" style="margin-bottom: 16px">
              以下是您的所有活动会话列表，您可以单独撤销或一次性撤销所有会话
            </NAlert>

            <div v-if="sessions.length > 0" class="list">
              <div v-for="session in sessions" :key="session.id" class="list-item">
                <div class="item-info">
                  <div class="session-id">会话: {{ session.id.substring(0, 8) }}...</div>
                  <div class="item-meta">
                    创建: {{ new Date(session.createdAt).toLocaleString('zh-CN') }} | 过期:
                    {{ new Date(session.expiresAt).toLocaleString('zh-CN') }}
                  </div>
                </div>
                <NPopconfirm @positive-click="handleRevokeSession(session.id)">
                  <template #trigger>
                    <NButton size="small" type="error">撤销</NButton>
                  </template>
                  确定要撤销此会话吗？
                </NPopconfirm>
              </div>
            </div>
            <div v-else class="empty-state">
              <NAlert type="info">暂无活动会话</NAlert>
            </div>

            <NButton
              v-if="sessions.length > 0"
              type="error"
              :loading="sessionLoading"
              style="margin-top: 16px"
              @click="handleRevokeAllSessions"
            >
              撤销所有会话
            </NButton>
          </div>
        </NTabPane>
      </NTabs>
    </NCard>

    <NModal v-model:show="showTotpSetup" preset="card" title="设置身份验证器" style="width: 450px">
      <NAlert type="info" style="margin-bottom: 16px">
        请使用身份验证器应用扫描此二维码，然后输入 6 位数字验证码进行验证
      </NAlert>

      <div class="totp-setup">
        <div class="qr-code">
          <NImage v-if="totpUrl" :src="totpUrl" width="200" height="200" />
        </div>
        <div class="secret">
          <p>或者手动输入此密钥：</p>
          <code>{{ totpSecret }}</code>
        </div>
        <NFormItem label="验证码" style="margin-top: 16px">
          <NInput
            v-model:value="totpVerifyCode"
            placeholder="请输入 6 位验证码"
            maxlength="6"
            @keydown.enter="handleVerifyTotp"
          />
        </NFormItem>
        <NSpace justify="end">
          <NButton @click="showTotpSetup = false">取消</NButton>
          <NButton type="primary" :loading="totpVerifyLoading" @click="handleVerifyTotp"
            >验证</NButton
          >
        </NSpace>
      </div>
    </NModal>

    <NModal v-model:show="showPasskeySetup" preset="card" title="添加通行密钥" style="width: 450px">
      <NForm :label-width="100">
        <NFormItem label="通行密钥名称">
          <NInput v-model:value="passkeyName" placeholder="例如：我的笔记本电脑" />
        </NFormItem>
      </NForm>
      <NSpace justify="end">
        <NButton @click="showPasskeySetup = false">取消</NButton>
        <NButton type="primary" :loading="registeringPasskey" @click="handleRegisterPasskey">
          创建通行密钥
        </NButton>
      </NSpace>
    </NModal>

    <NModal
      v-model:show="showDisableTotpModal"
      preset="card"
      title="禁用 TOTP"
      style="width: 400px"
    >
      <NAlert type="warning" style="margin-bottom: 16px"> 请输入您的 TOTP 验证码以确认禁用 </NAlert>
      <NFormItem label="验证码">
        <NInput
          v-model:value="disableTotpCode"
          placeholder="请输入 6 位验证码"
          maxlength="6"
          @keydown.enter="confirmDisableTotp"
        />
      </NFormItem>
      <NSpace justify="end">
        <NButton @click="showDisableTotpModal = false">取消</NButton>
        <NButton type="error" @click="confirmDisableTotp">确认禁用</NButton>
      </NSpace>
    </NModal>
  </div>
</template>

<style scoped>
.page-container {
  padding: 40px 0;
  max-width: 800px;
  margin: 0 auto;
}

.section {
  padding: 20px 0;
}

.password-strength {
  margin-bottom: 16px;
  max-width: 400px;
}

.strength-label {
  font-size: 13px;
  display: block;
  margin-top: 6px;
}

.list {
  margin-bottom: 16px;
}

.list-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: var(--social-bg);
  border-radius: 8px;
  margin-bottom: 8px;
}

.item-info {
  display: flex;
  flex-direction: column;
}

.item-meta {
  font-size: 13px;
  color: var(--text);
  margin-top: 4px;
}

.session-id {
  font-family: var(--mono);
  font-size: 13px;
  color: var(--text-h);
}

.totp-setup {
  text-align: center;
}

.qr-code {
  margin-bottom: 16px;
}

.secret {
  margin-bottom: 16px;
  text-align: left;
}

.secret p {
  font-size: 14px;
  color: var(--text);
  margin-bottom: 8px;
}

.secret code {
  display: block;
  padding: 12px;
  background: var(--social-bg);
  border-radius: 8px;
  font-family: var(--mono);
  word-break: break-all;
  font-size: 13px;
}

.empty-state {
  padding: 24px 0;
}
</style>
