<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NButton, NSpace, NResult, NSpin, useMessage } from 'naive-ui'
import { oidcApi } from '@/api/oidc'

const route = useRoute()
const router = useRouter()
const message = useMessage()

const uid = route.params.uid as string
const loading = ref(true)
const error = ref('')
const confirming = ref(false)
const declining = ref(false)

const clientName = ref('')
const clientLogo = ref('')
const scopes = ref<string[]>([])
const claims = ref<Record<string, any>>({})

function isSafeLogoUri(uri: string): boolean {
  if (!uri) return false
  try {
    const parsed = new URL(uri)
    return ['https:'].includes(parsed.protocol)
  } catch {
    return false
  }
}

onMounted(async () => {
  try {
    const response = await oidcApi.getInteraction(uid)
    if (response.data.success && response.data.data) {
      const data = response.data.data
      clientName.value = data.client?.name || '未知应用'
      const logoUri = data.client?.logoUri
      clientLogo.value = logoUri && isSafeLogoUri(logoUri) ? logoUri : ''
      scopes.value = data.scopes || []
      claims.value = data.claims || {}
    }
  } catch (err: any) {
    error.value = err.response?.data?.message || '无法加载授权信息'
  } finally {
    loading.value = false
  }
})

async function handleAuthorize() {
  confirming.value = true
  try {
    const response = await oidcApi.confirmInteraction(uid)
    if (response.data.success && response.data.data) {
      safeRedirect(response.data.data.redirectTo, '/profile')
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '授权失败')
    confirming.value = false
  }
}

async function handleDecline() {
  declining.value = true
  try {
    const response = await oidcApi.cancelInteraction(uid)
    if (response.data.success && response.data.data) {
      safeRedirect(response.data.data.redirectTo, '/profile')
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '取消失败')
    declining.value = false
  }
}

function getScopeDescription(scope: string): string {
  const descriptions: Record<string, string> = {
    openid: '验证您的身份',
    profile: '访问您的基本 profile 信息',
    email: '访问您的邮箱地址',
    groups: '访问您的用户组信息',
  }
  return descriptions[scope] || scope
}

function isValidRedirectUrl(url: string): boolean {
  try {
    const parsed = new URL(url)
    return parsed.origin === window.location.origin || url.startsWith('/')
  } catch {
    return url.startsWith('/')
  }
}

function safeRedirect(url: string, fallback: string = '/') {
  if (isValidRedirectUrl(url)) {
    window.location.href = url
  } else {
    message.error('无效的重定向地址，已返回首页')
    router.push(fallback)
  }
}
</script>

<template>
  <div class="auth-container">
    <NCard class="auth-card" :bordered="false">
      <template #header>
        <div class="page-header">
          <h1>授权请求</h1>
        </div>
      </template>

      <NSpin v-if="loading" />

      <NResult v-else-if="error" status="error" title="错误" :description="error">
        <template #footer>
          <NButton @click="router.push('/profile')">前往个人中心</NButton>
        </template>
      </NResult>

      <template v-else>
        <div class="client-info">
          <img v-if="clientLogo" :src="clientLogo" :alt="clientName" class="client-logo" />
          <div class="client-details">
            <h3>{{ clientName }}</h3>
            <p>请求访问您的账户</p>
          </div>
        </div>

        <div class="scopes">
          <h4>请求的权限：</h4>
          <ul class="scope-list">
            <li v-for="scope in scopes" :key="scope">
              <strong>{{ scope }}</strong>
              <span class="scope-desc">{{ getScopeDescription(scope) }}</span>
            </li>
          </ul>
        </div>

        <div v-if="Object.keys(claims).length > 0" class="claims">
          <h4>请求的声明：</h4>
          <ul class="claim-list">
            <li v-for="(value, key) in claims" :key="key">
              <strong>{{ key }}</strong
              >: {{ value }}
            </li>
          </ul>
        </div>

        <NSpace justify="end">
          <NButton :loading="declining" :disabled="confirming" @click="handleDecline">
            拒绝
          </NButton>
          <NButton
            type="primary"
            :loading="confirming"
            :disabled="declining"
            @click="handleAuthorize"
          >
            授权
          </NButton>
        </NSpace>
      </template>
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
  max-width: 480px;
}

.page-header {
  text-align: center;
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 24px;
  margin-bottom: 0;
  color: var(--text-h);
}

.client-info {
  display: flex;
  align-items: center;
  margin-bottom: 24px;
  padding: 20px;
  background: var(--social-bg);
  border-radius: 12px;
}

.client-logo {
  width: 56px;
  height: 56px;
  margin-right: 16px;
  object-fit: contain;
  border-radius: 8px;
}

.client-details h3 {
  margin: 0 0 4px 0;
  font-size: 18px;
}

.client-details p {
  margin: 0;
  color: var(--text);
  font-size: 14px;
}

.scopes {
  margin-bottom: 20px;
}

.scopes h4 {
  margin-bottom: 12px;
  font-size: 14px;
  color: var(--text);
}

.scope-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.scope-list li {
  padding: 12px 16px;
  margin-bottom: 8px;
  background: var(--social-bg);
  border-radius: 8px;
}

.scope-list li strong {
  display: block;
  color: var(--text-h);
  margin-bottom: 2px;
}

.scope-desc {
  display: block;
  font-size: 13px;
  color: var(--text);
}

.claims {
  margin-bottom: 20px;
}

.claims h4 {
  margin-bottom: 8px;
  font-size: 14px;
  color: var(--text);
}

.claim-list {
  padding-left: 20px;
  margin: 0;
}

.claim-list li {
  margin-bottom: 4px;
  font-size: 14px;
}
</style>
