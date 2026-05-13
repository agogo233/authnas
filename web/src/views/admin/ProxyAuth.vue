<script setup lang="ts">
import { ref, h, onMounted, computed } from 'vue'
import { 
  NCard, NDataTable, NButton, NSpace, NTag, NModal, NForm, NFormItem, 
  NInput, NSelect, NSwitch, useMessage, NPopconfirm
} from 'naive-ui'
import type { Group } from '@/types'
import { adminApi, type CreateProxyAuthRequest, type UpdateProxyAuthRequest } from '@/api/admin'

interface ProxyAuthListItem {
  id: string
  name: string
  proxyUrl: string
  enabled: boolean
  createdAt: string
}

interface ProxyAuth extends ProxyAuthListItem {
  headerName: string
  groupId?: string
  scopes?: string
}

const message = useMessage()

const proxyauths = ref<ProxyAuthListItem[]>([])
const groups = ref<Group[]>([])
const loading = ref(false)

const showProxyAuthModal = ref(false)
const editingProxyAuth = ref<ProxyAuth | null>(null)
const proxyAuthForm = ref<CreateProxyAuthRequest>({
  name: '',
  proxyUrl: '',
  headerName: '',
  scopes: '',
  groupId: undefined,
  enabled: true,
})

const groupOptions = computed(() => 
  groups.value.map(g => ({ label: g.name, value: g.id }))
)

const columns = [
  { title: '名称', key: 'name' },
  { title: '代理 URL', key: 'proxyUrl', ellipsis: true },
  { 
    title: '状态', 
    key: 'enabled', 
    render: (row: ProxyAuthListItem) => h(NTag, { type: row.enabled ? 'success' : 'default' }, { default: () => row.enabled ? '已启用' : '已禁用' }) 
  },
  { title: '创建时间', key: 'createdAt', render: (row: ProxyAuthListItem) => new Date(row.createdAt).toLocaleDateString('zh-CN') },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render: (row: ProxyAuthListItem) => h(NSpace, { size: 'small' }, {
      default: () => [
        h(NButton, { size: 'small', onClick: () => openEditModal(row as ProxyAuth) }, { default: () => '编辑' }),
        h(NPopconfirm, { onConfirm: () => handleDelete(row.id) }, {
          trigger: () => h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' }),
          default: () => '确定要删除此代理认证吗？'
        })
      ]
    })
  }
]

async function fetchProxyAuths() {
  loading.value = true
  try {
    const res = await adminApi.proxyauth.list()
    if (res.data.success && res.data.data) {
      proxyauths.value = res.data.data
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '获取代理认证配置失败')
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
  editingProxyAuth.value = null
  proxyAuthForm.value = {
    name: '',
    proxyUrl: '',
    headerName: '',
    scopes: '',
    groupId: undefined,
    enabled: true,
  }
  showProxyAuthModal.value = true
}

function openEditModal(proxyauth: ProxyAuth) {
  editingProxyAuth.value = proxyauth
  proxyAuthForm.value = {
    name: proxyauth.name,
    proxyUrl: proxyauth.proxyUrl,
    headerName: proxyauth.headerName,
    scopes: proxyauth.scopes || '',
    groupId: proxyauth.groupId,
    enabled: proxyauth.enabled,
  }
  showProxyAuthModal.value = true
}

async function handleSaveProxyAuth() {
  try {
    if (editingProxyAuth.value) {
      await adminApi.proxyauth.update(editingProxyAuth.value.id, proxyAuthForm.value as UpdateProxyAuthRequest)
      message.success('代理认证更新成功')
    } else {
      await adminApi.proxyauth.create(proxyAuthForm.value)
      message.success('代理认证创建成功')
    }
    showProxyAuthModal.value = false
    fetchProxyAuths()
  } catch (err: any) {
    message.error(err.response?.data?.message || '保存代理认证失败')
  }
}

async function handleDelete(id: string) {
  try {
    await adminApi.proxyauth.delete(id)
    message.success('代理认证删除成功')
    fetchProxyAuths()
  } catch (err: any) {
    message.error(err.response?.data?.message || '删除代理认证失败')
  }
}

onMounted(() => {
  fetchProxyAuths()
  fetchGroups()
})
</script>

<template>
  <div class="page-container">
    <NCard title="代理认证管理">
      <template #header-extra>
        <NButton type="primary" @click="openCreateModal">创建代理认证</NButton>
      </template>

      <NDataTable 
        :columns="columns" 
        :data="proxyauths" 
        :loading="loading"
        :bordered="false" 
      />
    </NCard>

    <NModal v-model:show="showProxyAuthModal" preset="card" :title="editingProxyAuth ? '编辑代理认证' : '创建代理认证'" style="width: 500px">
      <NForm :model="proxyAuthForm" label-placement="top">
        <NFormItem label="名称" required>
          <NInput v-model:value="proxyAuthForm.name" placeholder="我的代理" />
        </NFormItem>
        <NFormItem label="代理 URL" required>
          <NInput v-model:value="proxyAuthForm.proxyUrl" placeholder="https://proxy.example.com/auth" />
        </NFormItem>
        <NFormItem label="请求头名称" required>
          <NInput v-model:value="proxyAuthForm.headerName" placeholder="X-User-Token" />
        </NFormItem>
        <NFormItem label="用户组">
          <NSelect 
            v-model:value="proxyAuthForm.groupId" 
            :options="groupOptions" 
            placeholder="选择用户组（可选）"
            clearable
          />
        </NFormItem>
        <NFormItem label="作用域">
          <NInput v-model:value="proxyAuthForm.scopes" placeholder="openid profile email" />
        </NFormItem>
        <NFormItem label="启用">
          <NSwitch v-model:value="proxyAuthForm.enabled" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showProxyAuthModal = false">取消</NButton>
          <NButton type="primary" @click="handleSaveProxyAuth">保存</NButton>
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
