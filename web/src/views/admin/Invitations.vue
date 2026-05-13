<script setup lang="ts">
import { ref, h, onMounted, computed } from 'vue'
import { 
  NCard, NDataTable, NButton, NSpace, NTag, NModal, NForm, NFormItem, 
  NInput, NSelect, useMessage, NPopconfirm, NDatePicker
} from 'naive-ui'
import type { Group } from '@/types'
import { adminApi, type CreateInvitationRequest } from '@/api/admin'

interface InvitationListItem {
  id: string
  email: string
  username?: string
  expiresAt: string
  createdAt: string
}

const message = useMessage()

const invitations = ref<InvitationListItem[]>([])
const groups = ref<Group[]>([])
const loading = ref(false)

const showInviteModal = ref(false)
const inviteForm = ref({
  email: '',
  username: '',
  scopes: '',
  groupId: undefined as string | undefined,
  maxUses: '1',
  expiresAt: null as number | null,
})

function getInviteRequest(): CreateInvitationRequest {
  return {
    email: inviteForm.value.email,
    username: inviteForm.value.username || undefined,
    scopes: inviteForm.value.scopes || undefined,
    groupId: inviteForm.value.groupId || undefined,
    maxUses: inviteForm.value.maxUses ? parseInt(inviteForm.value.maxUses, 10) : undefined,
    expiresAt: inviteForm.value.expiresAt ? new Date(inviteForm.value.expiresAt).toISOString() : undefined,
  }
}

const groupOptions = computed(() => 
  groups.value.map(g => ({ label: g.name, value: g.id }))
)

const columns = [
  { title: '邮箱', key: 'email' },
  { title: '用户名', key: 'username' },
  { 
    title: '状态', 
    key: 'expiresAt', 
    render: (row: InvitationListItem) => {
      const expired = new Date(row.expiresAt) < new Date()
      return h(NTag, { type: expired ? 'error' : 'success' }, { default: () => expired ? '已过期' : '有效' })
    }
  },
  { title: '过期时间', key: 'expiresAt', render: (row: InvitationListItem) => new Date(row.expiresAt).toLocaleDateString('zh-CN') },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    render: (row: InvitationListItem) => h(NSpace, { size: 'small' }, {
      default: () => [
        h(NPopconfirm, { onConfirm: () => handleDelete(row.id) }, {
          trigger: () => h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' }),
          default: () => '确定要删除此邀请吗？'
        })
      ]
    })
  }
]

async function fetchInvitations() {
  loading.value = true
  try {
    const res = await adminApi.invitations.list()
    if (res.data.success && res.data.data) {
      invitations.value = res.data.data
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '获取邀请列表失败')
  } finally {
    loading.value = false
  }
}

async function fetchGroups() {
  try {
    const res = await adminApi.groups.list()
    if (res.data.success && res.data.data) {
      groups.value = res.data.data
    }
  } catch (err: any) {
    console.error('获取用户组列表失败:', err)
  }
}

function openCreateModal() {
  inviteForm.value = {
    email: '',
    username: '',
    scopes: '',
    groupId: undefined,
    maxUses: '1',
    expiresAt: null,
  }
  showInviteModal.value = true
}

async function handleCreateInvite() {
  try {
    await adminApi.invitations.create(getInviteRequest())
    message.success('邀请创建成功')
    showInviteModal.value = false
    fetchInvitations()
  } catch (err: any) {
    message.error(err.response?.data?.message || '创建邀请失败')
  }
}

async function handleDelete(id: string) {
  try {
    await adminApi.invitations.delete(id)
    message.success('邀请删除成功')
    fetchInvitations()
  } catch (err: any) {
    message.error(err.response?.data?.message || '删除邀请失败')
  }
}

onMounted(() => {
  fetchInvitations()
  fetchGroups()
})
</script>

<template>
  <div class="page-container">
    <NCard title="邀请管理">
      <template #header-extra>
        <NButton type="primary" @click="openCreateModal">创建邀请</NButton>
      </template>

      <NDataTable 
        :columns="columns" 
        :data="invitations" 
        :loading="loading"
        :bordered="false" 
      />
    </NCard>

    <NModal v-model:show="showInviteModal" preset="card" title="创建邀请" style="width: 500px">
      <NForm :model="inviteForm" label-placement="top">
        <NFormItem label="邮箱" required>
          <NInput v-model:value="inviteForm.email" placeholder="user@example.com" />
        </NFormItem>
        <NFormItem label="用户名（可选）">
          <NInput v-model:value="inviteForm.username" placeholder="期望的用户名" />
        </NFormItem>
        <NFormItem label="用户组">
          <NSelect 
            v-model:value="inviteForm.groupId" 
            :options="groupOptions" 
            placeholder="选择用户组（可选）"
            clearable
          />
        </NFormItem>
        <NFormItem label="作用域">
          <NInput v-model:value="inviteForm.scopes" placeholder="openid profile email" />
        </NFormItem>
        <NFormItem label="最大使用次数">
          <NInput v-model:value="inviteForm.maxUses" />
        </NFormItem>
        <NFormItem label="过期时间" required>
          <NDatePicker 
            v-model:value="inviteForm.expiresAt"
            type="datetime" 
            style="width: 100%"
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showInviteModal = false">取消</NButton>
          <NButton type="primary" @click="handleCreateInvite">创建</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.page-container {
  padding: 40px 0;
}
</style>
