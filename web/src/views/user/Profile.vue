<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  NCard,
  NDescriptions,
  NDescriptionsItem,
  NButton,
  NSpace,
  useMessage,
  NForm,
  NFormItem,
  NInput,
} from 'naive-ui'
import { userApi, authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/types'

const message = useMessage()
const authStore = useAuthStore()
const user = ref<User | null>(null)
const loading = ref(true)
const editing = ref(false)
const saving = ref(false)

const editForm = ref({
  name: '',
  email: '',
})

onMounted(async () => {
  await loadProfile()
})

async function loadProfile() {
  try {
    const response = await userApi.getMe()
    if (response.data.success && response.data.data) {
      user.value = response.data.data
      authStore.setUser(response.data.data)
      editForm.value.name = response.data.data.name || ''
      editForm.value.email = response.data.data.email || ''
    }
  } catch (_err) {
    message.error('加载用户信息失败')
  } finally {
    loading.value = false
  }
}

function startEdit() {
  if (user.value) {
    editForm.value.name = user.value.name || ''
    editForm.value.email = user.value.email || ''
    editing.value = true
  }
}

async function saveEdit() {
  if (!editForm.value.email) {
    message.error('邮箱不能为空')
    return
  }

  saving.value = true
  try {
    const response = await userApi.updateMe({
      name: editForm.value.name || undefined,
      email: editForm.value.email,
    })
    if (response.data.success && response.data.data) {
      user.value = response.data.data
      authStore.setUser(response.data.data)
      message.success('个人信息更新成功')
      editing.value = false
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '更新个人信息失败')
  } finally {
    saving.value = false
  }
}

function cancelEdit() {
  editing.value = false
}

async function handleResendVerification() {
  try {
    if (user.value?.email) {
      await authApi.sendVerifyEmail({ email: user.value.email })
      message.success('验证邮件已发送')
    } else {
      message.error('无法获取邮箱地址')
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '发送验证邮件失败')
  }
}
</script>

<template>
  <div class="page-container">
    <NCard title="个人资料" :bordered="false">
      <NDescriptions v-if="user && !editing" :column="1" bordered>
        <NDescriptionsItem label="用户名">{{ user.username }}</NDescriptionsItem>
        <NDescriptionsItem label="邮箱">{{ user.email || '未设置' }}</NDescriptionsItem>
        <NDescriptionsItem label="姓名">{{ user.name || '未设置' }}</NDescriptionsItem>
        <NDescriptionsItem label="邮箱已验证">
          <NButton v-if="!user.emailVerified" size="tiny" @click="handleResendVerification">
            立即验证
          </NButton>
          <span v-else class="badge badge-success">已验证</span>
        </NDescriptionsItem>
        <NDescriptionsItem label="已批准">{{ user.approved ? '是' : '否' }}</NDescriptionsItem>
        <NDescriptionsItem label="创建时间">{{
          new Date(user.createdAt).toLocaleDateString('zh-CN')
        }}</NDescriptionsItem>
      </NDescriptions>

      <NForm v-else-if="editing" :label-width="100" style="max-width: 400px">
        <NFormItem label="用户名">
          <NInput :value="user?.username" disabled />
        </NFormItem>
        <NFormItem label="邮箱" required>
          <NInput v-model:value="editForm.email" placeholder="请输入邮箱" />
        </NFormItem>
        <NFormItem label="姓名">
          <NInput v-model:value="editForm.name" placeholder="请输入姓名" />
        </NFormItem>
        <NSpace justify="end">
          <NButton @click="cancelEdit">取消</NButton>
          <NButton type="primary" :loading="saving" @click="saveEdit">保存</NButton>
        </NSpace>
      </NForm>

      <div v-if="!editing" class="page-actions">
        <NButton type="primary" @click="startEdit">编辑资料</NButton>
      </div>
    </NCard>
  </div>
</template>

<style scoped>
.page-container {
  padding: 32px 0;
}

.page-actions {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border);
}
</style>
