<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import {
  NCard,
  NDataTable,
  NButton,
  NAlert,
  NModal,
  NForm,
  NFormItem,
  NInput,
  useMessage,
  NPopconfirm,
} from 'naive-ui'
import type { Passkey } from '@/types'
import { passkeyApi } from '@/api/auth'

const message = useMessage()

const passkeys = ref<Passkey[]>([])
const loading = ref(false)
const showPasskeySetup = ref(false)
const passkeyName = ref('')
const registeringPasskey = ref(false)

async function loadPasskeys() {
  loading.value = true
  try {
    const response = await passkeyApi.getPasskeys()
    if (response.data.success && response.data.data) {
      passkeys.value = response.data.data
    }
  } catch (err: any) {
    message.error(err.response?.data?.message || '加载通行密钥失败')
  } finally {
    loading.value = false
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

const columns = [
  { title: '名称', key: 'name', render: (row: Passkey) => row.name || '通行密钥' },
  { title: '凭证 ID', key: 'credentialId', ellipsis: true },
  {
    title: '最后使用',
    key: 'lastUsedAt',
    render: (row: Passkey) =>
      row.lastUsedAt ? new Date(row.lastUsedAt).toLocaleDateString('zh-CN') : '从未使用',
  },
  {
    title: '创建时间',
    key: 'createdAt',
    render: (row: Passkey) => new Date(row.createdAt).toLocaleDateString('zh-CN'),
  },
  {
    title: '操作',
    key: 'actions',
    render: (row: Passkey) =>
      h(
        NPopconfirm,
        { onPositiveClick: () => handleDeletePasskey(row.id) },
        {
          trigger: () => h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' }),
          default: () => '确定要删除此通行密钥吗？',
        }
      ),
  },
]

onMounted(() => {
  loadPasskeys()
})
</script>

<template>
  <div class="page-container">
    <NCard title="通行密钥" :bordered="false">
      <template #header-extra>
        <NButton type="primary" @click="showPasskeySetup = true">添加通行密钥</NButton>
      </template>

      <NAlert type="info" style="margin-bottom: 16px">
        通行密钥允许您使用 WebAuthn 安全地无密码登录
      </NAlert>

      <NDataTable :columns="columns" :data="passkeys" :loading="loading" :bordered="false" />
    </NCard>

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
  </div>
</template>

<style scoped>
.page-container {
  padding: 40px 0;
  max-width: 900px;
  margin: 0 auto;
}
</style>
