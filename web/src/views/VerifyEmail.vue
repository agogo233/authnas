<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NResult, NButton, NSpace, NAlert, useMessage } from 'naive-ui'
import { authApi } from '@/api/auth'

const route = useRoute()
const router = useRouter()
const message = useMessage()

const status = ref<'success' | 'error'>('success')
const loading = ref(true)
const resendLoading = ref(false)
const resendSuccess = ref(false)

onMounted(async () => {
  const userId = route.query.userId as string
  const challenge = route.query.challenge as string
  if (!userId || !challenge) {
    status.value = 'error'
    loading.value = false
    return
  }

  try {
    await authApi.verifyEmail({ userId, challenge })
    status.value = 'success'
    message.success('邮箱验证成功')
  } catch (_err) {
    status.value = 'error'
  } finally {
    loading.value = false
  }
})

async function handleResend() {
  resendLoading.value = true
  try {
    const email = route.query.email as string
    if (email) {
      await authApi.sendVerifyEmail({ email })
    } else {
      message.error('无法获取邮箱地址')
      return
    }
    resendSuccess.value = true
    message.success('验证邮件已发送')
  } catch (err: any) {
    message.error(err.response?.data?.message || '发送验证邮件失败')
  } finally {
    resendLoading.value = false
  }
}
</script>

<template>
  <div class="auth-container">
    <NCard class="auth-card" :bordered="false">
      <NAlert v-if="resendSuccess" type="success" style="margin-bottom: 16px">
        验证邮件已发送，请查收您的收件箱
      </NAlert>

      <NResult
        v-if="!loading"
        :status="status === 'success' ? 'success' : 'error'"
        :title="status === 'success' ? '邮箱已验证' : '验证失败'"
        :description="status === 'success' ? '您的邮箱已成功验证。' : '验证链接无效或已过期'"
      >
        <template #footer>
          <NSpace justify="center">
            <NButton v-if="status === 'error'" type="primary" :loading="resendLoading" @click="handleResend">
              重新发送验证邮件
            </NButton>
            <NButton @click="router.push('/login')">
              前往登录
            </NButton>
          </NSpace>
        </template>
      </NResult>
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
</style>
