<script setup lang="ts">
import { ref } from 'vue'
import { NCard, NForm, NFormItem, NInput, NButton, NAlert, useMessage } from 'naive-ui'
import { authApi } from '@/api/auth'

const message = useMessage()
const email = ref('')
const loading = ref(false)
const error = ref('')
const success = ref(false)

async function handleSubmit() {
  if (!email.value) {
    error.value = '请输入您的邮箱'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const res = await authApi.forgotPassword({ email: email.value })
    if (res.data.success) {
      success.value = true
      message.success('密码重置邮件已发送')
    } else {
      error.value = res.data.message || '发送密码重置邮件失败'
    }
  } catch (err: any) {
    error.value = err.response?.data?.message || '发送密码重置邮件失败'
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
          <h1>重置密码</h1>
          <p>请输入您的注册邮箱</p>
        </div>
      </template>

      <NAlert v-if="error" type="error" style="margin-bottom: 20px">
        {{ error }}
      </NAlert>

      <NAlert v-if="success" type="success" style="margin-bottom: 20px">
        密码重置邮件已发送，请查收您的收件箱
      </NAlert>

      <NForm @submit.prevent="handleSubmit">
        <NFormItem label="邮箱" path="email">
          <NInput v-model:value="email" placeholder="请输入您的邮箱" size="large" />
        </NFormItem>

        <NButton type="primary" attr-type="submit" :loading="loading" block size="large">
          发送重置邮件
        </NButton>
      </NForm>

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
