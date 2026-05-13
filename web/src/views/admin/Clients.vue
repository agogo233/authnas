<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { 
  NCard, NDataTable, NButton, NSpace, NModal, NForm, NFormItem, 
  NInput, useMessage, NPopconfirm, NAlert, NInputGroup
} from 'naive-ui'
import { adminApi, type CreateClientRequest, type UpdateClientRequest } from '@/api/admin'

const message = useMessage()

interface ClientListItem {
  id: string
  clientId: string
  name: string
  logoUri?: string
  createdAt: string
}

interface Client extends ClientListItem {
  clientSecret?: string
  redirectUris: string
  postLogoutRedirectUris?: string
  grantTypes?: string
  responseTypes?: string
  scopes?: string
}

const clients = ref<ClientListItem[]>([])
const loading = ref(false)

const showClientModal = ref(false)
const editingClient = ref<Client | null>(null)
const clientForm = ref<CreateClientRequest>({
  clientId: '',
  name: '',
  logoUri: '',
  redirectUris: '',
  postLogoutRedirectUris: '',
  grantTypes: 'authorization_code',
  responseTypes: 'code',
  scopes: 'openid profile email',
})

const showSecretModal = ref(false)
const newClientSecret = ref('')

const columns = [
  { title: '名称', key: 'name' },
  { title: '客户端 ID', key: 'clientId', ellipsis: true },
  { title: '创建时间', key: 'createdAt', render: (row: ClientListItem) => new Date(row.createdAt).toLocaleDateString('zh-CN') },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render: (row: ClientListItem) => h(NSpace, { size: 'small' }, {
      default: () => [
        h(NButton, { size: 'small', onClick: () => openEditModal(row as Client) }, { default: () => '编辑' }),
        h(NPopconfirm, { onConfirm: () => handleDelete(row.id) }, {
          trigger: () => h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' }),
          default: () => '确定要删除此客户端吗？'
        })
      ]
    })
  }
]

async function fetchClients() {
  loading.value = true
  try {
    const res = await adminApi.clients.list()
    if (res.data.success && res.data.data) {
      clients.value = res.data.data
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '获取客户端列表失败')
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  editingClient.value = null
  clientForm.value = {
    clientId: '',
    name: '',
    logoUri: '',
    redirectUris: '',
    postLogoutRedirectUris: '',
    grantTypes: 'authorization_code',
    responseTypes: 'code',
    scopes: 'openid profile email',
  }
  showClientModal.value = true
}

function openEditModal(client: Client) {
  editingClient.value = client
  clientForm.value = {
    clientId: client.clientId,
    name: client.name,
    logoUri: client.logoUri || '',
    redirectUris: client.redirectUris || '',
    postLogoutRedirectUris: client.postLogoutRedirectUris || '',
    grantTypes: client.grantTypes || 'authorization_code',
    responseTypes: client.responseTypes || 'code',
    scopes: client.scopes || 'openid profile email',
  }
  showClientModal.value = true
}

async function handleSaveClient() {
  try {
    if (editingClient.value) {
      await adminApi.clients.update(editingClient.value.id, clientForm.value as UpdateClientRequest)
      message.success('客户端更新成功')
    } else {
      const res = await adminApi.clients.create(clientForm.value) as any
      if (res.data.data?.clientSecret) {
        newClientSecret.value = res.data.data.clientSecret
        showSecretModal.value = true
      }
      message.success('客户端创建成功')
    }
    showClientModal.value = false
    fetchClients()
  } catch (err: any) {
    message.error(err.response?.data?.message || '保存客户端失败')
  }
}

async function handleDelete(id: string) {
  try {
    await adminApi.clients.delete(id)
    message.success('客户端删除成功')
    fetchClients()
  } catch (err: any) {
    message.error(err.response?.data?.message || '删除客户端失败')
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  message.success('已复制到剪贴板')
}

onMounted(() => {
  fetchClients()
})
</script>

<template>
  <div class="page-container">
    <NCard title="OIDC 客户端管理">
      <template #header-extra>
        <NButton type="primary" @click="openCreateModal">创建客户端</NButton>
      </template>

      <NDataTable 
        :columns="columns" 
        :data="clients" 
        :loading="loading"
        :bordered="false" 
      />
    </NCard>

    <NModal v-model:show="showClientModal" preset="card" :title="editingClient ? '编辑客户端' : '创建客户端'" style="width: 600px">
      <NForm :model="clientForm" label-placement="top">
        <NFormItem label="客户端 ID" required>
          <NInput v-model:value="clientForm.clientId" placeholder="my-app" />
        </NFormItem>
        <NFormItem label="客户端名称" required>
          <NInput v-model:value="clientForm.name" placeholder="我的应用" />
        </NFormItem>
        <NFormItem label="Logo URI">
          <NInput v-model:value="clientForm.logoUri" placeholder="https://example.com/logo.png" />
        </NFormItem>
        <NFormItem label="回调 URI" required>
          <NInput 
            v-model:value="clientForm.redirectUris" 
            placeholder="https://example.com/callback" 
            type="textarea"
            :rows="2"
          />
        </NFormItem>
        <NFormItem label="退出后回调 URI">
          <NInput 
            v-model:value="clientForm.postLogoutRedirectUris" 
            placeholder="https://example.com" 
            type="textarea"
            :rows="2"
          />
        </NFormItem>
        <NFormItem label="授权类型">
          <NInput v-model:value="clientForm.grantTypes" placeholder="authorization_code" />
        </NFormItem>
        <NFormItem label="响应类型">
          <NInput v-model:value="clientForm.responseTypes" placeholder="code" />
        </NFormItem>
        <NFormItem label="作用域">
          <NInput v-model:value="clientForm.scopes" placeholder="openid profile email" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showClientModal = false">取消</NButton>
          <NButton type="primary" @click="handleSaveClient">保存</NButton>
        </NSpace>
      </template>
    </NModal>

    <NModal v-model:show="showSecretModal" preset="card" title="客户端凭证" style="width: 500px">
      <NSpace vertical :size="16">
        <NAlert type="warning">
          请立即保存客户端密钥，之后将无法再次查看
        </NAlert>
        <NFormItem label="客户端 ID">
          <NInput :value="editingClient?.clientId || ''" readonly />
        </NFormItem>
        <NFormItem label="客户端密钥">
          <NInputGroup>
            <NInput :value="newClientSecret" readonly />
            <NButton type="primary" @click="copyToClipboard(newClientSecret)">复制</NButton>
          </NInputGroup>
        </NFormItem>
      </NSpace>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showSecretModal = false">关闭</NButton>
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
